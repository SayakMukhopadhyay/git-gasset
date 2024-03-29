/*
Copyright © 2024 Sayak Mukhopadhyay

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenerateRandomString(t *testing.T) {
	type args struct {
		n        int
		randFunc func(n int) int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Generate a random string of length 10",
			args: args{n: 10, randFunc: func(_ int) int {
				return 0
			}},
			want: "0000000000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, GenerateRandomString(tt.args.n, tt.args.randFunc), "GenerateRandomString(%v)", tt.args.n)
		})
	}
}
