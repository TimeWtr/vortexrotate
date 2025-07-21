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

import "sync"

// OnceWithError 用于确保一个初始化函数只执行一次，并保存其返回的错误信息。
// 结构体包含一个 noCopy 字段防止拷贝，一个 sync.Once 实例保证单次执行，以及一个 error 字段保存执行结果。
type OnceWithError struct {
	nocopy noCopy
	once   sync.Once
	err    error
}

// Do 方法接收一个无参数、返回 error 的函数 f。
// 该方法确保 f 只被执行一次，并将其返回值保存在 OnceWithError 实例的 err 字段中。
// 参数:
//
//	f - 初始化函数，返回一个 error。
func (o *OnceWithError) Do(f func() error) {
	o.once.Do(func() {
		o.err = f()
	})
}

// Err 方法用于获取 OnceWithError 实例中保存的错误信息。
// 返回值:
//
//	error - 初始化函数执行后保存的错误信息。
func (o *OnceWithError) Err() error {
	return o.err
}

type noCopy struct{}

func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}
