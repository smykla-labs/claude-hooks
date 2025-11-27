// Package config provides configuration schema types for klaudiush validators.
package config

// ShellConfig groups all shell-related validator configurations.
type ShellConfig struct {
	// Backtick validator configuration
	Backtick *BacktickValidatorConfig `json:"backtick,omitempty" koanf:"backtick" toml:"backtick"`
}

// BacktickValidatorConfig configures the backtick validator.
type BacktickValidatorConfig struct {
	ValidatorConfig
}
