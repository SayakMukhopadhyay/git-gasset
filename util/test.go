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
	"context"
	"github.com/kopia/kopia/repo"
	"github.com/kopia/kopia/repo/blob"
	"github.com/kopia/kopia/repo/blob/s3"
	"github.com/kopia/kopia/snapshot"
	"github.com/kopia/kopia/snapshot/policy"
	"os"
	"path/filepath"
	"strings"
)

type OptionsForTest struct {
	TestWorkingDirectory     string
	OptionsWithGassetId      *Options
	OptionsWithNoGassetId    *Options
	OptionsWithHiddenSecrets *Options
}

func HandleAbsolutePath(wd string, path string) string {
	if !strings.HasPrefix(path, ".") {
		return path
	}
	return filepath.Join(wd, path)
}

type StubStorage struct {
	blob.Storage
}

func (s StubStorage) ListBlobs(context.Context, blob.ID, func(bm blob.Metadata) error) error {
	return nil
}

func SetupTestOptions(options *OptionsForTest) error {
	workingDirectory, err := os.Getwd()
	if err != nil {
		return err
	}
	options.TestWorkingDirectory = workingDirectory

	options.OptionsWithGassetId = &Options{
		WorkingDirectory: HandleAbsolutePath(options.TestWorkingDirectory, "../mocks"),
		Config: &Config{
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
		Password:       "password",
		Storage:        StubStorage{},
		GassetIdLength: 10,
		OsGetwd: func() (string, error) {
			return HandleAbsolutePath(options.TestWorkingDirectory, "."), nil
		},
		OsTempDir: func() string {
			return HandleAbsolutePath(options.TestWorkingDirectory, "../mocks/temp")
		},
		OsUserConfigDir: func() (string, error) {
			return HandleAbsolutePath(options.TestWorkingDirectory, "../mocks/user"), nil
		},
		RandIntn: func(n int) int {
			return 0
		},
		S3New: func(ctx context.Context, opt *s3.Options, create bool) (blob.Storage, error) {
			return StubStorage{}, nil
		},
		RepoConnect: func(ctx context.Context, configFile string, st blob.Storage, password string, options *repo.ConnectOptions) error {
			return nil
		},
		RepoInitialize: func(ctx context.Context, st blob.Storage, opt *repo.NewRepositoryOptions, password string) error {
			return nil
		},
		RepoOpen: func(ctx context.Context, configFile string, password string, options *repo.Options) (rep repo.Repository, err error) {
			return nil, nil
		},
		RepoWriteSession: func(ctx context.Context, r repo.Repository, opt repo.WriteSessionOptions, cb func(ctx context.Context, w repo.RepositoryWriter) error) error {
			return cb(ctx, nil)
		},
		PolicySetPolicy: func(ctx context.Context, r repo.RepositoryWriter, si snapshot.SourceInfo, pol *policy.Policy) error {
			return nil
		},
	}

	options.OptionsWithNoGassetId = options.OptionsWithGassetId.Clone()
	options.OptionsWithNoGassetId.Config.GassetId = ""

	options.OptionsWithHiddenSecrets = options.OptionsWithGassetId.Clone()
	options.OptionsWithHiddenSecrets.Config.Kopia.Storage.Config.(*s3.Options).AccessKeyID = "someaccesskey"
	options.OptionsWithHiddenSecrets.Config.Kopia.Storage.Config.(*s3.Options).SecretAccessKey = "somesecret"

	return nil
}
