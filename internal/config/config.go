package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Mode string

const (
	ModeRead      Mode = "read"
	ModeReadWrite Mode = "read_write"
)

type Config struct {
	Server ServerConfig `json:"server"`
	Roots  []RootConfig `json:"roots"`
}

type ServerConfig struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Transport string `json:"transport"`
}

type RootConfig struct {
	ID           string   `json:"id"`
	Path         string   `json:"path"`
	Mode         Mode     `json:"mode"`
	Include      []string `json:"include"`
	Exclude      []string `json:"exclude"`
	UseGitignore bool     `json:"use_gitignore"`
	MaxFileBytes int64    `json:"max_file_bytes"`
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) Validate() error {
	if c.Server.Name == "" {
		return fmt.Errorf("server.name is required")
	}
	if c.Server.Version == "" {
		return fmt.Errorf("server.version is required")
	}
	if c.Server.Transport == "" {
		return fmt.Errorf("server.transport is required")
	}
	if c.Server.Transport != "stdio" {
		return fmt.Errorf("unsupported server.transport %q", c.Server.Transport)
	}
	if len(c.Roots) == 0 {
		return fmt.Errorf("at least one root is required")
	}

	seen := make(map[string]struct{}, len(c.Roots))
	for i, root := range c.Roots {
		if root.ID == "" {
			return fmt.Errorf("roots[%d].id is required", i)
		}
		if _, ok := seen[root.ID]; ok {
			return fmt.Errorf("duplicate root id %q", root.ID)
		}
		seen[root.ID] = struct{}{}

		if root.Path == "" {
			return fmt.Errorf("roots[%d].path is required", i)
		}

		switch root.Mode {
		case ModeRead, ModeReadWrite:
		default:
			return fmt.Errorf("roots[%d].mode must be %q or %q", i, ModeRead, ModeReadWrite)
		}

		if root.MaxFileBytes < 0 {
			return fmt.Errorf("roots[%d].max_file_bytes must be >= 0", i)
		}
	}

	return nil
}