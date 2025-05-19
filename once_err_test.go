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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOnceWithError(t *testing.T) {
	testCases := []struct {
		name    string
		fn      func() error
		wantErr error
	}{
		{
			name: "error",
			fn: func() error {
				return errors.New("test error")
			},
			wantErr: errors.New("test error"),
		},
		{
			name: "nil",
			fn: func() error {
				return nil
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var onceErr OnceWithError
			onceErr.Do(tc.fn)
			assert.Equal(t, tc.wantErr, onceErr.Err())
		})
	}
}
