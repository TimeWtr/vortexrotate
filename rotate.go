// Copyright 2025 TimeWtr
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vortexrotate

import (
	"compress/gzip"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/TimeWtr/vortexrotate/errorx"
	"github.com/valyala/gozstd"
)

var (
	r             *Rotator
	onceWithError OnceWithError
)

type Option func(*Rotator) error

// WithCompress 开启压缩，压缩算法提供gzip、zstd和snappy三种算法，
// 当压缩算法为gzip时，可以设置压缩等级/级别，如果不设置，默认压缩级别
// 为gzip.DefaultCompression
func WithCompress(tp int, level ...int) Option {
	return func(r *Rotator) error {
		r.cpr.compress = true
		if tp < _minCompressType || tp > _maxCompressType {
			return errorx.ErrCompressType
		}

		r.cpr.compressType = tp
		var compressLevel int
		if len(level) > 0 {
			compressLevel = level[0]
		} else if tp == CompressTypeGzip {
			compressLevel = gzip.DefaultCompression
		}

		switch tp {
		case CompressTypeGzip:
			cs, err := NewGzip(nil, r.f, compressLevel)
			if err != nil {
				return err
			}
			r.cpr.cs = cs
		case CompressTypeZstd:
			r.cpr.cs = NewZstd(nil, r.f, gozstd.DefaultCompressionLevel)
		case CompressTypeSnappy:
			r.cpr.cs = NewSnappy(nil, r.f)
		default:
		}

		return nil
	}
}

// WithPeriod 设置保存周期
func WithPeriod(period uint16) Option {
	return func(r *Rotator) error {
		r.cleanup.period = period
		return nil
	}
}

// WithMaxCount 设置保存的最大文件数量
func WithMaxCount(_ uint16) Option {
	return func(_ *Rotator) error {
		// TODO 处理CleanUp初始化
		return nil
	}
}

// WithRotate 设置轮转配置，maxSize设置单个文件写入的最大字节，默认为100MB，当超过
// 限制后强制立即执行轮转，后台定时轮转的时间类型:
// Hour：一小时定时执行一次轮转
// Day：一天定时执行一次轮转
// Week：一周定时执行一次轮转
// Month: 一个月定时执行一次轮转
// 默认定时执行的时间类型是Hour。
func WithRotate(maxSize uint64, tp TimingType) Option {
	return func(r *Rotator) error {
		stg, err := NewMixStrategy(maxSize, tp)
		if err != nil {
			return err
		}
		r.stg = stg
		r.maxSize = maxSize

		return nil
	}
}

// Rotator 轮转器入口，执行真正的轮转和写入操作
// 根据轮转策略确定是否执行轮转，轮转策略包括：根据文件大小、定时以及混合策略，
// 如果需要轮转，根据新的文件名称执行轮转操作。文件轮转后根据压缩策略确定是否执行压缩操作，
// 以及根据清除策略确定是否执行清除，过期文件的操作。
type Rotator struct {
	// 文件存储目录
	dir string
	// 基础的文件名称，后续轮转的名称在该基础名称上生成新的名称
	filename string
	// 文件后缀名
	ext string
	// 当前写入的文件句柄
	f *os.File
	// 执行轮转的策略
	stg RotateStrategy
	// 文件写入锁保护
	writeLock sync.RWMutex
	// 压缩配置
	cpr Compress
	// 清理过期文件的配置
	cleanup *CleanUp
	// 关闭信号
	sig atomic.Int32
	// 轮转计数器
	counter atomic.Uint32
	// 日志
	l *log.Logger
	// 文件的最大写入字节
	maxSize uint64
}

// NewRotator 生产环境单例模式
func NewRotator(dir, filename string, opts ...Option) (*Rotator, error) {
	onceWithError.Do(func() error {
		rotator, err := newRotator(dir, filename, opts...)
		if err != nil {
			return err
		}

		r = rotator
		return nil
	})

	return r, onceWithError.err
}

func newRotator(dir, filename string, opts ...Option) (*Rotator, error) {
	const fileNameSliLength = 2
	sli := strings.Split(filename, ".")
	if len(sli) != fileNameSliLength {
		return nil, errorx.ErrFilename
	}

	rotator := &Rotator{
		dir:       dir,
		filename:  sli[0],
		ext:       sli[1],
		writeLock: sync.RWMutex{},
		l:         log.New(os.Stdout, "", log.LstdFlags),
		maxSize:   DefaultMaxSize,
	}

	rotator.sig.Store(0)
	rotator.counter.Store(1)

	if err := rotator.mkdirAll(); err != nil {
		return nil, err
	}

	f, err := os.OpenFile(rotator.newFile(), os.O_CREATE|os.O_RDWR|os.O_APPEND, ReadWriteFile)
	if err != nil {
		return nil, err
	}
	rotator.f = f

	for _, opt := range opts {
		if err := opt(rotator); err != nil {
			return nil, err
		}
	}

	if IsNil(rotator.stg) {
		rotator.stg, err = NewMixStrategy(DefaultMaxSize, Hour)
		if err != nil {
			return nil, err
		}
	}

	if rotator.cpr.compress && rotator.cpr.cs == nil {
		return nil, errorx.ErrCompress
	}

	go rotator.asyncWork()

	return rotator, nil
}

// Write 执行写入逻辑，判断大小是否已经达到最大大小，如果是则执行轮转逻辑
// 轮转后根据压缩配置执行压缩逻辑
func (r *Rotator) Write(p []byte) (int, error) {
	if r.sig.Load() == 1 {
		return 0, errorx.ErrRotateClosed
	}

	r.writeLock.Lock()
	defer r.writeLock.Unlock()
	if r.f == nil {
		return 0, os.ErrClosed
	}

	if r.stg.ShouldRotate(uint64(len(p))) {
		// 需要执行日志轮转
		if err := r.rotate(); err != nil {
			return 0, err
		}
	}

	n, err := r.f.Write(p)
	if err != nil {
		return n, err
	}

	return n, nil
}

func (r *Rotator) rotate() (err error) {
	_ = r.f.Close()
	if r.cpr.compress {
		r.l.Printf("rotate old file %s", r.f.Name())
		if err = r.cps(r.f.Name()); err != nil {
			fmt.Println("failed to cpr, cause: ", err.Error())
			return err
		}
	}

	f, err := os.OpenFile(r.newFile(), os.O_CREATE|os.O_RDWR, ReadWriteFile)
	if err != nil {
		return err
	}

	r.f = f

	return nil
}

// cps 执行压缩操作
func (r *Rotator) cps(oldPath string) error {
	wf := compressFn(oldPath, r.cpr.compressType)
	w, err := os.OpenFile(wf, os.O_RDWR|os.O_CREATE, ReadWriteFile)
	if err != nil {
		return err
	}

	f, err := os.Open(oldPath)
	if err != nil {
		return err
	}

	r.cpr.cs.Reset(w, f)
	if err := r.cpr.cs.Compress(); err != nil {
		return err
	}

	return nil
}

func (r *Rotator) Close() {
	r.writeLock.Lock()
	defer r.writeLock.Unlock()

	r.sig.Store(1)
	if r.f == nil {
		return
	}

	_ = r.f.Close()
}

// newFile 新的文件名称，组合日期(年月日)和当天的文件计数器来生成唯一的文件名称
func (r *Rotator) newFile() string {
	t := time.Now().Format(Layout)
	const template = "%s/%s/%s_%s_%04d.log"
	newFile := fmt.Sprintf(template, r.dir, t, r.filename, t, r.counter.Load())
	r.counter.Add(1)
	return newFile
}

// mkdirAll 创建文件目录，文件父目录是年月日时间，初始化时候会创建当天的目录，定时任务每天凌晨00:00会创建第二天的目录
func (r *Rotator) mkdirAll() error {
	t := time.Now().Format(Layout)
	return os.MkdirAll(fmt.Sprintf("%s/%s", r.dir, t), os.ModePerm)
}

// asyncWork 异步任务，用于接收定时轮转的信号，接收到之后立即执行文件轮转
func (r *Rotator) asyncWork() {
	const (
		ContextTimeout          = time.Second * 10
		InitExponentialInterval = time.Millisecond * 5
		TickerInterval          = time.Millisecond * 10
	)

	notify := r.stg.NotifyRotate()
	ticker := time.NewTicker(TickerInterval)
	defer ticker.Stop()

	for range ticker.C {
		select {
		case _, ok := <-notify:
			if !ok {
				r.l.Println("notify channel closed")
				return
			}

			r.writeLock.Lock()
			info, err := r.f.Stat()
			if err != nil {
				r.writeLock.Unlock()
				r.l.Println("failed to stat file, cause: ", err.Error())
				continue
			}

			if float64(info.Size()) < RotateSizeThreshold*float64(r.maxSize) {
				r.writeLock.Unlock()
				continue
			}

			err = r.rotate()
			r.writeLock.Unlock()
			if err != nil {
				r.l.Printf("asyncWork: rotate error: %v", err)
			}
		default:
		}
	}
}
