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
	"github.com/golang/snappy"
	"github.com/valyala/gozstd"
	"io"
	"os"
)

const (
	CompressTypeUnknown = iota
	CompressTypeGzip
	CompressTypeZstd
	CompressTypeSnappy

	_minCompressType = CompressTypeGzip
	_maxCompressType = CompressTypeSnappy
)

const bufferSize = 128 * 1024

// Gzip压缩的等级
const (
	GzipBestSpeed          = gzip.BestSpeed
	GzipBestCompression    = gzip.BestCompression
	GzipDefaultCompression = gzip.DefaultCompression
	GzipHuffmanOnly        = gzip.HuffmanOnly
)

func CompressFn(fn string, tp int) string {
	switch tp {
	case CompressTypeGzip:
		return fmt.Sprintf("%s.gz", fn)
	case CompressTypeZstd:
		return fmt.Sprintf("%s.zst", fn)
	case CompressTypeSnappy:
		return fmt.Sprintf("%s.snappy", fn)
	default:
		return ""
	}
}

// CompressStrategy 压缩策略，对文件执行压缩操作
type CompressStrategy interface {
	// Compress 执行压缩逻辑
	Compress() error
	// Reset 重置压缩
	Reset(w io.Writer, f *os.File)
}

type Gzip struct {
	w *gzip.Writer
	f *os.File
}

func NewGzip(outFile io.Writer, f *os.File, compressLevel int) (CompressStrategy, error) {
	w, err := gzip.NewWriterLevel(outFile, compressLevel)
	if err != nil {
		return nil, err
	}

	return &Gzip{
		w: w,
		f: f,
	}, nil
}

func (g *Gzip) Compress() error {
	defer func() {
		_ = g.w.Close()
	}()

	if g.f == nil {
		return os.ErrClosed
	}
	defer func() {
		_ = g.f.Close()
	}()

	bs := make([]byte, bufferSize)
	for {
		n, err := g.f.Read(bs)
		if err != nil && err != io.EOF {
			return err
		}

		if n == 0 || err == io.EOF {
			break
		}

		if _, err = g.w.Write(bs[:n]); err != nil {
			return err
		}
	}

	return g.w.Flush()
}

func (g *Gzip) Reset(w io.Writer, f *os.File) {
	g.w.Reset(w)
	g.f = f
}

type Zstd struct {
	w *gozstd.Writer
	f *os.File
	l int
}

func NewZstd(outFile io.Writer, f *os.File, compressLevel int) CompressStrategy {
	return &Zstd{
		w: gozstd.NewWriterLevel(outFile, compressLevel),
		f: f,
		l: compressLevel,
	}
}

func (z *Zstd) Compress() error {
	defer func() {
		_ = z.w.Close()
	}()

	if z.f == nil {
		return os.ErrClosed
	}
	defer func() {
		_ = z.f.Close()
	}()

	bs := make([]byte, bufferSize)
	for {
		n, err := z.f.Read(bs)
		if err != nil && err != io.EOF {
			return err
		}

		if err == io.EOF || n == 0 {
			break
		}

		if _, err = z.w.Write(bs[:n]); err != nil {
			return err
		}
	}

	return z.w.Flush()
}

func (z *Zstd) Reset(w io.Writer, f *os.File) {
	z.w = gozstd.NewWriterLevel(w, z.l)
	z.f = f
}

type Snappy struct {
	w *snappy.Writer
	f *os.File
}

func NewSnappy(outFile io.Writer, f *os.File) CompressStrategy {
	return &Snappy{
		w: snappy.NewWriter(outFile),
		f: f,
	}
}

func (s *Snappy) Compress() error {
	defer func() {
		_ = s.w.Close()
	}()

	if s.f == nil {
		return os.ErrClosed
	}
	defer func() {
		_ = s.f.Close()
	}()

	bs := make([]byte, bufferSize)
	for {
		n, err := s.f.Read(bs)
		if err != nil && err != io.EOF {
			return err
		}

		if err == io.EOF || n == 0 {
			break
		}

		if _, err = s.w.Write(bs[:n]); err != nil {
			return err
		}
	}

	return s.w.Flush()
}

func (s *Snappy) Reset(w io.Writer, f *os.File) {
	s.w.Reset(w)
	s.f = f
}
