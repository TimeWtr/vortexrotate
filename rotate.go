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
	"github.com/valyala/gozstd"
	"os"
	"sync"
	"sync/atomic"
)

const (
	// DefaultPeriod 默认保存的天数，30天
	DefaultPeriod = 30
	// DefaultMaxCount 默认保存的最大文件数量，100个
	DefaultMaxCount = 100
	// DefaultMaxSize 默认单个日志文件保存的最大大小，100MB
	DefaultMaxSize = 1024 * 1024 * 100
)

type Option func(*Rotator) error

// WithStrategy 设置轮转策略
func WithStrategy(stg RotateStrategy) Option {
	return func(r *Rotator) error {
		r.stg = stg
		return nil
	}
}

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
	// 当前写入的文件句柄
	f *os.File
	// 执行轮转的策略
	stg RotateStrategy
	// 锁保护
	lock sync.Mutex
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
	// 当前文件写入的大小
	size atomic.Uint64
	// 关闭信号
	sig atomic.Int32
}

func NewRotator(dir, filename string, opts ...Option) (*Rotator, error) {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	r := &Rotator{
		dir:      dir,
		filename: filename,
		f:        f,
		stg:      NewSizeStrategy(),
		period:   DefaultPeriod,
		maxCount: DefaultMaxCount,
		maxSize:  DefaultMaxSize,
		compress: false,
		lock:     sync.Mutex{},
	}
	r.size.Store(0)
	r.sig.Store(0)

	for _, opt := range opts {
		if err = opt(r); err != nil {
			return r, err
		}
	}

	if r.compress {
		// 如果没有设置压缩类型，默认使用Gzip压缩
		if r.compressType == CompressTypeUnknown {
			r.compressType = CompressTypeGzip
			r.cs, err = NewGzip(nil, f, GzipBestSpeed)
			if err != nil {
				return nil, err
			}
		}

		if r.compressLevel == 0 {
			if r.compressType == CompressTypeGzip {
				r.compressLevel = gzip.DefaultCompression
			}

			if r.compressType == CompressTypeZstd {
				r.compressLevel = gozstd.DefaultCompressionLevel
			}
		}
	}

	return r, nil
}

// Write 执行写入逻辑，判断大小是否已经达到最大大小，如果是则执行轮转逻辑
// 轮转后根据压缩配置执行压缩逻辑
func (r *Rotator) Write(p []byte) (int, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.f == nil {
		return 0, os.ErrClosed
	}

	if r.stg.ShouldRotate() {
		// 需要执行日志轮转
		if err := r.rotate(); err != nil {
			return 0, err
		}
	}

	n, err := r.f.Write(p)
	if err != nil {
		return n, err
	}

	r.size.Add(uint64(n))
	return n, nil
}

func (r *Rotator) rotate() error {
	if r.compress {
		if err := r.cps(); err != nil {
			return err
		}
	}

	f, err := os.OpenFile(r.stg.Filename(), os.O_RDONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	r.f = f

	return nil
}

// cps 执行压缩操作
func (r *Rotator) cps() error {
	wf := compressFn(r.f.Name(), r.compressType)
	w, err := os.OpenFile(wf, os.O_CREATE|os.O_APPEND|os.O_APPEND, 0666)
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
	r.sig.Store(1)
	if r.f == nil {
		return
	}

	_ = r.f.Close()
}
