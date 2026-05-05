package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type AuthMode string

const (
	AuthModeNone   AuthMode = "none"
	AuthModeBearer AuthMode = "bearer"
)

type AuthConfig struct {
	Mode AuthMode `json:"mode,omitempty"`

	// Used when mode is "bearer".
	TokenEnv string `json:"token_env,omitempty"`
}

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
	Addr      string `json:"addr,omitempty"`
	Path      string `json:"path,omitempty"`

	Auth *AuthConfig `json:"auth,omitempty"`

	// Deprecated: use auth.mode instead.
	RequireAuth bool `json:"require_auth,omitempty"`

	// Deprecated: use auth.token_env instead.
	AuthTokenEnv string `json:"auth_token_env,omitempty"`

	NgrokURL string `json:"ngrok_url,omitempty"`
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

func (c *Config) Validate() error {
	if c.Server.Name == "" {
		return fmt.Errorf("server.name is required")
	}
	if c.Server.Version == "" {
		return fmt.Errorf("server.version is required")
	}

	if c.Server.Transport == "" {
		c.Server.Transport = "stdio"
	}

	switch c.Server.Transport {
	case "stdio":
		// No extra settings required.

	case "http", "http_ngrok":
		if c.Server.Addr == "" {
			c.Server.Addr = "127.0.0.1:8080"
		}
		if c.Server.Path == "" {
			c.Server.Path = "/mcp"
		}
		if !strings.HasPrefix(c.Server.Path, "/") {
			return fmt.Errorf("server.path must start with /")
		}
		if err := c.Server.normalizeAuth(); err != nil {
			return err
		}

	default:
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

func (s *ServerConfig) normalizeAuth() error {
	if s.Auth == nil {
		if s.RequireAuth {
			s.Auth = &AuthConfig{
				Mode:     AuthModeBearer,
				TokenEnv: s.AuthTokenEnv,
			}
		} else {
			s.Auth = &AuthConfig{
				Mode: AuthModeNone,
			}
		}
	}

	if s.Auth.Mode == "" {
		if s.RequireAuth {
			s.Auth.Mode = AuthModeBearer
		} else {
			s.Auth.Mode = AuthModeNone
		}
	}

	if s.Auth.Mode == AuthModeBearer && s.Auth.TokenEnv == "" {
		s.Auth.TokenEnv = s.AuthTokenEnv
	}

	switch s.Auth.Mode {
	case AuthModeNone:
		return nil

	case AuthModeBearer:
		if s.Auth.TokenEnv == "" {
			return fmt.Errorf("server.auth.token_env is required when server.auth.mode is %q", AuthModeBearer)
		}
		return nil

	default:
		return fmt.Errorf("unsupported server.auth.mode %q", s.Auth.Mode)
	}
}
