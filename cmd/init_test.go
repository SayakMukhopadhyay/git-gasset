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

package cmd

import (
	"context"
	"fmt"
	"git-gasset/util"
	"github.com/kopia/kopia/repo"
	"github.com/kopia/kopia/repo/blob"
	"github.com/kopia/kopia/repo/blob/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func handleAbsolutePath(wd string, path string) string {
	if !strings.HasPrefix(path, ".") {
		return path
	}
	return filepath.Join(wd, path)
}

type stubStorage struct {
	blob.Storage
}

func (s stubStorage) ListBlobs(context.Context, blob.ID, func(bm blob.Metadata) error) error {
	return nil
}

type InitSuite struct {
	suite.Suite
	workingDirectory          string
	initOptions               *initOptions
	initOptionsWithNoGassetId *initOptions
}

func TestInitSuite(t *testing.T) {
	suite.Run(t, new(InitSuite))
}

func (suite *InitSuite) SetupSuite() {
	workingDirectory, err := os.Getwd()
	if err != nil {
		suite.T().FailNow()
	}
	suite.workingDirectory = workingDirectory

	suite.initOptions = &initOptions{
		workingDirectory: handleAbsolutePath(suite.workingDirectory, "../mocks"),
		config: &util.Config{
			Kopia: &repo.LocalConfig{
				Storage: &blob.ConnectionInfo{
					Type: "s3",
					Config: &s3.Options{
						BucketName:      "bucket-name",
						Prefix:          "prefix/",
						Endpoint:        "endpoint.digitaloceanspaces.com",
						AccessKeyID:     "accessid",
						SecretAccessKey: "secret",
						SessionToken:    "",
					},
				},
				ClientOptions: repo.ClientOptions{
					Hostname:                "host-pc",
					Username:                "user",
					Description:             "prefix",
					EnableActions:           false,
					FormatBlobCacheDuration: 900000000000,
				},
			},
			GassetId: "0000000000",
		},
		kopiaConfig: &repo.LocalConfig{
			Storage: &blob.ConnectionInfo{
				Type: "s3",
				Config: &s3.Options{
					BucketName:      "bucket-name",
					Prefix:          "prefix/",
					Endpoint:        "endpoint.digitaloceanspaces.com",
					AccessKeyID:     "accessid",
					SecretAccessKey: "secret",
					SessionToken:    "",
				},
			},
			ClientOptions: repo.ClientOptions{
				Hostname:                "host-pc",
				Username:                "user",
				Description:             "prefix",
				EnableActions:           false,
				FormatBlobCacheDuration: 900000000000,
			},
		},
		password:       "password",
		storage:        stubStorage{},
		gassetIdLength: 10,
		osGetwd: func() (string, error) {
			return handleAbsolutePath(suite.workingDirectory, "."), nil
		},
		osTempDir: func() string {
			return handleAbsolutePath(suite.workingDirectory, "../mocks/temp")
		},
		osUserConfigDir: func() (string, error) {
			return handleAbsolutePath(suite.workingDirectory, "../mocks/user"), nil
		},
		randIntn: func(n int) int {
			return 0
		},
		s3New: func(ctx context.Context, opt *s3.Options, create bool) (blob.Storage, error) {
			return stubStorage{}, nil
		},
		repoConnect: func(ctx context.Context, configFile string, st blob.Storage, password string, options *repo.ConnectOptions) error {
			return nil
		},
		repoInitialize: func(ctx context.Context, st blob.Storage, opt *repo.NewRepositoryOptions, password string) error {
			return nil
		},
	}

	suite.initOptionsWithNoGassetId = suite.initOptions.Clone()
	suite.initOptionsWithNoGassetId.config.GassetId = ""
}

func (suite *InitSuite) Test_initOptions_initWorkingDirectory() {
	tests := []struct {
		name    string
		fields  initOptions
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "Attempt initialising working directory",
			fields:  *suite.initOptions,
			want:    handleAbsolutePath(suite.workingDirectory, "../"),
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := tt.fields.initWorkingDirectory()
			if !tt.wantErr(suite.T(), err, fmt.Sprintf("initWorkingDirectory()")) {
				return
			}
			assert.Equalf(suite.T(), tt.want, tt.fields.workingDirectory, fmt.Sprintf("initWorkingDirectory()"))
		})
	}
}

func (suite *InitSuite) Test_initOptions_reloadKopiaConfig() {
	tests := []struct {
		name    string
		fields  initOptions
		want    initOptions
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "Attempt loading the kopia config from temp file",
			fields:  *suite.initOptions,
			want:    *suite.initOptions,
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := tt.fields.reloadKopiaConfig()
			if !tt.wantErr(suite.T(), err, fmt.Sprintf("reloadKopiaConfig()")) {
				return
			}
			assert.Equalf(suite.T(), tt.want.kopiaConfig, tt.fields.kopiaConfig, fmt.Sprintf("initWorkingDirectory()"))
			assert.Equalf(suite.T(), tt.want.config, tt.fields.config, fmt.Sprintf("initWorkingDirectory()"))
			assert.Equalf(suite.T(), tt.want.password, tt.fields.password, fmt.Sprintf("initWorkingDirectory()"))
		})
	}
}

func (suite *InitSuite) Test_initOptions_connect() {
	type args struct {
		create bool
	}
	tests := []struct {
		name    string
		fields  initOptions
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "Connect to an existing S3 bucket",
			fields:  *suite.initOptions,
			args:    args{create: false},
			wantErr: assert.NoError,
		},
		{
			name:    "Connect to an existing S3 bucket with no gasset id registered",
			fields:  *suite.initOptionsWithNoGassetId,
			args:    args{create: false},
			wantErr: assert.Error,
		},
		{
			name:    "Create S3 bucket",
			fields:  *suite.initOptions,
			args:    args{create: true},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := tt.fields.connect(tt.args.create)
			if !tt.wantErr(suite.T(), err, fmt.Sprintf("connect(%v)", tt.args.create)) {
				return
			}
		})
	}
}

func (suite *InitSuite) Test_initOptions_connectRepo() {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  initOptions
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:   "Connect to an S3 repository",
			fields: *suite.initOptions,
			args: args{
				ctx: context.Background(),
			},
			wantErr: assert.NoError,
		},
		{
			name:   "Fail to connect to a repository without a gasset id",
			fields: *suite.initOptionsWithNoGassetId,
			args: args{
				ctx: context.Background(),
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := tt.fields.connectRepo(tt.args.ctx)
			tt.wantErr(suite.T(), err, fmt.Sprintf("connectRepo(%v)", tt.args.ctx))
		})
	}
}

func (suite *InitSuite) Test_initOptions_createRepo() {
	type args struct {
		ctx     context.Context
		storage blob.Storage
	}
	tests := []struct {
		name    string
		fields  initOptions
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:   "Create an S3 repository",
			fields: *suite.initOptions,
			args: args{
				ctx:     context.Background(),
				storage: stubStorage{},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := tt.fields.createRepo(tt.args.ctx)
			if !tt.wantErr(suite.T(), err, fmt.Sprintf("createRepo(%v)", tt.args.ctx)) {
				return
			}
		})
	}
}

func (suite *InitSuite) Test_initOptions_getKopiaUserConfigPath() {
	tests := []struct {
		name    string
		fields  initOptions
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "Attempt to get the kopia user config path",
			fields:  *suite.initOptions,
			want:    handleAbsolutePath(suite.workingDirectory, "../mocks/user/git-gasset/kopia-0000000000.config"),
			wantErr: assert.NoError,
		},
		{
			name:    "Should throw error on empty gasset id",
			fields:  *suite.initOptionsWithNoGassetId,
			want:    "",
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			got, err := tt.fields.getKopiaUserConfigPath()
			if !tt.wantErr(suite.T(), err, fmt.Sprintf("getKopiaUserConfigPath()")) {
				return
			}
			assert.Equalf(suite.T(), tt.want, got, "getKopiaUserConfigPath()")
		})
	}
}

func (suite *InitSuite) Test_initOptions_ensureEmpty() {
	type args struct {
		ctx     context.Context
		storage blob.Storage
	}
	tests := []struct {
		name    string
		fields  initOptions
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:   "Check if no error is thrown",
			fields: *suite.initOptions,
			args: args{
				ctx:     context.Background(),
				storage: stubStorage{},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := tt.fields.ensureEmpty(tt.args.ctx, tt.args.storage)
			if !tt.wantErr(suite.T(), err, fmt.Sprintf("ensureEmpty(%v, %v)", tt.args.ctx, tt.args.storage)) {
				return
			}
		})
	}
}

func (suite *InitSuite) Test_initOptions_initPolicy() {
	tests := []struct {
		name   string
		fields initOptions
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			//tt.fields.initPolicy()
		})
	}
}
