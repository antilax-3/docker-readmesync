// Package config loads the readmesync.json file from the /config volume.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
)

// Config is the readmesync.json file.
type Config struct {
	// DockerHubUsername is the Docker Hub account the descriptions are written as.
	DockerHubUsername string `json:"dockerhub_username"`
	// DockerHubPassword is the account's password or a personal access token with read and write scope. Docker Hub
	// exchanges either for a bearer token; a token is needed once two-factor authentication or SSO is enforced.
	DockerHubPassword string `json:"dockerhub_password"`
	// Port is the port the API listens on.
	Port int `json:"port"`
}

// Default is the config written when none exists, and the source of any optional field a config leaves out.
var Default = Config{
	DockerHubUsername: "user",
	DockerHubPassword: "password",
	Port:              80,
}

// ErrDefaultWritten reports that no config existed and Default was written in its place, which has to be filled in
// before the service can do anything.
var ErrDefaultWritten = errors.New("copied default config")

var requiredFields = []string{"dockerhub_username", "dockerhub_password"}

// Load reads the config at path. When there is none, it writes Default there, readable by its owner alone since it
// will hold the Docker Hub credentials, and returns ErrDefaultWritten.
func Load(path string) (Config, error) {
	data, err := os.ReadFile(path) //nolint:gosec // path is the fixed config location, never request input.
	if errors.Is(err, fs.ErrNotExist) {
		return Config{}, writeDefault(path)
	}

	if err != nil {
		return Config{}, fmt.Errorf("unable to read configuration file %s: %w", path, err)
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return Config{}, fmt.Errorf("unable to parse configuration file %s, please check JSON validity: %w", path, err)
	}

	for _, field := range requiredFields {
		if _, ok := fields[field]; !ok {
			return Config{}, fmt.Errorf("missing required field '%s' from %s", field, path)
		}
	}

	config := Default
	if err := json.Unmarshal(data, &config); err != nil {
		return Config{}, fmt.Errorf("unable to parse configuration file %s, please check JSON validity: %w", path, err)
	}

	return config, nil
}

func writeDefault(path string) error {
	data, err := json.MarshalIndent(Default, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
		return fmt.Errorf("unable to write default configuration file %s: %w", path, err)
	}

	return fmt.Errorf("%w to %s", ErrDefaultWritten, path)
}
