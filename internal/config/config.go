// Package config loads and validates the YAML configuration for left and right FND sides.
package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const defaultPollInterval = 60 * time.Second

// Target identifies an SNS provider and the account handle to poll.
type Target struct {
	Provider string
	Handle   string
}

// SideConfig represents the YAML configuration for one side of the FND display.
type SideConfig struct {
	YouTube   string `yaml:"youtube"`
	Insta     string `yaml:"insta"`
	Instagram string `yaml:"instagram"`
	Twitch    string `yaml:"twitch"`
}

// Config represents the application YAML configuration.
type Config struct {
	Interval time.Duration `yaml:"interval"`
	Left     SideConfig    `yaml:"left"`
	Right    SideConfig    `yaml:"right"`
}

// Load reads and validates a YAML configuration file from path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config %q: %w", path, err)
	}

	if _, err := cfg.Left.Resolve(); err != nil {
		return nil, fmt.Errorf("left config: %w", err)
	}
	if _, err := cfg.Right.Resolve(); err != nil {
		return nil, fmt.Errorf("right config: %w", err)
	}

	if cfg.Interval <= 0 {
		cfg.Interval = defaultPollInterval
	}

	return &cfg, nil
}

// Resolve validates that exactly one SNS target is configured and returns it.
func (s SideConfig) Resolve() (Target, error) {
	candidates := []Target{
		{Provider: "youtube", Handle: strings.TrimSpace(s.YouTube)},
		{Provider: "insta", Handle: strings.TrimSpace(s.Insta)},
		{Provider: "instagram", Handle: strings.TrimSpace(s.Instagram)},
		{Provider: "twitch", Handle: strings.TrimSpace(s.Twitch)},
	}

	selected := make([]Target, 0, len(candidates))
	for _, c := range candidates {
		if c.Handle != "" {
			selected = append(selected, c)
		}
	}

	if len(selected) == 0 {
		return Target{}, fmt.Errorf("no sns provider configured")
	}
	if len(selected) > 1 {
		return Target{}, fmt.Errorf("multiple sns providers configured; specify only one")
	}
	return selected[0], nil
}
