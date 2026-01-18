package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Bankroll contains the bankroll configuration per platform.
type Bankroll struct {
	Polymarket float64 `yaml:"polymarket"`
	Kalshi     float64 `yaml:"kalshi"`
}

// Scan contains the scanning configuration.
type Scan struct {
	IntervalSeconds int `yaml:"interval_seconds"`
}

// Parameters contains the trading parameters.
type Parameters struct {
	ProbabilityThreshold   float64 `yaml:"probability_threshold"`
	VolatilitySafetyMargin float64 `yaml:"volatility_safety_margin"`
	StopLossPercent        float64 `yaml:"stop_loss_percent"`
	KellyFraction          float64 `yaml:"kelly_fraction"`
}

// Database contains the database configuration.
type Database struct {
	Path string `yaml:"path"`
}

// Config is the main configuration struct.
type Config struct {
	Bankroll   Bankroll   `yaml:"bankroll"`
	Scan       Scan       `yaml:"scan"`
	Parameters Parameters `yaml:"parameters"`
	Database   Database   `yaml:"database"`
}

// LoadConfig loads configuration from a YAML file.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	return &cfg, nil
}
