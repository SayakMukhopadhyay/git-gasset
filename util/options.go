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
	"errors"
	"github.com/kopia/kopia/repo"
	"github.com/kopia/kopia/repo/blob"
	"github.com/kopia/kopia/repo/blob/s3"
	"github.com/kopia/kopia/repo/blob/throttling"
	"github.com/kopia/kopia/repo/content"
	"github.com/kopia/kopia/snapshot"
	"github.com/kopia/kopia/snapshot/policy"
	"path/filepath"
)

type Options struct {
	WorkingDirectory string
	Config           *Config
	KopiaConfig      *repo.LocalConfig
	Password         string
	Storage          blob.Storage
	GassetIdLength   int
	OsGetwd          func() (string, error)
	OsTempDir        func() string
	OsUserConfigDir  func() (string, error)
	RandIntn         func(n int) int
	S3New            func(ctx context.Context, opt *s3.Options, createIfNotExist bool) (blob.Storage, error)
	RepoConnect      func(ctx context.Context, configFile string, st blob.Storage, password string, options *repo.ConnectOptions) error
	RepoInitialize   func(ctx context.Context, st blob.Storage, opt *repo.NewRepositoryOptions, password string) error
	RepoOpen         func(ctx context.Context, configFile string, password string, options *repo.Options) (rep repo.Repository, err error)
	RepoWriteSession func(ctx context.Context, r repo.Repository, opt repo.WriteSessionOptions, cb func(ctx context.Context, w repo.RepositoryWriter) error) error
	PolicySetPolicy  func(ctx context.Context, r repo.RepositoryWriter, si snapshot.SourceInfo, pol *policy.Policy) error
}

func (op *Options) InitWorkingDirectory() error {
	// Get the current working directory
	workingDirectory, err := op.OsGetwd()
	if err != nil {
		return err
	}
	path, err := GetGitWorkingDirectory(workingDirectory)
	if err != nil {
		return err
	}
	op.WorkingDirectory = path
	return nil
}

// ReloadKopiaConfig  saves the "kopia" section of the .gasset file and reloads it using kopia APIs.
// This ensures that the kopia config conforms to the structure required.
func (op *Options) ReloadKopiaConfig() error {
	config, err := GetConfig(op.WorkingDirectory)
	if err != nil {
		return err
	}
	op.Config = config

	tempPath := filepath.Join(op.OsTempDir(), "kopia.config")
	if err = WriteTempKopiaConfig(tempPath, config); err != nil {
		return err
	}
	kopiaConfig, err := repo.LoadConfigFromFile(tempPath)
	if err != nil {
		return err
	}
	op.KopiaConfig = kopiaConfig
	op.Config.Kopia = kopiaConfig

	accessKey, secretKey, password, err := LoadKopiaSecretsFromEnv(op.WorkingDirectory)
	if err != nil {
		return err
	}
	if typedConfig, ok := kopiaConfig.Storage.Config.(*s3.Options); ok {
		typedConfig.AccessKeyID = accessKey
		typedConfig.SecretAccessKey = secretKey
	}
	op.Password = password
	return nil
}

func (op *Options) GetKopiaUserConfigPath() (string, error) {
	if op.Config.GassetId == "" {
		return "", errors.New("gasset id is empty")
	}
	userDir, err := op.OsUserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userDir, "git-gasset", "kopia-"+op.Config.GassetId+".config"), nil
}

func (op *Options) Clone() *Options {
	copyKopia := func(l *repo.LocalConfig) *repo.LocalConfig {
		var apiServer *repo.APIServerInfo
		var storage *blob.ConnectionInfo
		var caching *content.CachingOptions
		var clientOptions repo.ClientOptions

		if l.APIServer != nil {
			apiServer = &repo.APIServerInfo{
				BaseURL:                             l.APIServer.BaseURL,
				TrustedServerCertificateFingerprint: l.APIServer.TrustedServerCertificateFingerprint,
				DisableGRPC:                         l.APIServer.DisableGRPC,
			}
		}

		if l.Storage != nil {
			storage = &blob.ConnectionInfo{
				Type:   l.Storage.Type,
				Config: l.Storage.Config,
			}
		}

		if l.Caching != nil {
			caching = &content.CachingOptions{
				CacheDirectory:              l.Caching.CacheDirectory,
				ContentCacheSizeBytes:       l.Caching.ContentCacheSizeBytes,
				ContentCacheSizeLimitBytes:  l.Caching.ContentCacheSizeLimitBytes,
				MetadataCacheSizeBytes:      l.Caching.MetadataCacheSizeBytes,
				MetadataCacheSizeLimitBytes: l.Caching.MetadataCacheSizeLimitBytes,
				MaxListCacheDuration:        l.Caching.MaxListCacheDuration,
				MinMetadataSweepAge:         l.Caching.MinMetadataSweepAge,
				MinContentSweepAge:          l.Caching.MinContentSweepAge,
				MinIndexSweepAge:            l.Caching.MinIndexSweepAge,
				HMACSecret:                  l.Caching.HMACSecret,
			}
		}

		if l.ClientOptions != (repo.ClientOptions{}) {
			var throttlingParam *throttling.Limits

			if l.ClientOptions.Throttling != nil {
				throttlingParam = &throttling.Limits{
					ReadsPerSecond:         l.ClientOptions.Throttling.ReadsPerSecond,
					WritesPerSecond:        l.ClientOptions.Throttling.WritesPerSecond,
					ListsPerSecond:         l.ClientOptions.Throttling.ListsPerSecond,
					UploadBytesPerSecond:   l.ClientOptions.Throttling.UploadBytesPerSecond,
					DownloadBytesPerSecond: l.ClientOptions.Throttling.DownloadBytesPerSecond,
					ConcurrentReads:        l.ClientOptions.Throttling.ConcurrentReads,
					ConcurrentWrites:       l.ClientOptions.Throttling.ConcurrentWrites,
				}
			}

			clientOptions = repo.ClientOptions{
				Hostname:                l.ClientOptions.Hostname,
				Username:                l.ClientOptions.Username,
				ReadOnly:                l.ClientOptions.ReadOnly,
				PermissiveCacheLoading:  l.ClientOptions.PermissiveCacheLoading,
				Description:             l.ClientOptions.Description,
				EnableActions:           l.ClientOptions.EnableActions,
				FormatBlobCacheDuration: l.ClientOptions.FormatBlobCacheDuration,
				Throttling:              throttlingParam,
			}
		}
		return &repo.LocalConfig{
			APIServer:     apiServer,
			Storage:       storage,
			Caching:       caching,
			ClientOptions: clientOptions,
		}
	}
	return &Options{
		WorkingDirectory: op.WorkingDirectory,
		Config: &Config{
			Kopia:    copyKopia(op.Config.Kopia),
			GassetId: op.Config.GassetId,
		},
		KopiaConfig:      copyKopia(op.KopiaConfig),
		Password:         op.Password,
		Storage:          op.Storage,
		GassetIdLength:   op.GassetIdLength,
		OsGetwd:          op.OsGetwd,
		OsTempDir:        op.OsTempDir,
		OsUserConfigDir:  op.OsUserConfigDir,
		RandIntn:         op.RandIntn,
		S3New:            op.S3New,
		RepoConnect:      op.RepoConnect,
		RepoInitialize:   op.RepoInitialize,
		RepoOpen:         op.RepoOpen,
		RepoWriteSession: op.RepoWriteSession,
		PolicySetPolicy:  op.PolicySetPolicy,
	}
}
