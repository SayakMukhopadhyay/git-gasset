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
	"errors"
	"github.com/joho/godotenv"
	"github.com/kopia/kopia/repo"
	"github.com/kopia/kopia/repo/blob/s3"
	"github.com/spf13/cobra"
	"log"
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

func InitRun(cmd *cobra.Command, args []string) error {
	log.Println("init called")
	workingDirectory, err := getWorkingDirectory()
	if err != nil {
		return err
	}
	config, err := loadKopiaConfig(workingDirectory)
	log.Println(config)
	if err != nil {
		return err
	}

	return nil
}

func getWorkingDirectory() (string, error) {
	// Get the current working directory
	workingDirectory, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	return getGitWorkingDirectory(workingDirectory)
}

func getGitWorkingDirectory(path string) (string, error) {
	if info, err := os.Stat(filepath.Join(path, ".git")); os.IsNotExist(err) || !info.IsDir() {
		parent := filepath.Dir(path)
		if parent == path {
			return "", errors.New("not a git repository")
		}
		return getGitWorkingDirectory(parent)
	}
	return path, nil
}

func loadKopiaConfig(path string) (*repo.LocalConfig, error) {
	config, err := repo.LoadConfigFromFile(filepath.Join(path, ".kopia"))
	accessKey, secretKey, password := loadKopiaSecretsFromEnv(path)
	log.Println(accessKey, secretKey, password)
	if typedConfig, ok := config.Storage.Config.(*s3.Options); ok {
		typedConfig.AccessKeyID = accessKey
		typedConfig.SecretAccessKey = secretKey
	}
	if err != nil {
		return nil, err
	}
	return config, nil
}

func loadKopiaSecretsFromEnv(path string) (string, string, string) {
	godotenv.Load(filepath.Join(path, ".env"))
	return os.Getenv("KOPIA_ACCESS_ID"), os.Getenv("KOPIA_ACCESS_SECRET"), os.Getenv("KOPIA_PASSWORD")
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
