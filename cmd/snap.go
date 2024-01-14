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
	"git-gasset/util"
	"github.com/kopia/kopia/fs"
	"github.com/kopia/kopia/fs/localfs"
	"github.com/kopia/kopia/repo"
	"github.com/kopia/kopia/repo/blob/s3"
	"github.com/kopia/kopia/snapshot"
	"github.com/kopia/kopia/snapshot/policy"
	"github.com/kopia/kopia/snapshot/snapshotfs"
	"github.com/spf13/cobra"
	"log"
	"math/rand"
	"os"
	"path/filepath"
)

// snapCmd represents the snap command
var snapCmd = &cobra.Command{
	Use:   "snap",
	Short: "Takes a snapshot of the assets",
	Long: `Takes a snapshot of the assets.

It uses the locations key in the .gasset.yaml file to determine the 
assets to be snapshotted.`,
	RunE: SnapRun,
}

func init() {
	rootCmd.AddCommand(snapCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// snapCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// snapCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func SnapRun(cmd *cobra.Command, args []string) error {
	log.Println("snap called")

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

	return createSnapshot(&options)
}

func createSnapshot(op *util.Options) error {
	ctx := context.Background()

	kopiaUserConfigPath, err := op.GetKopiaUserConfigPath()
	if err != nil {
		return err
	}

	rep, err := op.RepoOpen(ctx, kopiaUserConfigPath, op.Password, &repo.Options{})
	if err != nil {
		return err
	}
	defer rep.Close(ctx)

	return op.RepoWriteSession(ctx, rep, repo.WriteSessionOptions{
		Purpose: "Create snapshot",
	}, func(ctx context.Context, writer repo.RepositoryWriter) error {
		uploader := snapshotfs.NewUploader(writer)
		uploader.MaxUploadBytes = 0 << 20 // 2^20 or 1 MiB

		for _, dirPath := range op.Config.Dirs {
			fsEntry, err := localfs.NewEntry(dirPath)
			if err != nil {
				return err
			}
			info := snapshot.SourceInfo{
				Host:     rep.ClientOptions().Hostname,
				UserName: rep.ClientOptions().Username,
				Path:     filepath.Join(op.WorkingDirectory, dirPath),
			}

			if err := snapshotSingleSource(ctx, fsEntry, writer, uploader, info); err != nil {
				return err
			}
		}
		return nil
	})
}

// mostly from github.com/kopia/kopia/cli.commandSnapshotCreate.snapshotSingleSource
func snapshotSingleSource(ctx context.Context, fsEntry fs.Entry, rep repo.RepositoryWriter, uploader *snapshotfs.Uploader, sourceInfo snapshot.SourceInfo) error {
	previousManifests, err := findPreviousSnapshotManifest(ctx, rep, sourceInfo)
	if err != nil {
		return err
	}

	policyTree, err := policy.TreeForSource(ctx, rep, sourceInfo)
	if err != nil {
		return err
	}

	manifest, err := uploader.Upload(ctx, fsEntry, policyTree, sourceInfo, previousManifests...)
	if err != nil {
		return err
	}

	//Todo: Add a description to the manifest
	manifest.Description = ""
	manifest.Tags = nil

	// Update pinning not required
	// startTimeOverride and endTimeOverride not required

	ignoreIdenticalSnapshot := policyTree.EffectivePolicy().RetentionPolicy.IgnoreIdenticalSnapshots.OrDefault(false)
	if ignoreIdenticalSnapshot && len(previousManifests) > 0 {
		if previousManifests[0].RootObjectID() == manifest.RootObjectID() {
			log.Println("Not saving snapshot because no files have been changed since previous snapshot")
			return nil
		}
	}

	if _, err = snapshot.SaveSnapshot(ctx, rep, manifest); err != nil {
		return err
	}

	if _, err = policy.ApplyRetentionPolicy(ctx, rep, sourceInfo, false); err != nil {
		return err
	}

	return nil
}

// mostly from github.com/kopia/kopia/cli.findPreviousSnapshotManifest
func findPreviousSnapshotManifest(ctx context.Context, rep repo.Repository, sourceInfo snapshot.SourceInfo) ([]*snapshot.Manifest, error) {
	manifests, err := snapshot.ListSnapshots(ctx, rep, sourceInfo)
	if err != nil {
		return nil, err
	}

	var previousComplete *snapshot.Manifest

	var previousCompleteStartTime fs.UTCTimestamp

	var result []*snapshot.Manifest

	for _, manifest := range manifests {
		if manifest.IncompleteReason == "" && (previousComplete == nil || manifest.StartTime.After(previousComplete.StartTime)) {
			previousComplete = manifest
			previousCompleteStartTime = manifest.StartTime
		}
	}

	if previousComplete != nil {
		result = append(result, previousComplete)
	}

	for _, manifest := range manifests {
		if manifest.IncompleteReason != "" && manifest.StartTime.After(previousCompleteStartTime) {
			result = append(result, manifest)
		}
	}

	return result, nil
}
