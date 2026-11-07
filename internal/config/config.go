package config

import (
	"fmt"
	"time"
)

// AppConfig represents parsed CLI configuration.
type AppConfig struct {
	Mode          string // "normal", "sexy", "halo", "lizard", "custom"
	CustomPath    string
	Sensitivity   float64 // 0.01 to 1.0 (default 0.25)
	CooldownMs    int     // Cooldown in ms (default 750)
	VolumeScaling bool
	SpeedRatio    float64 // Playback speed multiplier (default 1.0)
	StdioMode     bool
}

// DefaultConfig provides recommended defaults.
func DefaultConfig() AppConfig {
	return AppConfig{
		Mode:          "normal",
		Sensitivity:   0.25,
		CooldownMs:    750,
		VolumeScaling: false,
		SpeedRatio:    1.0,
		StdioMode:     false,
	}
}

// Validate checks configuration values for safety and mutually exclusive modes.
func (c *AppConfig) Validate() error {
	if c.Sensitivity < 0.01 || c.Sensitivity > 1.0 {
		return fmt.Errorf("--sensitivity must be between 0.01 and 1.0 (got %.4f)", c.Sensitivity)
	}
	if c.CooldownMs <= 0 {
		return fmt.Errorf("--cooldown must be positive (got %d)", c.CooldownMs)
	}
	if c.SpeedRatio <= 0 {
		return fmt.Errorf("--speed must be positive (got %.2f)", c.SpeedRatio)
	}
	if c.Mode == "custom" && c.CustomPath == "" {
		return fmt.Errorf("--custom directory path must be provided")
	}
	return nil
}

// CooldownDuration returns the cooldown as a time.Duration.
func (c *AppConfig) CooldownDuration() time.Duration {
	return time.Duration(c.CooldownMs) * time.Millisecond
}
