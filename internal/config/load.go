package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// AirHome returns AIR's data directory (~/.air), creating it if it doesn't
// exist. Checkpoints, logs, and the global config file all live under it.
func AirHome() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("config: resolve home directory: %w", err)
	}
	dir := filepath.Join(home, ".air")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("config: create %s: %w", dir, err)
	}
	return dir, nil
}

// Load builds AIR's configuration by layering, in order: built-in defaults,
// then ~/.air/config.yaml if present, then ./air.yaml if present in the
// current directory. Later layers override earlier ones.
func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")

	d := Default()
	v.SetDefault("providers", d.Providers)
	v.SetDefault("retry_failed", d.RetryFailed)
	v.SetDefault("checkpoint_interval", d.CheckpointInterval)
	v.SetDefault("summary_model", d.SummaryModel)
	v.SetDefault("parallel_review", d.ParallelReview)

	home, err := AirHome()
	if err != nil {
		return nil, err
	}
	if err := mergeIfExists(v, filepath.Join(home, "config.yaml"), false); err != nil {
		return nil, err
	}
	if err := mergeIfExists(v, "air.yaml", true); err != nil {
		return nil, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("config: unmarshal: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func mergeIfExists(v *viper.Viper, path string, merge bool) error {
	if _, err := os.Stat(path); err != nil {
		return nil
	}
	v.SetConfigFile(path)
	var readErr error
	if merge {
		readErr = v.MergeInConfig()
	} else {
		readErr = v.ReadInConfig()
	}
	if readErr != nil {
		return fmt.Errorf("config: read %s: %w", path, readErr)
	}
	return nil
}

// Validate checks the config for values the router/agent layer can't
// recover from at runtime.
func (c *Config) Validate() error {
	if len(c.Providers) == 0 {
		return fmt.Errorf("config: providers list is empty")
	}
	if c.CheckpointInterval <= 0 {
		return fmt.Errorf("config: checkpoint_interval must be positive, got %s", c.CheckpointInterval)
	}
	return nil
}
