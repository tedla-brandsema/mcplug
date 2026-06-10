package config

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

//go:embed mcpfs.cfg.json
var embeddedGlobalConfig []byte

//go:embed mcpfs.starter.cfg.json
var starterConfig []byte

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

type Config struct {
	Server ServerConfig `json:"server"`

	// MCPServers configures the upstream MCP servers whose tools MCPFS
	// aggregates. The shape is compatible with the Claude/Cursor
	// `mcpServers` convention; url, headers, disabled, optional, cwd,
	// includeTools, and excludeTools are MCPFS extensions.
	MCPServers map[string]MCPServer `json:"mcpServers,omitempty"`
}

type MCPServer struct {
	// Stdio upstream: command is executed verbatim, never through a shell.
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	Cwd     string            `json:"cwd,omitempty"`

	// HTTP upstream.
	URL     string            `json:"url,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`

	// Disabled entries are ignored entirely. Optional entries may fail at
	// startup without aborting MCPFS; their tools stay absent until restart.
	Disabled bool `json:"disabled,omitempty"`
	Optional bool `json:"optional,omitempty"`

	IncludeTools []string `json:"includeTools,omitempty"`
	ExcludeTools []string `json:"excludeTools,omitempty"`
}

// IsHTTP reports whether the entry is an HTTP upstream.
func (s MCPServer) IsHTTP() bool { return s.URL != "" }

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

		// 0600: configs may hold header/env secrets.
		if err := os.WriteFile(path, embeddedGlobalConfig, 0o600); err != nil {
			return Config{}, fmt.Errorf("write config: %w", err)
		}
	}

	return Load(path)
}

// WriteStarter writes the commented starter config (example mcpServers
// entries, disabled) to path unless a file already exists there. It reports
// whether a new file was created.
func WriteStarter(path string) (bool, error) {
	if path == "" {
		return false, fmt.Errorf("config path is required")
	}

	if _, err := os.Stat(path); err == nil {
		return false, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return false, fmt.Errorf("stat config: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return false, fmt.Errorf("create config dir: %w", err)
	}

	// 0600: configs may hold header/env secrets.
	if err := os.WriteFile(path, starterConfig, 0o600); err != nil {
		return false, fmt.Errorf("write config: %w", err)
	}

	return true, nil
}

func DefaultGlobalPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}

	return filepath.Join(dir, "mcpfs", GlobalConfigFileName), nil
}

// WorldReadableWarning returns a non-empty warning when the config file at
// path is world-readable while cfg carries headers or env values, which may
// contain secrets. The caller decides how to log it.
func WorldReadableWarning(path string, cfg Config) string {
	hasSecrets := false
	for _, srv := range cfg.MCPServers {
		if len(srv.Headers) > 0 || len(srv.Env) > 0 {
			hasSecrets = true
			break
		}
	}
	if !hasSecrets {
		return ""
	}

	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm()&0o004 == 0 {
		return ""
	}

	return fmt.Sprintf("config %s contains headers/env values but is world-readable; consider chmod 600", path)
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

	if err := c.validateMCPServers(); err != nil {
		return err
	}

	return nil
}

// validateMCPServers structurally validates every entry, including disabled
// ones; it never checks command availability or network reachability.
func (c *Config) validateMCPServers() error {
	prefixes := make(map[string]string, len(c.MCPServers))

	for name, srv := range c.MCPServers {
		if name == "" {
			return fmt.Errorf("mcpServers entries must have a non-empty name")
		}

		sanitized, err := SanitizeServerName(name)
		if err != nil {
			return fmt.Errorf("mcpServers[%q]: %w", name, err)
		}
		if other, ok := prefixes[sanitized]; ok {
			return fmt.Errorf("mcpServers[%q] and mcpServers[%q] both sanitize to tool prefix %q", name, other, sanitized)
		}
		prefixes[sanitized] = name

		hasCommand := srv.Command != ""
		hasURL := srv.URL != ""
		switch {
		case hasCommand && hasURL:
			return fmt.Errorf("mcpServers[%q]: command and url are mutually exclusive", name)
		case !hasCommand && !hasURL:
			return fmt.Errorf("mcpServers[%q]: exactly one of command or url is required", name)
		}

		if hasURL {
			u, err := url.Parse(srv.URL)
			if err != nil {
				return fmt.Errorf("mcpServers[%q]: invalid url: %w", name, err)
			}
			if u.Scheme != "http" && u.Scheme != "https" {
				return fmt.Errorf("mcpServers[%q]: url scheme must be http or https", name)
			}
			if u.Host == "" {
				return fmt.Errorf("mcpServers[%q]: url host is required", name)
			}
			if len(srv.Args) > 0 {
				return fmt.Errorf("mcpServers[%q]: args only apply to command upstreams", name)
			}
			if len(srv.Env) > 0 {
				return fmt.Errorf("mcpServers[%q]: env only applies to command upstreams", name)
			}
			if srv.Cwd != "" {
				return fmt.Errorf("mcpServers[%q]: cwd only applies to command upstreams", name)
			}
		} else if len(srv.Headers) > 0 {
			return fmt.Errorf("mcpServers[%q]: headers only apply to url upstreams", name)
		}

		for k := range srv.Env {
			if k == "" {
				return fmt.Errorf("mcpServers[%q]: env keys must not be empty", name)
			}
		}
		for k := range srv.Headers {
			if k == "" {
				return fmt.Errorf("mcpServers[%q]: header keys must not be empty", name)
			}
		}

		if len(srv.IncludeTools) > 0 && len(srv.ExcludeTools) > 0 {
			return fmt.Errorf("mcpServers[%q]: includeTools and excludeTools are mutually exclusive", name)
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
