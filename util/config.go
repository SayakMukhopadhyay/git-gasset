package util

import (
	"encoding/json"
	"errors"
	"github.com/joho/godotenv"
	"github.com/kopia/kopia/repo"
	"os"
	"path/filepath"
)

type Config struct {
	Kopia    *repo.LocalConfig `json:"kopia,omitempty"`
	GassetId string            `json:"gassetId,omitempty"`
}

func GetConfig(path string) (*Config, error) {
	configBytes, err := os.ReadFile(filepath.Join(path, ".gasset"))
	if err != nil {
		return nil, err
	}

	config := Config{}

	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func UpdateGassetId(path string, gassetId string) error {
	config, err := GetConfig(path)
	if err != nil {
		return err
	}

	config.GassetId = gassetId
	return UpdateConfig(path, config)
}

func UpdateConfig(path string, config *Config) error {
	configBytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, configBytes, 644)
}

func GetTempKopiaConfigPath(config *Config) (string, error) {
	kopiaConfigBytes, err := json.MarshalIndent(config.Kopia, "", "  ")
	if err != nil {
		return "", err
	}

	tempPath := filepath.Join(os.TempDir(), "kopia.config")

	err = os.WriteFile(tempPath, kopiaConfigBytes, 0644)
	if err != nil {
		return "", err
	}

	return tempPath, nil
}

func LoadKopiaSecretsFromEnv(path string) (string, string, string, error) {
	err := godotenv.Load(filepath.Join(path, ".env"))
	if err != nil {
		return "", "", "", err
	}
	return os.Getenv("KOPIA_ACCESS_ID"), os.Getenv("KOPIA_ACCESS_SECRET"), os.Getenv("KOPIA_PASSWORD"), nil
}

func GetGitWorkingDirectory(path string) (string, error) {
	if info, err := os.Stat(filepath.Join(path, ".git")); os.IsNotExist(err) || !info.IsDir() {
		parent := filepath.Dir(path)
		if parent == path {
			return "", errors.New("not a git repository")
		}
		return GetGitWorkingDirectory(parent)
	}
	return path, nil
}
