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
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/semaphore"
)

// initForTest 为测试程序创建多实例初始化，生产环境使用单例初始化
func initForTest(compress bool) error {
	var err error
	if compress {
		r, err = newRotator("./tests",
			fmt.Sprintf("testdata_%d.log", rand.Intn(1000)),
			WithCompress(CompressTypeGzip, GzipBestCompression),
			WithRotate(1024*1024*10, _Second))
	} else {
		r, err = newRotator("./tests",
			fmt.Sprintf("testdata_%d.log", rand.Intn(1000)),
			WithCompress(CompressTypeGzip, GzipBestSpeed),
			WithRotate(1024*1024*10, _Second))
	}

	return err
}

func TestNewFile(t *testing.T) {
	tf := time.Now().Format(Layout)
	testCases := []struct {
		name    string
		wantRes string
	}{
		{
			name:    "0002 count",
			wantRes: fmt.Sprintf("./tests/%s/testdata_%s_0002.log", tf, tf),
		},
		{
			name:    "0003 count",
			wantRes: fmt.Sprintf("./tests/%s/testdata_%s_0003.log", tf, tf),
		},
		{
			name:    "0004 count",
			wantRes: fmt.Sprintf("./tests/%s/testdata_%s_0004.log", tf, tf),
		},
	}

	r, err := NewRotator("./tests", "testdata.log")
	assert.Nil(t, err)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			f := r.newFile()
			assert.Equal(t, tc.wantRes, f)
		})
	}
}

func TestNewRotator_Compress(t *testing.T) {
	err := initForTest(true)
	assert.Nil(t, err)
	defer r.Close()

	const template = "测试数据，需要写入文件中，当前写入编号为：%d，测试内容。。。。。。。。。。\n"
	for i := 0; i < 100000; i++ {
		c := fmt.Sprintf(template, i)
		_, err = r.Write([]byte(c))
		assert.Nil(t, err)
	}
}

func TestNewRotator_No_Compress(t *testing.T) {
	err := initForTest(false)
	assert.Nil(t, err)
	defer r.Close()

	template := "测试数据，需要写入文件中，当前写入编号为：%d，测试内容。。。。。。。。。。\n"
	for i := 0; i < 30000; i++ {
		c := fmt.Sprintf(template, i)
		_, err = r.Write([]byte(c))
		assert.Nil(t, err)
	}
}

func TestNewRotator_Concurrent(t *testing.T) {
	err := initForTest(true)
	assert.Nil(t, err)
	defer r.Close()

	sem := semaphore.NewWeighted(100)
	template := "测试数据，需要写入文件中，当前写入编号为：%d，测试内容。。。。。。。。。。\n"
	var wg sync.WaitGroup
	for i := 0; i < 10000000; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		err = sem.Acquire(ctx, 1)
		cancel()
		assert.Nil(t, err)

		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			defer sem.Release(1)

			c := fmt.Sprintf(template, idx)
			_, localErr := r.Write([]byte(c))
			assert.Nil(t, localErr)
		}(i)
	}

	wg.Wait()
}
