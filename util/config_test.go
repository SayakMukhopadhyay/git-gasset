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
	"os"
	"testing"
)

func deleteFile(path string) error {
	return os.Remove(path)
}

type ConfigSuite struct {
	suite.Suite
	op OptionsForTest
}

func TestConfigSuite(t *testing.T) {
	suite.Run(t, new(ConfigSuite))
}

func (suite *ConfigSuite) SetupSuite() {
	err := SetupTestOptions(&suite.op)
	if err != nil {
		suite.T().FailNow()
	}
}

func (suite *ConfigSuite) TestGetConfig() {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    *Config
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Attempt to read a config file",
			args: args{
				path: "../mocks/",
			},
			want:    suite.op.OptionsWithHiddenSecrets.Config,
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			path := HandleAbsolutePath(suite.op.TestWorkingDirectory, tt.args.path)
			got, err := GetConfig(path)
			if !tt.wantErr(suite.T(), err, fmt.Sprintf("GetConfig(%v)", path)) {
				return
			}
			assert.Equalf(suite.T(), tt.want, got, "GetConfig(%v)", path)
		})
	}
}

func (suite *ConfigSuite) TestUpdateGassetId() {
	type args struct {
		path     string
		gassetId string
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Attempt to update a gasset id",
			args: args{
				path:     "../mocks/",
				gassetId: suite.op.OptionsWithGassetId.Config.GassetId,
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			path := HandleAbsolutePath(suite.op.TestWorkingDirectory, tt.args.path)
			tt.wantErr(suite.T(), UpdateGassetId(path, tt.args.gassetId), fmt.Sprintf("UpdateGassetId(%v, %v)", path, tt.args.gassetId))
		})
	}
}

func (suite *ConfigSuite) TestUpdateConfig() {
	type args struct {
		path   string
		config *Config
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Attempt to update a config file",
			args: args{
				path:   "../mocks/temp/.gasset",
				config: suite.op.OptionsWithGassetId.Config,
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			path := HandleAbsolutePath(suite.op.TestWorkingDirectory, tt.args.path)
			tt.wantErr(suite.T(), UpdateConfig(path, tt.args.config), fmt.Sprintf("UpdateConfig(%v, %v)", tt.args.path, tt.args.config))
			deleteFile(path)
		})
	}
}

func (suite *ConfigSuite) TestWriteTempKopiaConfig() {
	type args struct {
		path   string
		config *Config
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Attempt to write a temp kopia config file",
			args: args{
				path:   "../mocks/temp/kopia.config",
				config: suite.op.OptionsWithGassetId.Config,
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			path := HandleAbsolutePath(suite.op.TestWorkingDirectory, tt.args.path)
			tt.wantErr(suite.T(), WriteTempKopiaConfig(path, tt.args.config), fmt.Sprintf("WriteTempKopiaConfig(%v, %v)", tt.args.path, tt.args.config))
			deleteFile(path)
		})
	}
}

func (suite *ConfigSuite) TestLoadKopiaSecretsFromEnv() {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   string
		want2   string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "Attempt from .env file",
			args:    args{path: "../mocks"},
			want:    "accessid",
			want1:   "secret",
			want2:   "password",
			wantErr: assert.NoError,
		},
		{
			name:    "Attempt from a location without a .env file",
			args:    args{path: "../mocks/deep"},
			want:    "",
			want1:   "",
			want2:   "",
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			path := HandleAbsolutePath(suite.op.TestWorkingDirectory, tt.args.path)
			got, got1, got2, err := LoadKopiaSecretsFromEnv(path)
			if !tt.wantErr(suite.T(), err, fmt.Sprintf("LoadKopiaSecretsFromEnv(%v)", path)) {
				return
			}
			assert.Equalf(suite.T(), tt.want, got, "LoadKopiaSecretsFromEnv(%v)", path)
			assert.Equalf(suite.T(), tt.want1, got1, "LoadKopiaSecretsFromEnv(%v)", path)
			assert.Equalf(suite.T(), tt.want2, got2, "LoadKopiaSecretsFromEnv(%v)", path)
		})
	}
}

func (suite *ConfigSuite) TestGetGitWorkingDirectory() {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "Attempt from deep inside the git repository",
			args:    args{path: "../.idea/runConfigurations"},
			want:    "../",
			wantErr: assert.NoError,
		},
		{
			name:    "Attempt from the working directory of the git repository",
			args:    args{path: "../"},
			want:    "../",
			wantErr: assert.NoError,
		},
		{
			name:    "Attempt from deep inside the git repository which has a .git file",
			args:    args{path: "../mocks/deep/deeper"},
			want:    "../",
			wantErr: assert.NoError,
		},
		{
			name:    "Attempt from deep inside the git repository which has a .git file",
			args:    args{path: "/"},
			want:    "",
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			path := HandleAbsolutePath(suite.op.TestWorkingDirectory, tt.args.path)
			wantPath := HandleAbsolutePath(suite.op.TestWorkingDirectory, tt.want)
			got, err := GetGitWorkingDirectory(path)
			if !tt.wantErr(suite.T(), err, fmt.Sprintf("GetGitWorkingDirectory(%v)", path)) {
				return
			}
			assert.Equalf(suite.T(), wantPath, got, "GetGitWorkingDirectory(%v)", path)
		})
	}
}
