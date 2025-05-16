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

package pkg

import (
	"github.com/TimeWtr/vortexrotate"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNil(t *testing.T) {
	res := IsNil(interface{}(nil))
	assert.Equal(t, true, res)

	var i vortexrotate.RotateStrategy
	res = IsNil(i)
	assert.Equal(t, true, res)
}

func TestIsNil(t *testing.T) {
	type test struct {
		stg vortexrotate.RotateStrategy
	}
	app := &test{}
	if IsNil(app.stg) {
		t.Log("nil")
	} else {
		t.Log("not nil")
	}

	stg, err := vortexrotate.NewMixStrategy(1024, vortexrotate._Second)
	assert.Nil(t, err)
	app.stg = stg
	if IsNil(app.stg) {
		t.Log("nil")
	} else {
		t.Log("not nil")
	}

	app.stg = nil
	if IsNil(app.stg) {
		t.Log("nil")
	} else {
		t.Log("not nil")
	}
}
