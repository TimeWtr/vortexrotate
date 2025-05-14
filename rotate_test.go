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
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewFile(t *testing.T) {
	tf := time.Now().Format("20060102")
	testCases := []struct {
		name    string
		wantRes string
	}{
		{
			name:    "0001 count",
			wantRes: fmt.Sprintf("./tests/%s/testdata_%s_0001.log", tf, tf),
		},
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
