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

package errorx

import (
	"errors"
)

var (
	ErrCompressType = errors.New("cpr type not support")
	ErrTimeType     = errors.New("rotate cron time type not support")
	ErrRotateClosed = errors.New("rotate is closed")
	ErrCompress     = errors.New("cpr strategy error")
)

var ErrFilename = errors.New("filename must contain exactly one '.' character")

type Error struct {
	err error
}

func (e *Error) UnWrapError() error {
	return e.err
}

func (e *Error) Error() string {
	if e.err == nil {
		return "<nil>"
	}

	return e.err.Error()
}
