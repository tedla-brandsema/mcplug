package config

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//go:embed mcpfs.cfg.json
var embeddedGlobalConfig []byte

const GlobalConfigFileName = "mcpfs.cfg.json"

type AuthMode string

const (
	AuthModeNone   AuthMode = "none"
	AuthModeBearer AuthMode = "bearer"
	AuthModeOIDC   AuthMode = "oidc"
)

type AuthConfig struct {
	Mode AuthMode `json:"mode,omitempty"`

	// Used when mode is "bearer".
	TokenEnv string `json:"token_env,omitempty"`

	// Used when mode is "oidc".
	Issuer          string   `json:"issuer,omitempty"`
	Audience        string   `json:"audience,omitempty"`
	JWKSURL         string   `json:"jwks_url,omitempty"`
	AllowedEmails   []string `json:"allowed_emails,omitempty"`
	AllowedSubjects []string `json:"allowed_subjects,omitempty"`
}

type Mode string

const (
	ModeRead      Mode = "read"
	ModeReadWrite Mode = "read_write"
)

type CommandMode string

const (
	CommandModeDisabled   CommandMode = "disabled"
	CommandModePredefined CommandMode = "predefined"
	CommandModeUnguarded  CommandMode = "unguarded"
)

type Config struct {
	Server   ServerConfig  `json:"server"`
	Roots    []RootConfig  `json:"roots"`
	Commands CommandConfig `json:"commands,omitempty"`
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

type CommandConfig struct {
	Mode     CommandMode     `json:"mode,omitempty"`
	Defaults CommandDefaults `json:"defaults,omitempty"`
	Items    []CommandItem   `json:"items,omitempty"`
}

type CommandDefaults struct {
	TimeoutSeconds int `json:"timeout_seconds,omitempty"`
	MaxOutputBytes int `json:"max_output_bytes,omitempty"`
}

type CommandItem struct {
	ID             string   `json:"id"`
	Description    string   `json:"description,omitempty"`
	RootID         string   `json:"root_id"`
	Workdir        string   `json:"workdir,omitempty"`
	Command        []string `json:"command"`
	TimeoutSeconds int      `json:"timeout_seconds,omitempty"`
	MaxOutputBytes int      `json:"max_output_bytes,omitempty"`
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	return Decode(data)
}

func LoadOrCreate(path string) (Config, string, error) {
	if path != "" {
		cfg, err := Load(path)
		if err != nil {
			return Config{}, "", err
		}
		return cfg, path, nil
	}

	globalPath, err := DefaultGlobalPath()
	if err != nil {
		return Config{}, "", err
	}

	cfg, err := LoadOrCreateGlobal(globalPath)
	if err != nil {
		return Config{}, "", err
	}

	return cfg, globalPath, nil
}

func LoadOrCreateGlobal(path string) (Config, error) {
	if path == "" {
		return Config{}, fmt.Errorf("config path is required")
	}

	if _, err := os.Stat(path); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return Config{}, fmt.Errorf("stat config: %w", err)
		}

		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return Config{}, fmt.Errorf("create config dir: %w", err)
		}

		if err := os.WriteFile(path, embeddedGlobalConfig, 0o644); err != nil {
			return Config{}, fmt.Errorf("write config: %w", err)
		}
	}

	return Load(path)
}

func DefaultGlobalPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}

	return filepath.Join(dir, "mcpfs", GlobalConfigFileName), nil
}

func Decode(data []byte) (Config, error) {
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

	if err := c.Commands.Validate(seen); err != nil {
		return err
	}

	return nil
}

func (c *CommandConfig) Validate(rootIDs map[string]struct{}) error {
	if c.Mode == "" {
		c.Mode = CommandModeDisabled
	}

	switch c.Mode {
	case CommandModeDisabled, CommandModePredefined, CommandModeUnguarded:
		// Valid.
	default:
		return fmt.Errorf("commands.mode must be %q, %q, or %q", CommandModeDisabled, CommandModePredefined, CommandModeUnguarded)
	}

	if c.Defaults.TimeoutSeconds < 0 {
		return fmt.Errorf("commands.defaults.timeout_seconds must be >= 0")
	}
	if c.Defaults.MaxOutputBytes < 0 {
		return fmt.Errorf("commands.defaults.max_output_bytes must be >= 0")
	}

	seen := make(map[string]struct{}, len(c.Items))
	for i, item := range c.Items {
		if item.ID == "" {
			return fmt.Errorf("commands.items[%d].id is required", i)
		}
		if _, ok := seen[item.ID]; ok {
			return fmt.Errorf("duplicate command id %q", item.ID)
		}
		seen[item.ID] = struct{}{}

		if item.RootID == "" {
			return fmt.Errorf("commands.items[%d].root_id is required", i)
		}
		if _, ok := rootIDs[item.RootID]; !ok {
			return fmt.Errorf("commands.items[%d].root_id %q does not match a configured root", i, item.RootID)
		}

		if len(item.Command) == 0 {
			return fmt.Errorf("commands.items[%d].command is required", i)
		}
		for j, arg := range item.Command {
			if arg == "" {
				return fmt.Errorf("commands.items[%d].command[%d] must not be empty", i, j)
			}
		}

		if item.TimeoutSeconds < 0 {
			return fmt.Errorf("commands.items[%d].timeout_seconds must be >= 0", i)
		}
		if item.MaxOutputBytes < 0 {
			return fmt.Errorf("commands.items[%d].max_output_bytes must be >= 0", i)
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

	case AuthModeOIDC:
		if s.Auth.Issuer == "" {
			return fmt.Errorf("server.auth.issuer is required when server.auth.mode is %q", AuthModeOIDC)
		}
		if s.Auth.Audience == "" {
			return fmt.Errorf("server.auth.audience is required when server.auth.mode is %q", AuthModeOIDC)
		}
		if s.Auth.JWKSURL == "" {
			return fmt.Errorf("server.auth.jwks_url is required when server.auth.mode is %q", AuthModeOIDC)
		}
		if len(s.Auth.AllowedEmails) == 0 && len(s.Auth.AllowedSubjects) == 0 {
			return fmt.Errorf("server.auth.allowed_emails or server.auth.allowed_subjects is required when server.auth.mode is %q", AuthModeOIDC)
		}
		return nil

	default:
		return fmt.Errorf("unsupported server.auth.mode %q", s.Auth.Mode)
	}
}
