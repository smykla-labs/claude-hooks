// Package provider provides multi-source configuration loading with precedence.
package provider

import (
	"errors"
	"fmt"

	"github.com/smykla-labs/klaudiush/internal/config"
	pkgconfig "github.com/smykla-labs/klaudiush/pkg/config"
)

// ErrNoConfig is returned when no configuration is available from any source.
var ErrNoConfig = errors.New("no configuration available")

// Source represents a configuration source (file, env, flags).
type Source interface {
	// Name returns the source name for debugging/logging.
	Name() string

	// Load loads configuration from this source.
	// Returns ErrNoConfig if no configuration is available.
	Load() (*pkgconfig.Config, error)

	// IsAvailable checks if this source has configuration available.
	IsAvailable() bool
}

// Provider loads configuration from multiple sources with precedence.
// Precedence order (highest to lowest):
// 1. CLI Flags
// 2. Environment Variables
// 3. Project Config
// 4. Global Config
// 5. Defaults
type Provider struct {
	// sources is the list of configuration sources in precedence order.
	sources []Source

	// merger merges configurations.
	merger *config.Merger

	// validator validates configurations.
	validator *config.Validator

	// cache stores the loaded configuration.
	cache *Cache
}

// NewProvider creates a new Provider with the given sources.
// Sources should be provided in precedence order (highest priority first).
func NewProvider(sources ...Source) *Provider {
	return &Provider{
		sources:   sources,
		merger:    config.NewMerger(),
		validator: config.NewValidator(),
		cache:     NewCache(),
	}
}

// Load loads and merges configuration from all sources.
// Returns a fully merged and validated configuration.
// Caches the result for subsequent calls.
func (p *Provider) Load() (*pkgconfig.Config, error) {
	// Check cache first
	if cfg := p.cache.Get(); cfg != nil {
		return cfg, nil
	}

	// Start with defaults
	configs := []*pkgconfig.Config{config.DefaultConfig()}

	// Load from each source in reverse order (lowest priority first)
	// This ensures higher priority sources override lower priority sources
	for i := len(p.sources) - 1; i >= 0; i-- {
		source := p.sources[i]

		cfg, err := source.Load()
		if err != nil {
			if errors.Is(err, ErrNoConfig) || errors.Is(err, config.ErrConfigNotFound) {
				// No config from this source, skip it
				continue
			}

			return nil, fmt.Errorf("failed to load config from %s: %w", source.Name(), err)
		}

		configs = append(configs, cfg)
	}

	// Merge all configurations
	merged := p.merger.Merge(configs...)

	// Validate the merged configuration
	if err := p.validator.Validate(merged); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Cache the result
	p.cache.Set(merged)

	return merged, nil
}

// Reload clears the cache and loads configuration again.
// Useful for testing or when configuration files change.
func (p *Provider) Reload() (*pkgconfig.Config, error) {
	p.cache.Clear()

	return p.Load()
}

// Sources returns the list of sources in precedence order.
func (p *Provider) Sources() []Source {
	return p.sources
}

// NewDefaultProvider creates a Provider with standard sources.
// Sources in precedence order:
// 1. Flags (highest)
// 2. Environment
// 3. Project Config
// 4. Global Config
// 5. Defaults (lowest, always merged in)
func NewDefaultProvider(flags map[string]any) (*Provider, error) {
	loader := config.NewLoader()

	sources := []Source{
		NewFlagSource(flags),
		NewEnvSource(),
		NewProjectFileSource(loader),
		NewGlobalFileSource(loader),
	}

	return NewProvider(sources...), nil
}
