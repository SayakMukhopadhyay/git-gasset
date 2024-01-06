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
	workingDirectory string
	initOptions      *initOptions
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

func (suite *InitSuite) Test_initOptions_loadKopiaConfig() {
	tests := []struct {
		name    string
		fields  initOptions
		want    initOptions
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "Attempt loading the kopia config",
			fields:  *suite.initOptions,
			want:    *suite.initOptions,
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := tt.fields.loadKopiaConfig()
			if !tt.wantErr(suite.T(), err, fmt.Sprintf("loadKopiaConfig()")) {
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
			err := tt.fields.createRepo(tt.args.ctx, tt.args.storage)
			if !tt.wantErr(suite.T(), err, fmt.Sprintf("createRepo(%v, %v)", tt.args.ctx, tt.args.storage)) {
				return
			}
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
