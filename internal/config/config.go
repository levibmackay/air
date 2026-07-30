// Package config loads and validates AIR's YAML configuration via Viper.
package config

import "time"

// Config is AIR's top-level configuration, loaded from ~/.air/config.yaml
// with an optional ./air.yaml project override.
type Config struct {
	// Providers is the ordered list of provider keys to try, e.g.
	// "claude", "codex", "gemini", or "ollama:qwen3-coder:30b".
	Providers []string `mapstructure:"providers" yaml:"providers"`

	RetryFailed        bool          `mapstructure:"retry_failed" yaml:"retry_failed"`
	CheckpointInterval time.Duration `mapstructure:"checkpoint_interval" yaml:"checkpoint_interval"`
	SummaryModel       string        `mapstructure:"summary_model" yaml:"summary_model"`
	ParallelReview     bool          `mapstructure:"parallel_review" yaml:"parallel_review"`
}

// Default returns AIR's built-in default configuration, used when no config
// file is present.
func Default() *Config {
	return &Config{
		Providers:          []string{"claude"},
		RetryFailed:        true,
		CheckpointInterval: 2 * time.Minute,
		SummaryModel:       "local",
		ParallelReview:     false,
	}
}
