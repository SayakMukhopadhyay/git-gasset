package cmd

import (
	"testing"
)

func Test_initOptions_loadKopiaConfig(t *testing.T) {
	tests := []struct {
		name    string
		fields  *initOptions
		wantErr bool
	}{
		{
			name:    "Attempt from deep inside the git repository which has a .git file",
			fields:  &initOptions{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op := &initOptions{
				workingDirectory: tt.fields.workingDirectory,
				config:           tt.fields.config,
				kopiaConfig:      tt.fields.kopiaConfig,
				password:         tt.fields.password,
			}
			if err := op.loadKopiaConfig(); (err != nil) != tt.wantErr {
				t.Errorf("loadKopiaConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
