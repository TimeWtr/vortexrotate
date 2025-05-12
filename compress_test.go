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
	"github.com/stretchr/testify/assert"
	"github.com/valyala/gozstd"
	"os"
	"path/filepath"
	"testing"
)

func TestNewGzip(t *testing.T) {
	w, err := os.OpenFile(filepath.Join("./tests", "test.l.gz"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	f, err := os.Open(filepath.Join("./tests", "test.log"))
	assert.NoError(t, err)
	if err != nil {
		return
	}

	gs, err := NewGzip(w, f, GzipBestSpeed)
	assert.NoError(t, err)
	err = gs.Compress()
	assert.NoError(t, err)
	t.Log("Gzip compress finished")
}

func TestNewZstd(t *testing.T) {
	w, err := os.OpenFile(filepath.Join("./tests", "test.l.zst"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	f, err := os.Open(filepath.Join("./tests", "test.log"))
	assert.NoError(t, err)
	if err != nil {
		return
	}

	zstd := NewZstd(w, f, gozstd.DefaultCompressionLevel)
	err = zstd.Compress()
	assert.NoError(t, err)
	t.Log("Zstd compress finished")
}
