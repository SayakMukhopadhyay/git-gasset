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
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type OptionsSuite struct {
	suite.Suite
	op OptionsForTest
}

func TestOptionsSuite(t *testing.T) {
	suite.Run(t, new(OptionsSuite))
}

func (suite *OptionsSuite) SetupSuite() {
	err := SetupTestOptions(&suite.op)
	if err != nil {
		suite.T().FailNow()
	}
}

func (suite *OptionsSuite) TestInitWorkingDirectory() {
	tests := []struct {
		name    string
		fields  Options
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "Attempt initialising working directory",
			fields:  *suite.op.OptionsWithGassetId,
			want:    HandleAbsolutePath(suite.op.TestWorkingDirectory, "../"),
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := tt.fields.InitWorkingDirectory()
			if !tt.wantErr(suite.T(), err, fmt.Sprintf("initWorkingDirectory()")) {
				return
			}
			assert.Equalf(suite.T(), tt.want, tt.fields.WorkingDirectory, fmt.Sprintf("initWorkingDirectory()"))
		})
	}
}

func (suite *OptionsSuite) TestReloadKopiaConfig() {
	tests := []struct {
		name    string
		fields  Options
		want    Options
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "Attempt loading the kopia config from temp file",
			fields:  *suite.op.OptionsWithGassetId,
			want:    *suite.op.OptionsWithGassetId,
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := tt.fields.ReloadKopiaConfig()
			if !tt.wantErr(suite.T(), err, fmt.Sprintf("reloadKopiaConfig()")) {
				return
			}
			assert.Equalf(suite.T(), tt.want.Config, tt.fields.Config, fmt.Sprintf("initWorkingDirectory()"))
			assert.Equalf(suite.T(), tt.want.Password, tt.fields.Password, fmt.Sprintf("initWorkingDirectory()"))
		})
	}
}

func (suite *OptionsSuite) TestGetKopiaUserConfigPath() {
	tests := []struct {
		name    string
		fields  Options
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "Attempt to get the kopia user config path",
			fields:  *suite.op.OptionsWithGassetId,
			want:    HandleAbsolutePath(suite.op.TestWorkingDirectory, "../mocks/user/git-gasset/kopia-0000000000.config"),
			wantErr: assert.NoError,
		},
		{
			name:    "Should throw error on empty gasset id",
			fields:  *suite.op.OptionsWithNoGassetId,
			want:    "",
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			got, err := tt.fields.GetKopiaUserConfigPath()
			if !tt.wantErr(suite.T(), err, fmt.Sprintf("getKopiaUserConfigPath()")) {
				return
			}
			assert.Equalf(suite.T(), tt.want, got, "getKopiaUserConfigPath()")
		})
	}
}
