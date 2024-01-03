package util

import (
	"testing"
)

func TestGetGitWorkingDirectory(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Attempt from deep inside the git repository",
			args: args{path: "D:\\Sayak\\Work\\Personal\\git-gasset\\.idea\\runConfigurations"},
			want: "D:\\Sayak\\Work\\Personal\\git-gasset",
		},
		{
			name: "Attempt from the working directory of the git repository",
			args: args{path: "D:\\Sayak\\Work\\Personal\\git-gasset"},
			want: "D:\\Sayak\\Work\\Personal\\git-gasset",
		},
		{
			name: "Attempt from deep inside the git repository which has a .git file",
			args: args{path: "D:\\Sayak\\Work\\Personal\\git-gasset\\mocks\\deep\\deeper"},
			want: "D:\\Sayak\\Work\\Personal\\git-gasset",
		},
		{
			name: "Attempt from deep inside the git repository which has a .git file",
			args: args{path: "D:\\"},
			want: "D:\\Sayak\\Work\\Personal\\git-gasset",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := GetGitWorkingDirectory(tt.args.path); got != tt.want {
				t.Errorf("getGitWorkingDirectory() = %v, want %v", got, tt.want)
			}
		})
	}
}
