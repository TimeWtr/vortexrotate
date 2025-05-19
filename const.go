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

import "os"

const Layout = "20060102"

const (
	// DefaultPeriod 默认保存的天数，30天
	DefaultPeriod = 30
	// DefaultMaxCount 默认保存的最大文件数量，100个
	DefaultMaxCount = 100
	// DefaultMaxSize 默认单个日志文件保存的最大大小，100MB
	DefaultMaxSize = 1024 * 1024 * 100
)

// 文件系统操作权限组
const (
	ReadOnlyFile  os.FileMode = 0o444 // 只读文件
	ReadWriteFile os.FileMode = 0o644 // 读写文件
)
