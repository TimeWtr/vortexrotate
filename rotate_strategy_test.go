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
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/semaphore"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func (s *MixStrategy) testSetLastTime(duration time.Duration) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.lastTime = time.Now().Add(-duration).UnixMilli()
}

func TestNewMixStrategy(t *testing.T) {
	ms, err := NewMixStrategy(1024*2, _Second)
	assert.NoError(t, err)
	defer ms.Close()

	ch := ms.NotifyRotate()
	closeCh := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()

		for {
			select {
			case <-closeCh:
				return
			case _, ok := <-ch:
				if !ok {
					return
				}
				t.Log("【消费者】rotate event receive success!")
			default:
			}
		}
	}()

	go func() {
		defer wg.Done()
		defer close(closeCh)

		for i := 0; i < 500; i++ {
			ms.testSetLastTime(150 * time.Millisecond)
			size := uint64(rand.Intn(200))
			if ms.ShouldRotate(size) {
				t.Logf("should rotate, index: %d, size: %d\n", i, size)
				continue
			}
			//t.Logf("should not rotate, write data successfully, index: %d, size: %d\n", i, size)
			time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
		}
	}()
	wg.Wait()
}

func TestNewMixStrategy_Concurrent(t *testing.T) {
	ms, err := NewMixStrategy(1024*2, _Second)
	assert.NoError(t, err)
	defer ms.Close()

	ch := ms.NotifyRotate()
	closeCh := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()

		for {
			select {
			case <-closeCh:
				return
			case _, ok := <-ch:
				if !ok {
					return
				}
				t.Log("【消费者】rotate event receive success!")
			default:
			}
		}
	}()

	go func() {
		defer wg.Done()
		defer close(closeCh)
		sem := semaphore.NewWeighted(100)
		for i := 0; i < 2000; i++ {
			err = sem.Acquire(context.TODO(), 1)
			assert.NoError(t, err)

			go func() {
				defer sem.Release(1)
				ms.testSetLastTime(150 * time.Millisecond)
				size := uint64(rand.Intn(200))
				if ms.ShouldRotate(size) {
					t.Logf("should rotate, index: %d, size: %d\n", i, size)
					return
				}
				//t.Logf("should not rotate, write data successfully, index: %d, size: %d\n", i, size)
				time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
			}()
		}
	}()
	wg.Wait()
}
