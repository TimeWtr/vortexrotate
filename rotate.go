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
	"github.com/valyala/gozstd"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	r             *Rotator
	onceWithError OnceWithError
)

const (
	WritingStatus = iota
	RotatingStatus
	RejectStatus
)

type Option func(*Rotator) error

// WithCompress 开启压缩
func WithCompress() Option {
	return func(r *Rotator) error {
		r.compress = true
		return nil
	}
}

// WithCompressType 压缩的算法类型
func WithCompressType(tp int) Option {
	return func(r *Rotator) error {
		if tp < _minCompressType || tp > _maxCompressType {
			return ErrCompressType
		}

		r.compressType = tp
		return nil
	}
}

// WithCompressLevel 压缩的级别
func WithCompressLevel(level int) Option {
	return func(r *Rotator) error {
		r.compressLevel = level
		return nil
	}
}

// WithPeriod 设置保存周期
func WithPeriod(period uint16) Option {
	return func(r *Rotator) error {
		r.period = period
		return nil
	}
}

// WithMaxCount 设置保存的最大文件数量
func WithMaxCount(count uint16) Option {
	return func(r *Rotator) error {
		r.maxCount = count
		return nil
	}
}

// WithMaxSize 设置单个文件写入的最大字节
func WithMaxSize(maxSize uint64) Option {
	return func(r *Rotator) error {
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
	suffix string
	// 当前写入的文件句柄
	f *os.File
	// 执行轮转的策略
	stg RotateStrategy
	// 锁保护
	lock sync.RWMutex
	// 当前的状态：写入、拒绝写入、轮转中
	status atomic.Uint32
	// 最大保留周期，天数
	period uint16
	// 保留的最大文件数量
	maxCount uint16
	// 清理过期文件的策略
	cleanup CleanUpStrategy
	// 是否执行压缩操作
	compress bool
	// 压缩类型(Gzip/Zstd)
	compressType int
	// 压缩级别
	compressLevel int
	// 压缩的策略
	cs CompressStrategy
	// 单个文件最大的大小
	maxSize uint64
	// 关闭信号
	sig atomic.Int32
	// 轮转计数器
	counter atomic.Uint32
}

// NewRotator 生产环境单例模式
func NewRotator(dir string, filename string, opts ...Option) (*Rotator, error) {
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
	sli := strings.Split(filename, ".")
	if len(sli) != 2 {
		return nil, ErrFilename
	}

	var rotator *Rotator
	rotator = &Rotator{
		dir:      dir,
		filename: sli[0],
		suffix:   sli[1],
		period:   DefaultPeriod,
		maxCount: DefaultMaxCount,
		maxSize:  DefaultMaxSize,
		compress: false,
		lock:     sync.RWMutex{},
	}

	rotator.sig.Store(0)
	rotator.counter.Store(1)

	if err := rotator.mkdirAll(); err != nil {
		return nil, err
	}

	f, err := os.OpenFile(rotator.newFile(), os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	rotator.f = f
	rotator.status.Store(WritingStatus)

	for _, opt := range opts {
		if err = opt(rotator); err != nil {
			return nil, err
		}
	}

	rotator.stg, err = NewMixStrategy(rotator.maxSize, Hour)
	if err != nil {
		return nil, err
	}

	if rotator.compress {
		// 如果没有设置压缩类型，默认使用Gzip压缩
		if rotator.compressType == CompressTypeUnknown {
			rotator.compressType = CompressTypeGzip
			rotator.cs, err = NewGzip(nil, f, GzipBestSpeed)
			if err != nil {
				return nil, err
			}
		}

		if rotator.compressLevel == 0 {
			if rotator.compressType == CompressTypeGzip {
				rotator.compressLevel = gzip.DefaultCompression
			}

			if rotator.compressType == CompressTypeZstd {
				rotator.compressLevel = gozstd.DefaultCompressionLevel
			}
		}
	}

	return rotator, nil
}

// Write 执行写入逻辑，判断大小是否已经达到最大大小，如果是则执行轮转逻辑
// 轮转后根据压缩配置执行压缩逻辑
func (r *Rotator) Write(p []byte) (int, error) {
	if r.sig.Load() == 1 {
		return 0, ErrRotateClosed
	}

	r.lock.Lock()
	defer r.lock.Unlock()
	for {
		if r.status.Load() == WritingStatus {
			break
		}

		time.Sleep(time.Millisecond)
	}

	if r.f == nil {
		return 0, os.ErrClosed
	}

	if r.stg.ShouldRotate(uint64(len(p))) {
		// 需要执行日志轮转
		//if err := r.rotate(); err != nil {
		//	return 0, err
		//}
	}

	n, err := r.f.Write(p)
	if err != nil {
		return n, err
	}

	return n, nil
}

func (r *Rotator) rotate() (err error) {
	r.status.Store(RotatingStatus)
	defer r.status.Store(WritingStatus)

	if r.compress {
		if err = r.cps(); err != nil {
			return err
		}
	}
	if r.f != nil {
		_ = r.f.Close()
	}

	f, err := os.OpenFile(r.newFile(), os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	r.f = f

	return nil
}

// cps 执行压缩操作
func (r *Rotator) cps() error {
	wf := compressFn(r.f.Name(), r.compressType)
	w, err := os.OpenFile(wf, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	r.cs.Reset(w, r.f)
	if err = r.cs.Compress(); err != nil {
		return err
	}

	return nil
}

func (r *Rotator) Close() {
	r.lock.Lock()
	defer r.lock.Unlock()

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
	return os.MkdirAll(fmt.Sprintf("%s/%s", r.dir, t), 0777)
}
