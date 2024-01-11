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
	"github.com/kopia/kopia/repo/blob"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type InitSuite struct {
	suite.Suite
	*util.OptionsForTest
}

func TestInitSuite(t *testing.T) {
	suite.Run(t, new(InitSuite))
}

func (suite *InitSuite) SetupSuite() {
	suite.OptionsForTest = &util.OptionsForTest{}
	if err := util.SetupTestOptions(suite.OptionsForTest); err != nil {
		suite.T().FailNow()
	}
}

func (suite *InitSuite) Test_initOptions_connect() {
	type args struct {
		options *util.Options
		create  bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "Connect to an existing S3 bucket",
			args:    args{options: suite.OptionsWithGassetId, create: false},
			wantErr: assert.NoError,
		},
		{
			name:    "Connect to an existing S3 bucket with no gasset id registered",
			args:    args{options: suite.OptionsWithNoGassetId, create: false},
			wantErr: assert.Error,
		},
		{
			name:    "Create S3 bucket",
			args:    args{options: suite.OptionsWithGassetId, create: true},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := connect(tt.args.options, tt.args.create)
			if !tt.wantErr(suite.T(), err, fmt.Sprintf("connect(%v)", tt.args.create)) {
				return
			}
		})
	}
}

func (suite *InitSuite) Test_initOptions_connectRepo() {
	type args struct {
		ctx     context.Context
		options *util.Options
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Connect to an S3 repository",
			args: args{
				ctx:     context.Background(),
				options: suite.OptionsWithGassetId,
			},
			wantErr: assert.NoError,
		},
		{
			name: "Fail to connect to a repository without a gasset id",
			args: args{
				ctx:     context.Background(),
				options: suite.OptionsWithNoGassetId,
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := connectRepo(tt.args.ctx, tt.args.options)
			tt.wantErr(suite.T(), err, fmt.Sprintf("connectRepo(%v)", tt.args.ctx))
		})
	}
}

func (suite *InitSuite) Test_initOptions_createRepo() {
	type args struct {
		ctx     context.Context
		options *util.Options
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Create an S3 repository",
			args: args{
				ctx:     context.Background(),
				options: suite.OptionsWithGassetId,
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := createRepo(tt.args.ctx, tt.args.options)
			if !tt.wantErr(suite.T(), err, fmt.Sprintf("createRepo(%v)", tt.args.ctx)) {
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
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Check if no error is thrown",
			args: args{
				ctx:     context.Background(),
				storage: util.StubStorage{},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := ensureEmpty(tt.args.ctx, tt.args.storage)
			if !tt.wantErr(suite.T(), err, fmt.Sprintf("ensureEmpty(%v, %v)", tt.args.ctx, tt.args.storage)) {
				return
			}
		})
	}
}

func (suite *InitSuite) Test_initOptions_initPolicy() {
	type args struct {
		ctx     context.Context
		options *util.Options
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Check if no error is thrown",
			args: args{
				ctx:     context.Background(),
				options: suite.OptionsWithGassetId,
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := initPolicy(tt.args.ctx, tt.args.options)
			if !tt.wantErr(suite.T(), err, fmt.Sprintf("initPolicy(%v)", tt.args.ctx)) {
				return
			}
		})
	}
}
