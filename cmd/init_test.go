package cmd

import (
	"github.com/kopia/kopia/repo"
	"github.com/kopia/kopia/repo/blob"
	"github.com/kopia/kopia/repo/blob/s3"
	"github.com/kopia/kopia/repo/blob/throttling"
	"reflect"
	"testing"
)

func Test_getGitWorkingDirectory(t *testing.T) {
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
			if got, _ := getGitWorkingDirectory(tt.args.path); got != tt.want {
				t.Errorf("getGitWorkingDirectory() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_loadKopiaConfig(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    *repo.LocalConfig
		wantErr bool
	}{
		{
			name: "Attempt from deep inside the git repository which has a .git file",
			args: args{path: "D:\\Sayak\\Work\\Personal\\git-gasset\\mocks"},
			want: &repo.LocalConfig{
				APIServer: nil,
				Storage: &blob.ConnectionInfo{
					Type: "s3",
					Config: &s3.Options{
						BucketName:      "bucket-name",
						Prefix:          "prefix/",
						Endpoint:        "endpoint.digitaloceanspaces.com",
						DoNotUseTLS:     false,
						DoNotVerifyTLS:  false,
						RootCA:          nil,
						AccessKeyID:     "",
						SecretAccessKey: "",
						SessionToken:    "",
						Region:          "",
						Limits:          throttling.Limits{},
						PointInTime:     nil,
					},
				},
				Caching: nil,
				ClientOptions: repo.ClientOptions{
					Hostname:                "host-pc",
					Username:                "user",
					ReadOnly:                false,
					PermissiveCacheLoading:  false,
					Description:             "prefix",
					EnableActions:           false,
					FormatBlobCacheDuration: 900000000000,
					Throttling:              nil,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := loadKopiaConfig(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadKopiaConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("loadKopiaConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}
