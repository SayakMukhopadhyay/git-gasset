package util

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type ConfigSuite struct {
	suite.Suite
	Data []string
}

func (suite *ConfigSuite) SetupTest() {
	suite.Data = []string{"a", "b", "c"}
}

func TestGetGitWorkingDirectory(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name        string
		args        args
		expected    string
		expectedErr error
	}{
		{
			name:     "Attempt from deep inside the git repository",
			args:     args{path: "D:\\Sayak\\Work\\Personal\\git-gasset\\.idea\\runConfigurations"},
			expected: "D:\\Sayak\\Work\\Personal\\git-gasset",
		},
		{
			name:     "Attempt from the working directory of the git repository",
			args:     args{path: "D:\\Sayak\\Work\\Personal\\git-gasset"},
			expected: "D:\\Sayak\\Work\\Personal\\git-gasset",
		},
		{
			name:     "Attempt from deep inside the git repository which has a .git file",
			args:     args{path: "D:\\Sayak\\Work\\Personal\\git-gasset\\mocks\\deep\\deeper"},
			expected: "D:\\Sayak\\Work\\Personal\\git-gasset",
		},
		{
			name:        "Attempt from deep inside the git repository which has a .git file",
			args:        args{path: "D:\\"},
			expected:    "",
			expectedErr: errors.New("not a git repository"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetGitWorkingDirectory(tt.args.path)
			if err != nil && tt.expectedErr != nil {
				assert.ErrorAs(t, err, &tt.expectedErr)
			}
			if err == nil && tt.expectedErr == nil {
				assert.Equal(t, got, tt.expected)
			}
		})
	}
}
