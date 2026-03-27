package monitor

import (
	"encoding/json"
	"errors"
	"os"
	"time"
)

type fileConfig struct {
	PingTargets    []string `json:"ping_targets"`
	DNSTargets     []string `json:"dns_targets"`
	PingIntervalS  int      `json:"ping_interval_s"`
	SpeedIntervalM int      `json:"speed_interval_m"`
	PingCount      int      `json:"ping_count"`
	DownloadURL    string   `json:"download_url"`
	UploadURL      string   `json:"upload_url"`
}

// LoadConfigFile reads a JSON config file and returns a Config.
// Fields that are missing or zero-valued fall back to DefaultConfig.
// Returns (default config, nil) if the file does not exist.
func LoadConfigFile(path string) (Config, error) {
	def := DefaultConfig()

	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return def, nil
		}
		return def, err
	}
	defer f.Close()

	var fc fileConfig
	if err := json.NewDecoder(f).Decode(&fc); err != nil {
		return def, err
	}

	cfg := def
	if len(fc.PingTargets) > 0 {
		cfg.PingTargets = fc.PingTargets
	}
	if len(fc.DNSTargets) > 0 {
		cfg.DNSTargets = fc.DNSTargets
	}
	if fc.PingIntervalS > 0 {
		cfg.PingInterval = time.Duration(fc.PingIntervalS) * time.Second
	}
	if fc.SpeedIntervalM > 0 {
		cfg.SpeedInterval = time.Duration(fc.SpeedIntervalM) * time.Minute
	}
	if fc.PingCount > 0 {
		cfg.PingCount = fc.PingCount
	}
	if fc.DownloadURL != "" {
		cfg.DownloadURL = fc.DownloadURL
	}
	if fc.UploadURL != "" {
		cfg.UploadURL = fc.UploadURL
	}
	return cfg, nil
}
