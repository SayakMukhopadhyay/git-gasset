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
	"github.com/kopia/kopia/snapshot/policy"
	"github.com/spf13/cobra"
	"log"
	"math/rand"
	"os"
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

func InitRun(cmd *cobra.Command, _ []string) error {
	log.Println("init called")

	options := util.Options{
		GassetIdLength:   8,
		OsGetwd:          os.Getwd,
		OsTempDir:        os.TempDir,
		OsUserConfigDir:  os.UserConfigDir,
		RandIntn:         rand.Intn,
		S3New:            s3.New,
		RepoConnect:      repo.Connect,
		RepoInitialize:   repo.Initialize,
		RepoOpen:         repo.Open,
		RepoWriteSession: repo.WriteSession,
		PolicySetPolicy:  policy.SetPolicy,
	}

	if err := options.InitWorkingDirectory(); err != nil {
		return err
	}

	if err := options.ReloadKopiaConfig(); err != nil {
		return err
	}

	doCreate, err := cmd.Flags().GetBool("create")
	if err != nil {
		return err
	}

	return connect(&options, doCreate)
}

func connect(op *util.Options, create bool) error {
	ctx := context.Background()

	storage, err := op.S3New(ctx, op.KopiaConfig.Storage.Config.(*s3.Options), false)
	if err != nil {
		return err
	}
	op.Storage = storage

	if create {
		if err := createRepo(ctx, op); err != nil {
			return err
		}
	}

	if err := connectRepo(ctx, op); err != nil {
		return err
	}
	return nil
}

func connectRepo(ctx context.Context, op *util.Options) error {
	kopiaUserConfigPath, err := op.GetKopiaUserConfigPath()
	if err != nil {
		return err
	}
	return op.RepoConnect(ctx, kopiaUserConfigPath, op.Storage, op.Password, &repo.ConnectOptions{
		ClientOptions:  op.KopiaConfig.ClientOptions,
		CachingOptions: content.CachingOptions{},
	})
}

func createRepo(ctx context.Context, op *util.Options) error {
	if err := ensureEmpty(ctx, op.Storage); err != nil {
		return err
	}

	if err := op.RepoInitialize(ctx, op.Storage, nil, op.Password); err != nil {
		return err
	}

	// Set a random id as gasset id once the repo is initialized
	op.Config.GassetId = util.GenerateRandomString(op.GassetIdLength, op.RandIntn)

	if err := connectRepo(ctx, op); err != nil {
		return err
	}

	if err := initPolicy(ctx, op); err != nil {
		return err
	}

	return util.UpdateGassetId(op.WorkingDirectory, op.Config.GassetId)
}

// mostly from github.com/kopia/kopia/cli.commandRepositoryCreate.ensureEmpty
func ensureEmpty(ctx context.Context, storage blob.Storage) error {
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

func initPolicy(ctx context.Context, op *util.Options) error {
	kopiaUserConfigPath, err := op.GetKopiaUserConfigPath()
	if err != nil {
		return err
	}
	rep, err := op.RepoOpen(ctx, kopiaUserConfigPath, op.Password, &repo.Options{})
	if err != nil {
		return err
	}
	if rep != nil {
		defer rep.Close(ctx)
	}
	return op.RepoWriteSession(ctx, rep, repo.WriteSessionOptions{
		Purpose: "Initialize repository with default policy",
	}, func(ctx context.Context, writer repo.RepositoryWriter) error {
		// Not needed once https://github.com/kopia/kopia/issues/3556 is closed and released
		newOptionalInt := func(b policy.OptionalInt) *policy.OptionalInt {
			return &b
		}

		defaultPolicy := &policy.Policy{
			RetentionPolicy: policy.RetentionPolicy{
				KeepLatest:               newOptionalInt(0),
				KeepHourly:               newOptionalInt(0),
				KeepDaily:                newOptionalInt(0),
				KeepWeekly:               newOptionalInt(0),
				KeepMonthly:              newOptionalInt(0),
				KeepAnnual:               newOptionalInt(0),
				IgnoreIdenticalSnapshots: policy.NewOptionalBool(false),
			},
			FilesPolicy:         policy.DefaultPolicy.FilesPolicy,
			ErrorHandlingPolicy: policy.DefaultPolicy.ErrorHandlingPolicy,
			SchedulingPolicy:    policy.DefaultPolicy.SchedulingPolicy,
			CompressionPolicy:   policy.DefaultPolicy.CompressionPolicy,
			Actions:             policy.DefaultPolicy.Actions,
			LoggingPolicy:       policy.DefaultPolicy.LoggingPolicy,
			UploadPolicy:        policy.DefaultPolicy.UploadPolicy,
		}

		return op.PolicySetPolicy(ctx, writer, policy.GlobalPolicySourceInfo, defaultPolicy)
	})
}
