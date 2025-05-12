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

// RotateStrategy 文件轮转的策略
type RotateStrategy interface {
	// ShouldRotate 是否应该执行文件轮转
	ShouldRotate() bool
	// Filename 轮转后新文件的名称
	Filename() string
}

type TimeStrategy struct {
}

func (t *TimeStrategy) ShouldRotate() bool {
	//TODO implement me
	panic("implement me")
}

func (t *TimeStrategy) Filename() string {
	//TODO implement me
	panic("implement me")
}

type SizeStrategy struct {
}

func NewSizeStrategy() RotateStrategy {
	return &SizeStrategy{}
}

func (s *SizeStrategy) ShouldRotate() bool {
	//TODO implement me
	panic("implement me")
}

func (s *SizeStrategy) Filename() string {
	//TODO implement me
	panic("implement me")
}
