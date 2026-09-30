package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, contents string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "readmesync.json")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}

	return path
}

func TestLoadWritesTheDefaultWhenMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "readmesync.json")

	if _, err := Load(path); !errors.Is(err, ErrDefaultWritten) {
		t.Fatalf("got %v, want ErrDefaultWritten", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	if mode := info.Mode().Perm(); mode != 0o600 {
		t.Fatalf("default config mode is %o, want 600", mode)
	}

	config, err := Load(path)
	if err != nil || config != Default {
		t.Fatalf("reloaded %+v, %v, want the default config", config, err)
	}
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name     string
		contents string
		want     Config
		err      string
	}{
		{
			name:     "full config",
			contents: `{"dockerhub_username":"nightah","dockerhub_password":"dckr_pat_x","port":8080}`,
			want:     Config{DockerHubUsername: "nightah", DockerHubPassword: "dckr_pat_x", Port: 8080},
		},
		{
			name:     "port defaults to 80",
			contents: `{"dockerhub_username":"nightah","dockerhub_password":"hunter2"}`,
			want:     Config{DockerHubUsername: "nightah", DockerHubPassword: "hunter2", Port: 80},
		},
		{
			name:     "password is required",
			contents: `{"dockerhub_username":"nightah"}`,
			err:      "missing required field 'dockerhub_password'",
		},
		{
			name:     "invalid json",
			contents: `{`,
			err:      "please check JSON validity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := Load(write(t, tt.contents))

			if tt.err != "" {
				if err == nil || !strings.Contains(err.Error(), tt.err) {
					t.Fatalf("got %v, want an error containing %q", err, tt.err)
				}

				return
			}

			if err != nil || config != tt.want {
				t.Fatalf("got %+v, %v, want %+v", config, err, tt.want)
			}
		})
	}
}
