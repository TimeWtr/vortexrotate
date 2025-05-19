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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/gozstd"
)

func TestNewGzip_Compress(t *testing.T) {
	w, err := os.OpenFile(filepath.Join("tests",
		compressFn("test.log", CompressTypeGzip)), os.O_CREATE|os.O_RDWR, ReadWriteFile)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	f, err := os.Open(filepath.Join("tests", "test.log"))
	assert.NoError(t, err)
	if err != nil {
		return
	}

	gs, err := NewGzip(w, f, GzipBestSpeed)
	assert.NoError(t, err)
	err = gs.Compress()
	assert.NoError(t, err)
	t.Log("Gzip cpr finished")
}

func TestNewGzip_Reset(t *testing.T) {
	w, err := os.OpenFile(filepath.Join("tests",
		compressFn("test.log", CompressTypeGzip)), os.O_CREATE|os.O_RDWR, ReadWriteFile)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	f, err := os.Open(filepath.Join("tests", "test.log"))
	assert.NoError(t, err)
	if err != nil {
		return
	}

	gs, err := NewGzip(w, f, GzipBestSpeed)
	assert.NoError(t, err)
	err = gs.Compress()
	assert.NoError(t, err)

	t.Log("Gzip cpr finished")

	w, err = os.OpenFile(filepath.Join("tests",
		compressFn("test.reset", CompressTypeGzip)), os.O_CREATE|os.O_RDWR, ReadWriteFile)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	f, err = os.Open(filepath.Join("tests", "test.reset.log"))
	assert.NoError(t, err)
	if err != nil {
		return
	}

	gs.Reset(w, f)
	err = gs.Compress()
	assert.NoError(t, err)
	t.Log("Gzip reset cpr finished")
}

func TestNewZstd_Compress(t *testing.T) {
	w, err := os.OpenFile(filepath.Join("tests",
		compressFn("test.log", CompressTypeZstd)), os.O_CREATE|os.O_RDWR, ReadWriteFile)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	f, err := os.Open(filepath.Join("tests", "test.log"))
	assert.NoError(t, err)
	if err != nil {
		return
	}

	zstd := NewZstd(w, f, gozstd.DefaultCompressionLevel)
	err = zstd.Compress()
	assert.NoError(t, err)
	t.Log("Zstd cpr finished")
}

func TestNewZstd_Reset(t *testing.T) {
	w, err := os.OpenFile(filepath.Join("tests",
		compressFn("test.log", CompressTypeZstd)), os.O_CREATE|os.O_RDWR, ReadWriteFile)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	f, err := os.Open(filepath.Join("tests", "test.log"))
	assert.NoError(t, err)
	if err != nil {
		return
	}

	zstd := NewZstd(w, f, gozstd.DefaultCompressionLevel)
	err = zstd.Compress()
	assert.NoError(t, err)
	t.Log("Zstd cpr finished")

	w, err = os.OpenFile(filepath.Join("tests",
		compressFn("test.reset", CompressTypeZstd)), os.O_CREATE|os.O_RDWR, ReadWriteFile)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	f, err = os.Open(filepath.Join("tests", "test.reset.log"))
	assert.NoError(t, err)
	if err != nil {
		return
	}

	zstd.Reset(w, f)
	err = zstd.Compress()
	assert.NoError(t, err)
	t.Log("Zstd reset cpr finished")
}

func TestNewSnappy_Compress(t *testing.T) {
	w, err := os.OpenFile(filepath.Join("tests",
		compressFn("test.log", CompressTypeSnappy)), os.O_CREATE|os.O_RDWR, ReadWriteFile)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	f, err := os.Open(filepath.Join("tests", "test.log"))
	assert.NoError(t, err)
	if err != nil {
		return
	}

	snappy := NewSnappy(w, f)
	err = snappy.Compress()
	assert.NoError(t, err)
	t.Log("Snappy cpr finished")
}

func TestNewSnappy_Reset(t *testing.T) {
	w, err := os.OpenFile(filepath.Join("tests",
		compressFn("test.log", CompressTypeSnappy)), os.O_CREATE|os.O_RDWR, ReadWriteFile)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	f, err := os.Open(filepath.Join("tests", "test.log"))
	assert.NoError(t, err)
	if err != nil {
		return
	}

	s := NewSnappy(w, f)
	err = s.Compress()
	assert.NoError(t, err)
	t.Log("Snappy cpr finished")

	w, err = os.OpenFile(filepath.Join("tests",
		compressFn("test.reset", CompressTypeSnappy)), os.O_CREATE|os.O_RDWR, ReadWriteFile)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	f, err = os.Open(filepath.Join("tests", "test.reset.log"))
	assert.NoError(t, err)
	if err != nil {
		return
	}

	s.Reset(w, f)
	err = s.Compress()
	assert.NoError(t, err)
	t.Log("Snappy reset cpr finished")
}
