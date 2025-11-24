// Package provider provides multi-source configuration loading with precedence.
package provider

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/smykla-labs/klaudiush/internal/config"
	pkgconfig "github.com/smykla-labs/klaudiush/pkg/config"
)

// ErrUnknownValidator is returned when an unknown validator name is provided.
var ErrUnknownValidator = errors.New("unknown validator")

// GlobalFileSource loads configuration from the global config file.
type GlobalFileSource struct {
	loader *config.Loader
}

// NewGlobalFileSource creates a new GlobalFileSource.
func NewGlobalFileSource(loader *config.Loader) *GlobalFileSource {
	return &GlobalFileSource{loader: loader}
}

// Name returns the source name.
func (*GlobalFileSource) Name() string {
	return "global config file"
}

// Load loads configuration from the global config file.
func (s *GlobalFileSource) Load() (*pkgconfig.Config, error) {
	cfg, err := s.loader.LoadGlobal()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// IsAvailable checks if the global config file exists.
func (s *GlobalFileSource) IsAvailable() bool {
	return s.loader.HasGlobalConfig()
}

// ProjectFileSource loads configuration from the project config file.
type ProjectFileSource struct {
	loader *config.Loader
}

// NewProjectFileSource creates a new ProjectFileSource.
func NewProjectFileSource(loader *config.Loader) *ProjectFileSource {
	return &ProjectFileSource{loader: loader}
}

// Name returns the source name.
func (*ProjectFileSource) Name() string {
	return "project config file"
}

// Load loads configuration from the project config file.
func (s *ProjectFileSource) Load() (*pkgconfig.Config, error) {
	cfg, err := s.loader.LoadProject()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// IsAvailable checks if a project config file exists.
func (s *ProjectFileSource) IsAvailable() bool {
	return s.loader.HasProjectConfig()
}

// EnvSource loads configuration from environment variables.
// Environment variables follow the pattern: KLAUDIUSH_SECTION_SUBSECTION_FIELD
// Examples:
// - KLAUDIUSH_USE_SDK_GIT=true
// - KLAUDIUSH_VALIDATORS_GIT_COMMIT_ENABLED=false
// - KLAUDIUSH_VALIDATORS_GIT_COMMIT_SEVERITY=warning
type EnvSource struct{}

// NewEnvSource creates a new EnvSource.
func NewEnvSource() *EnvSource {
	return &EnvSource{}
}

// Name returns the source name.
func (*EnvSource) Name() string {
	return "environment variables"
}

// Load loads configuration from environment variables.
func (*EnvSource) Load() (*pkgconfig.Config, error) {
	cfg := &pkgconfig.Config{}
	hasAny := false

	// Global: KLAUDIUSH_USE_SDK_GIT
	if val := os.Getenv("KLAUDIUSH_USE_SDK_GIT"); val != "" {
		if cfg.Global == nil {
			cfg.Global = &pkgconfig.GlobalConfig{}
		}

		useSDK, err := strconv.ParseBool(val)
		if err != nil {
			return nil, fmt.Errorf("invalid value for KLAUDIUSH_USE_SDK_GIT: %w", err)
		}

		cfg.Global.UseSDKGit = &useSDK
		hasAny = true
	}

	// Global: KLAUDIUSH_DEFAULT_TIMEOUT
	if val := os.Getenv("KLAUDIUSH_DEFAULT_TIMEOUT"); val != "" {
		if cfg.Global == nil {
			cfg.Global = &pkgconfig.GlobalConfig{}
		}

		duration, err := time.ParseDuration(val)
		if err != nil {
			return nil, fmt.Errorf("invalid value for KLAUDIUSH_DEFAULT_TIMEOUT: %w", err)
		}

		cfg.Global.DefaultTimeout = pkgconfig.Duration(duration)
		hasAny = true
	}

	// Load validator-specific environment variables
	if err := loadValidatorEnvVars(cfg); err != nil {
		return nil, err
	}

	// Check if any validators were configured
	if cfg.Validators != nil {
		hasAny = true
	}

	if !hasAny {
		return nil, ErrNoConfig
	}

	return cfg, nil
}

// IsAvailable checks if any environment variables are set.
func (*EnvSource) IsAvailable() bool {
	// Check for common env vars
	envVars := []string{
		"KLAUDIUSH_USE_SDK_GIT",
		"KLAUDIUSH_DEFAULT_TIMEOUT",
		"KLAUDIUSH_VALIDATORS_GIT_COMMIT_ENABLED",
		"KLAUDIUSH_VALIDATORS_GIT_COMMIT_SEVERITY",
	}

	for _, key := range envVars {
		if os.Getenv(key) != "" {
			return true
		}
	}

	// Check for any KLAUDIUSH_ prefixed variables
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "KLAUDIUSH_") {
			return true
		}
	}

	return false
}

// loadValidatorEnvVars loads validator-specific environment variables.
func loadValidatorEnvVars(cfg *pkgconfig.Config) error {
	// Git validators
	if err := loadGitValidatorEnvVars(cfg); err != nil {
		return err
	}

	// File validators
	if err := loadFileValidatorEnvVars(cfg); err != nil {
		return err
	}

	// Notification validators
	return loadNotificationValidatorEnvVars(cfg)
}

// loadGitValidatorEnvVars loads git validator environment variables.
func loadGitValidatorEnvVars(cfg *pkgconfig.Config) error {
	// Commit validator
	if err := loadValidatorBaseEnvVars(
		"KLAUDIUSH_VALIDATORS_GIT_COMMIT",
		func() *pkgconfig.ValidatorConfig {
			if cfg.Validators == nil {
				cfg.Validators = &pkgconfig.ValidatorsConfig{}
			}

			if cfg.Validators.Git == nil {
				cfg.Validators.Git = &pkgconfig.GitConfig{}
			}

			if cfg.Validators.Git.Commit == nil {
				cfg.Validators.Git.Commit = &pkgconfig.CommitValidatorConfig{}
			}

			return &cfg.Validators.Git.Commit.ValidatorConfig
		},
	); err != nil {
		return err
	}

	// Push validator
	if err := loadValidatorBaseEnvVars(
		"KLAUDIUSH_VALIDATORS_GIT_PUSH",
		func() *pkgconfig.ValidatorConfig {
			if cfg.Validators == nil {
				cfg.Validators = &pkgconfig.ValidatorsConfig{}
			}

			if cfg.Validators.Git == nil {
				cfg.Validators.Git = &pkgconfig.GitConfig{}
			}

			if cfg.Validators.Git.Push == nil {
				cfg.Validators.Git.Push = &pkgconfig.PushValidatorConfig{}
			}

			return &cfg.Validators.Git.Push.ValidatorConfig
		},
	); err != nil {
		return err
	}

	// Add validator
	return loadValidatorBaseEnvVars(
		"KLAUDIUSH_VALIDATORS_GIT_ADD",
		func() *pkgconfig.ValidatorConfig {
			if cfg.Validators == nil {
				cfg.Validators = &pkgconfig.ValidatorsConfig{}
			}

			if cfg.Validators.Git == nil {
				cfg.Validators.Git = &pkgconfig.GitConfig{}
			}

			if cfg.Validators.Git.Add == nil {
				cfg.Validators.Git.Add = &pkgconfig.AddValidatorConfig{}
			}

			return &cfg.Validators.Git.Add.ValidatorConfig
		},
	)
}

// loadFileValidatorEnvVars loads file validator environment variables.
func loadFileValidatorEnvVars(cfg *pkgconfig.Config) error {
	// Markdown validator
	if err := loadValidatorBaseEnvVars(
		"KLAUDIUSH_VALIDATORS_FILE_MARKDOWN",
		func() *pkgconfig.ValidatorConfig {
			if cfg.Validators == nil {
				cfg.Validators = &pkgconfig.ValidatorsConfig{}
			}

			if cfg.Validators.File == nil {
				cfg.Validators.File = &pkgconfig.FileConfig{}
			}

			if cfg.Validators.File.Markdown == nil {
				cfg.Validators.File.Markdown = &pkgconfig.MarkdownValidatorConfig{}
			}

			return &cfg.Validators.File.Markdown.ValidatorConfig
		},
	); err != nil {
		return err
	}

	// ShellScript validator
	return loadValidatorBaseEnvVars(
		"KLAUDIUSH_VALIDATORS_FILE_SHELLSCRIPT",
		func() *pkgconfig.ValidatorConfig {
			if cfg.Validators == nil {
				cfg.Validators = &pkgconfig.ValidatorsConfig{}
			}

			if cfg.Validators.File == nil {
				cfg.Validators.File = &pkgconfig.FileConfig{}
			}

			if cfg.Validators.File.ShellScript == nil {
				cfg.Validators.File.ShellScript = &pkgconfig.ShellScriptValidatorConfig{}
			}

			return &cfg.Validators.File.ShellScript.ValidatorConfig
		},
	)
}

// loadNotificationValidatorEnvVars loads notification validator environment variables.
func loadNotificationValidatorEnvVars(cfg *pkgconfig.Config) error {
	// Bell validator
	return loadValidatorBaseEnvVars(
		"KLAUDIUSH_VALIDATORS_NOTIFICATION_BELL",
		func() *pkgconfig.ValidatorConfig {
			if cfg.Validators == nil {
				cfg.Validators = &pkgconfig.ValidatorsConfig{}
			}

			if cfg.Validators.Notification == nil {
				cfg.Validators.Notification = &pkgconfig.NotificationConfig{}
			}

			if cfg.Validators.Notification.Bell == nil {
				cfg.Validators.Notification.Bell = &pkgconfig.BellValidatorConfig{}
			}

			return &cfg.Validators.Notification.Bell.ValidatorConfig
		},
	)
}

// loadValidatorBaseEnvVars loads base validator environment variables (enabled, severity).
func loadValidatorBaseEnvVars(prefix string, getValidator func() *pkgconfig.ValidatorConfig) error {
	// Enabled
	if val := os.Getenv(prefix + "_ENABLED"); val != "" {
		enabled, err := strconv.ParseBool(val)
		if err != nil {
			return fmt.Errorf("invalid value for %s_ENABLED: %w", prefix, err)
		}

		validator := getValidator()
		validator.Enabled = &enabled
	}

	// Severity
	if val := os.Getenv(prefix + "_SEVERITY"); val != "" {
		validator := getValidator()
		validator.Severity = pkgconfig.Severity(val)
	}

	return nil
}

// FlagSource loads configuration from CLI flags.
type FlagSource struct {
	flags map[string]any
}

// NewFlagSource creates a new FlagSource.
func NewFlagSource(flags map[string]any) *FlagSource {
	return &FlagSource{flags: flags}
}

// Name returns the source name.
func (*FlagSource) Name() string {
	return "CLI flags"
}

// Load loads configuration from CLI flags.
func (s *FlagSource) Load() (*pkgconfig.Config, error) {
	if len(s.flags) == 0 {
		return nil, ErrNoConfig
	}

	cfg := &pkgconfig.Config{}

	// Process flags
	for key, value := range s.flags {
		if err := s.applyFlag(cfg, key, value); err != nil {
			return nil, fmt.Errorf("failed to apply flag %s: %w", key, err)
		}
	}

	return cfg, nil
}

// IsAvailable checks if any flags are set.
func (s *FlagSource) IsAvailable() bool {
	return len(s.flags) > 0
}

// applyFlag applies a single flag to the configuration.
//
//nolint:gocognit // Flag parsing is inherently complex
func (*FlagSource) applyFlag(cfg *pkgconfig.Config, key string, value any) error {
	// Example flag mappings:
	// --use-sdk-git=true
	// --disable=commit,markdown
	// --commit-title-max=60
	// --timeout=15s
	switch key {
	case "use-sdk-git":
		if cfg.Global == nil {
			cfg.Global = &pkgconfig.GlobalConfig{}
		}

		if boolVal, ok := value.(bool); ok {
			cfg.Global.UseSDKGit = &boolVal
		}

	case "timeout":
		if cfg.Global == nil {
			cfg.Global = &pkgconfig.GlobalConfig{}
		}

		if strVal, ok := value.(string); ok {
			duration, err := time.ParseDuration(strVal)
			if err != nil {
				return fmt.Errorf("invalid timeout duration: %w", err)
			}

			cfg.Global.DefaultTimeout = pkgconfig.Duration(duration)
		}

	case "disable":
		// Format: --disable=commit,markdown,push
		if strVal, ok := value.(string); ok {
			for v := range strings.SplitSeq(strVal, ",") {
				v = strings.TrimSpace(v)
				if err := disableValidator(cfg, v); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// disableValidator disables a specific validator by name.
//
//nolint:gocognit // Validator mapping is inherently complex
func disableValidator(cfg *pkgconfig.Config, name string) error {
	disabled := false

	if cfg.Validators == nil {
		cfg.Validators = &pkgconfig.ValidatorsConfig{}
	}

	switch name {
	case "commit":
		if cfg.Validators.Git == nil {
			cfg.Validators.Git = &pkgconfig.GitConfig{}
		}

		if cfg.Validators.Git.Commit == nil {
			cfg.Validators.Git.Commit = &pkgconfig.CommitValidatorConfig{}
		}

		cfg.Validators.Git.Commit.Enabled = &disabled

	case "push":
		if cfg.Validators.Git == nil {
			cfg.Validators.Git = &pkgconfig.GitConfig{}
		}

		if cfg.Validators.Git.Push == nil {
			cfg.Validators.Git.Push = &pkgconfig.PushValidatorConfig{}
		}

		cfg.Validators.Git.Push.Enabled = &disabled

	case "add":
		if cfg.Validators.Git == nil {
			cfg.Validators.Git = &pkgconfig.GitConfig{}
		}

		if cfg.Validators.Git.Add == nil {
			cfg.Validators.Git.Add = &pkgconfig.AddValidatorConfig{}
		}

		cfg.Validators.Git.Add.Enabled = &disabled

	case "markdown":
		if cfg.Validators.File == nil {
			cfg.Validators.File = &pkgconfig.FileConfig{}
		}

		if cfg.Validators.File.Markdown == nil {
			cfg.Validators.File.Markdown = &pkgconfig.MarkdownValidatorConfig{}
		}

		cfg.Validators.File.Markdown.Enabled = &disabled

	case "shellscript":
		if cfg.Validators.File == nil {
			cfg.Validators.File = &pkgconfig.FileConfig{}
		}

		if cfg.Validators.File.ShellScript == nil {
			cfg.Validators.File.ShellScript = &pkgconfig.ShellScriptValidatorConfig{}
		}

		cfg.Validators.File.ShellScript.Enabled = &disabled

	case "terraform":
		if cfg.Validators.File == nil {
			cfg.Validators.File = &pkgconfig.FileConfig{}
		}

		if cfg.Validators.File.Terraform == nil {
			cfg.Validators.File.Terraform = &pkgconfig.TerraformValidatorConfig{}
		}

		cfg.Validators.File.Terraform.Enabled = &disabled

	default:
		return fmt.Errorf("%w: %s", ErrUnknownValidator, name)
	}

	return nil
}
