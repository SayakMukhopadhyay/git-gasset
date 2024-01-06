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
}

func InitRun(cmd *cobra.Command, args []string) error {
	log.Println("init called")

	initOptions := initOptions{}

	if err := initOptions.initWorkingDirectory(); err != nil {
		return err
	}

	if err := initOptions.loadKopiaConfig(); err != nil {
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
	workingDirectory, err := os.Getwd()
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

func (op *initOptions) loadKopiaConfig() error {
	config, err := util.GetConfig(op.workingDirectory)
	if err != nil {
		return err
	}
	op.config = config

	tempPath := filepath.Join(os.TempDir(), "kopia.config")
	if err = util.WriteTempKopiaConfig(tempPath, config); err != nil {
		return err
	}
	kopiaConfig, err := repo.LoadConfigFromFile(tempPath)
	if err != nil {
		return err
	}
	op.kopiaConfig = kopiaConfig

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

	storage, err := s3.New(ctx, op.kopiaConfig.Storage.Config.(*s3.Options), false)
	if err != nil {
		return err
	}

	userDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	if create {
		err := op.createRepo(ctx, storage)
		if err != nil {
			return err
		}
	}

	configPath := filepath.Join(userDir, "git-gasset", "kopia-"+op.config.GassetId+".config")

	err = repo.Connect(ctx, configPath, storage, op.password, &repo.ConnectOptions{
		ClientOptions:  op.kopiaConfig.ClientOptions,
		CachingOptions: content.CachingOptions{},
	})
	if err != nil {
		return err
	}

	if create {

	}
	return nil
}

func (op *initOptions) createRepo(ctx context.Context, storage blob.Storage) error {
	if err := op.ensureEmpty(ctx, storage); err != nil {
		return err
	}

	if err := repo.Initialize(ctx, storage, nil, op.password); err != nil {
		return err
	}

	gassetId := util.GenerateRandomString(8, getRandIntn)
	op.config.GassetId = gassetId
	return util.UpdateGassetId(op.workingDirectory, gassetId)
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

func (op *initOptions) createPolicy() {

}

func getRandIntn(n int) int {
	return rand.Intn(n)
}
