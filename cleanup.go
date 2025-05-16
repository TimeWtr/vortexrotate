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
	"sync/atomic"
	"time"
)

// Cleanup 清理过期文件的配置
type Cleanup struct {
	// 最大保留周期，天数
	period uint16
	// 保留的最大文件数量
	maxCount uint16
	// 清理过期文件的策略
	cleanup CleanUpStrategy
}

type CleanUpStrategy interface {
	CleanUp() error
	Add()
}

// FileCountCleanUp 根据文件最大数量来确定是否执行清理
type FileCountCleanUp struct {
	count atomic.Uint64
	// 最大数量
	maxCount uint64
}

func NewFileCountCleanUp(maxCount uint64) CleanUpStrategy {
	fc := FileCountCleanUp{
		maxCount: maxCount,
	}
	fc.count.Store(0)

	return &fc
}

func (f *FileCountCleanUp) CleanUp() error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if f.count.Load() < f.maxCount {
				continue
			}

			// 执行清理逻辑

		default:
		}
	}
}

func (f *FileCountCleanUp) Add() {
	f.count.Add(1)
}
