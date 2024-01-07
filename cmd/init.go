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
	"errors"
	"fmt"
	"git-gasset/util"
	"github.com/kopia/kopia/repo"
	"github.com/kopia/kopia/repo/blob"
	"github.com/kopia/kopia/repo/blob/s3"
	"github.com/kopia/kopia/repo/blob/throttling"
	"github.com/kopia/kopia/repo/content"
	"github.com/spf13/cobra"
	"log"
	"math/rand"
	"os"
	"path/filepath"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Creates or connects to the Kopia repository",
	Long: `Creates or connects to the Kopia repository

Checks the existence of the Kopia config file and if exists uses
it to connect and if not, creates the repository.`,
	RunE: InitRun,
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	initCmd.Flags().BoolP("create", "c", false, "Creates the repository if not exists")
}

type initOptions struct {
	workingDirectory string
	config           *util.Config
	kopiaConfig      *repo.LocalConfig
	password         string
	storage          blob.Storage
	gassetIdLength   int
	osGetwd          func() (string, error)
	osTempDir        func() string
	osUserConfigDir  func() (string, error)
	randIntn         func(n int) int
	s3New            func(ctx context.Context, opt *s3.Options, createIfNotExist bool) (blob.Storage, error)
	repoConnect      func(ctx context.Context, configFile string, st blob.Storage, password string, options *repo.ConnectOptions) error
	repoInitialize   func(ctx context.Context, st blob.Storage, opt *repo.NewRepositoryOptions, password string) error
}

func InitRun(cmd *cobra.Command, _ []string) error {
	log.Println("init called")

	initOptions := initOptions{
		gassetIdLength:  8,
		osGetwd:         os.Getwd,
		osTempDir:       os.TempDir,
		osUserConfigDir: os.UserConfigDir,
		randIntn:        rand.Intn,
		s3New:           s3.New,
		repoConnect:     repo.Connect,
		repoInitialize:  repo.Initialize,
	}

	if err := initOptions.initWorkingDirectory(); err != nil {
		return err
	}

	if err := initOptions.reloadKopiaConfig(); err != nil {
		return err
	}

	doCreate, err := cmd.Flags().GetBool("create")
	if err != nil {
		return err
	}

	err = initOptions.connect(doCreate)
	if err != nil {
		return err
	}
	return nil
}

func (op *initOptions) initWorkingDirectory() error {
	// Get the current working directory
	workingDirectory, err := op.osGetwd()
	if err != nil {
		return err
	}
	path, err := util.GetGitWorkingDirectory(workingDirectory)
	if err != nil {
		return err
	}
	op.workingDirectory = path
	return nil
}

// This function saves the "kopia" section of the .gasset file and reloads it using kopia APIs.
// This ensures that the kopia config conforms to the structure required.
func (op *initOptions) reloadKopiaConfig() error {
	config, err := util.GetConfig(op.workingDirectory)
	if err != nil {
		return err
	}
	op.config = config

	tempPath := filepath.Join(op.osTempDir(), "kopia.config")
	if err = util.WriteTempKopiaConfig(tempPath, config); err != nil {
		return err
	}
	kopiaConfig, err := repo.LoadConfigFromFile(tempPath)
	if err != nil {
		return err
	}
	op.kopiaConfig = kopiaConfig
	op.config.Kopia = kopiaConfig

	accessKey, secretKey, password, err := util.LoadKopiaSecretsFromEnv(op.workingDirectory)
	if err != nil {
		return err
	}
	if typedConfig, ok := kopiaConfig.Storage.Config.(*s3.Options); ok {
		typedConfig.AccessKeyID = accessKey
		typedConfig.SecretAccessKey = secretKey
	}
	op.password = password
	return nil
}

func (op *initOptions) connect(create bool) error {
	ctx := context.Background()

	storage, err := op.s3New(ctx, op.kopiaConfig.Storage.Config.(*s3.Options), false)
	if err != nil {
		return err
	}
	op.storage = storage

	if create {
		if err := op.createRepo(ctx); err != nil {
			return err
		}
	}

	if err := op.connectRepo(ctx); err != nil {
		return err
	}

	if create {

	}
	return nil
}

func (op *initOptions) connectRepo(ctx context.Context) error {
	kopiaUserConfigPath, err := op.getKopiaUserConfigPath()
	if err != nil {
		return err
	}
	return op.repoConnect(ctx, kopiaUserConfigPath, op.storage, op.password, &repo.ConnectOptions{
		ClientOptions:  op.kopiaConfig.ClientOptions,
		CachingOptions: content.CachingOptions{},
	})
}

func (op *initOptions) createRepo(ctx context.Context) error {
	if err := op.ensureEmpty(ctx, op.storage); err != nil {
		return err
	}

	if err := op.repoInitialize(ctx, op.storage, nil, op.password); err != nil {
		return err
	}

	// Set a random id as gasset id once the repo is initialized
	op.config.GassetId = util.GenerateRandomString(op.gassetIdLength, op.randIntn)

	// Yep, calling the caller connect function but with the "create" flag set to false.
	// After all, connect just tries to connect to the repo and will fail if it can't when the "create" flag is false.
	// This ensures that this function is not called again and thus ensuring that an endless loop doesn't happen
	if err := op.connect(false); err != nil {
		return err
	}

	//op.initPolicy()
	return util.UpdateGassetId(op.workingDirectory, op.config.GassetId)
}

func (op *initOptions) getKopiaUserConfigPath() (string, error) {
	if op.config.GassetId == "" {
		return "", errors.New("gasset id is empty")
	}
	userDir, err := op.osUserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userDir, "git-gasset", "kopia-"+op.config.GassetId+".config"), nil
}

// mostly from github.com/kopia/kopia/cli.commandRepositoryCreate.ensureEmpty
func (op *initOptions) ensureEmpty(ctx context.Context, storage blob.Storage) error {
	hasDataError := errors.New("has data")

	err := storage.ListBlobs(ctx, "", func(cb blob.Metadata) error {
		return hasDataError
	})
	if err == nil {
		return nil
	}

	if errors.Is(err, hasDataError) {
		return errors.New("found existing data in storage location")
	}

	return fmt.Errorf("error listing blobs: %w", err)
}

//func (op *initOptions) initPolicy(ctx context.Context) {
//	repo, err := repo.Open(ctx)
//}

func (op *initOptions) Clone() *initOptions {
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
	return &initOptions{
		workingDirectory: op.workingDirectory,
		config: &util.Config{
			Kopia:    copyKopia(op.config.Kopia),
			GassetId: op.config.GassetId,
		},
		kopiaConfig:     copyKopia(op.kopiaConfig),
		password:        op.password,
		storage:         op.storage,
		gassetIdLength:  op.gassetIdLength,
		osGetwd:         op.osGetwd,
		osTempDir:       op.osTempDir,
		osUserConfigDir: op.osUserConfigDir,
		randIntn:        op.randIntn,
		s3New:           op.s3New,
		repoConnect:     op.repoConnect,
		repoInitialize:  op.repoInitialize,
	}
}
