package tests

import (
	"testing"
	"time"

	"bonk/internal/config"
)

func TestAppConfig(t *testing.T) {
	t.Run("default config is valid", func(t *testing.T) {
		cfg := config.DefaultConfig()
		if err := cfg.Validate(); err != nil {
			t.Errorf("default config should be valid: %v", err)
		}
		if cfg.CooldownDuration() != 750*time.Millisecond {
			t.Errorf("expected 750ms cooldown, got %v", cfg.CooldownDuration())
		}
	})

	t.Run("invalid sensitivity errors", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Sensitivity = 0.001
		if err := cfg.Validate(); err == nil {
			t.Errorf("expected error for sensitivity < 0.01")
		}

		cfg.Sensitivity = 1.5
		if err := cfg.Validate(); err == nil {
			t.Errorf("expected error for sensitivity > 1.0")
		}
	})

	t.Run("invalid cooldown errors", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.CooldownMs = 0
		if err := cfg.Validate(); err == nil {
			t.Errorf("expected error for zero cooldown")
		}

		cfg.CooldownMs = -100
		if err := cfg.Validate(); err == nil {
			t.Errorf("expected error for negative cooldown")
		}
	})

	t.Run("custom mode requires path", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Mode = "custom"
		cfg.CustomPath = ""
		if err := cfg.Validate(); err == nil {
			t.Errorf("expected error for custom mode without path")
		}
	})
}
