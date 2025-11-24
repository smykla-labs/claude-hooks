package provider_test

import (
	"testing"

	"github.com/smykla-labs/klaudiush/internal/config/provider"
	pkgconfig "github.com/smykla-labs/klaudiush/pkg/config"
)

func TestCache(t *testing.T) {
	t.Parallel()

	t.Run("new cache is empty", func(t *testing.T) {
		t.Parallel()

		cache := provider.NewCache()
		if cache.Has() {
			t.Error("expected new cache to be empty")
		}

		if cfg := cache.Get(); cfg != nil {
			t.Errorf("expected nil config, got %+v", cfg)
		}
	})

	t.Run("set and get", func(t *testing.T) {
		t.Parallel()

		cache := provider.NewCache()
		cfg := &pkgconfig.Config{}

		cache.Set(cfg)

		if !cache.Has() {
			t.Error("expected cache to have config after set")
		}

		got := cache.Get()
		if got != cfg {
			t.Error("expected to get the same config that was set")
		}
	})

	t.Run("clear", func(t *testing.T) {
		t.Parallel()

		cache := provider.NewCache()
		cfg := &pkgconfig.Config{}

		cache.Set(cfg)
		cache.Clear()

		if cache.Has() {
			t.Error("expected cache to be empty after clear")
		}

		if got := cache.Get(); got != nil {
			t.Errorf("expected nil after clear, got %+v", got)
		}
	})
}
