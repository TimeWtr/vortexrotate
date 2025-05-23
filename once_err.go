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

// OnceWithError once执行返回运行的错误信息
type OnceWithError struct {
	nocopy noCopy
	once   sync.Once
	err    error
}

func (o *OnceWithError) Do(f func() error) {
	o.once.Do(func() {
		o.err = f()
	})
}

func (o *OnceWithError) Err() error {
	return o.err
}

type noCopy struct{}

func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}
