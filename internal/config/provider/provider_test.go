package provider_test

import (
	"errors"
	"testing"

	"github.com/smykla-labs/klaudiush/internal/config"
	"github.com/smykla-labs/klaudiush/internal/config/provider"
	pkgconfig "github.com/smykla-labs/klaudiush/pkg/config"
)

var errTest = errors.New("test error")

type mockSource struct {
	name      string
	config    *pkgconfig.Config
	err       error
	available bool
}

func (m *mockSource) Name() string {
	return m.name
}

func (m *mockSource) Load() (*pkgconfig.Config, error) {
	return m.config, m.err
}

func (m *mockSource) IsAvailable() bool {
	return m.available
}

func TestProvider(t *testing.T) {
	t.Parallel()

	t.Run("load with defaults only", func(t *testing.T) {
		t.Parallel()

		p := provider.NewProvider()

		cfg, err := p.Load()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if cfg == nil {
			t.Fatal("expected config to be non-nil")
		}

		// Should have defaults
		if cfg.Global == nil {
			t.Error("expected global config to be set from defaults")
		}
	})

	t.Run("load with source", func(t *testing.T) {
		t.Parallel()

		enabled := false
		source := &mockSource{
			name: "test",
			config: &pkgconfig.Config{
				Validators: &pkgconfig.ValidatorsConfig{
					Git: &pkgconfig.GitConfig{
						Commit: &pkgconfig.CommitValidatorConfig{
							ValidatorConfig: pkgconfig.ValidatorConfig{
								Enabled: &enabled,
							},
						},
					},
				},
			},
			available: true,
		}

		p := provider.NewProvider(source)

		cfg, err := p.Load()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Check that source config was merged
		if cfg.Validators == nil || cfg.Validators.Git == nil || cfg.Validators.Git.Commit == nil {
			t.Fatal("expected commit config to be present")
		}

		if cfg.Validators.Git.Commit.Enabled == nil || *cfg.Validators.Git.Commit.Enabled != false {
			t.Error("expected commit validator to be disabled from source")
		}
	})

	t.Run("load with source error", func(t *testing.T) {
		t.Parallel()

		source := &mockSource{
			name:      "test",
			err:       errTest,
			available: true,
		}

		p := provider.NewProvider(source)

		_, err := p.Load()
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, errTest) {
			t.Errorf("expected error to wrap test error, got %v", err)
		}
	})

	t.Run("load with no config source", func(t *testing.T) {
		t.Parallel()

		source := &mockSource{
			name:      "test",
			err:       provider.ErrNoConfig,
			available: false,
		}

		p := provider.NewProvider(source)

		cfg, err := p.Load()
		// Should still succeed with defaults
		if err != nil {
			t.Fatalf("expected no error with defaults, got %v", err)
		}

		if cfg == nil {
			t.Fatal("expected config from defaults")
		}
	})

	t.Run("reload clears cache", func(t *testing.T) {
		t.Parallel()

		p := provider.NewProvider()

		// Load once
		cfg1, err := p.Load()
		if err != nil {
			t.Fatalf("first load failed: %v", err)
		}

		// Load again (should be cached)
		cfg2, err := p.Load()
		if err != nil {
			t.Fatalf("second load failed: %v", err)
		}

		if cfg1 != cfg2 {
			t.Error("expected second load to return cached config")
		}

		// Reload (should clear cache)
		cfg3, err := p.Reload()
		if err != nil {
			t.Fatalf("reload failed: %v", err)
		}

		// New load should be different instance
		if cfg1 == cfg3 {
			t.Error("expected reload to create new config instance")
		}
	})

	t.Run("invalid config fails validation", func(t *testing.T) {
		t.Parallel()

		source := &mockSource{
			name: "test",
			config: &pkgconfig.Config{
				Validators: &pkgconfig.ValidatorsConfig{
					Git: &pkgconfig.GitConfig{
						Commit: &pkgconfig.CommitValidatorConfig{
							ValidatorConfig: pkgconfig.ValidatorConfig{
								Severity: pkgconfig.Severity("invalid"),
							},
						},
					},
				},
			},
			available: true,
		}

		p := provider.NewProvider(source)

		_, err := p.Load()
		if err == nil {
			t.Fatal("expected validation error for invalid severity")
		}

		if !errors.Is(err, config.ErrInvalidSeverity) {
			t.Errorf("expected ErrInvalidSeverity, got %v", err)
		}
	})
}

func TestNewDefaultProvider(t *testing.T) {
	t.Parallel()

	t.Run("creates provider with standard sources", func(t *testing.T) {
		t.Parallel()

		p, err := provider.NewDefaultProvider(nil)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if p == nil {
			t.Fatal("expected provider to be non-nil")
		}

		sources := p.Sources()
		if len(sources) != 4 {
			t.Errorf("expected 4 sources, got %d", len(sources))
		}

		// Verify source order (highest priority first)
		expectedNames := []string{
			"CLI flags",
			"environment variables",
			"project config file",
			"global config file",
		}
		for i, source := range sources {
			if source.Name() != expectedNames[i] {
				t.Errorf("source %d: expected name %q, got %q", i, expectedNames[i], source.Name())
			}
		}
	})

	t.Run("loads configuration", func(t *testing.T) {
		t.Parallel()

		p, err := provider.NewDefaultProvider(nil)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		cfg, err := p.Load()
		if err != nil {
			t.Fatalf("load failed: %v", err)
		}

		if cfg == nil {
			t.Fatal("expected config to be non-nil")
		}

		// Should have defaults
		if cfg.Global == nil {
			t.Error("expected global config from defaults")
		}

		if cfg.Validators == nil {
			t.Error("expected validators config from defaults")
		}
	})
}
