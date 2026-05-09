package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateDefaultsTransportToStdio(t *testing.T) {
	cfg := validConfig()
	cfg.Server.Transport = ""

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if cfg.Server.Transport != "stdio" {
		t.Fatalf("Transport = %q, want %q", cfg.Server.Transport, "stdio")
	}
}

func TestValidateHTTPDefaultsAddrPathAndAuth(t *testing.T) {
	cfg := validConfig()
	cfg.Server.Transport = "http"
	cfg.Server.Addr = ""
	cfg.Server.Path = ""
	cfg.Server.Auth = nil
	cfg.Server.RequireAuth = false
	cfg.Server.AuthTokenEnv = ""

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if cfg.Server.Addr != "127.0.0.1:8080" {
		t.Fatalf("Addr = %q, want %q", cfg.Server.Addr, "127.0.0.1:8080")
	}
	if cfg.Server.Path != "/mcp" {
		t.Fatalf("Path = %q, want %q", cfg.Server.Path, "/mcp")
	}
	if cfg.Server.Auth == nil {
		t.Fatal("Auth is nil")
	}
	if cfg.Server.Auth.Mode != AuthModeNone {
		t.Fatalf("Auth.Mode = %q, want %q", cfg.Server.Auth.Mode, AuthModeNone)
	}
}

func TestValidateLegacyRequireAuthMapsToBearer(t *testing.T) {
	cfg := validConfig()
	cfg.Server.Transport = "http"
	cfg.Server.Auth = nil
	cfg.Server.RequireAuth = true
	cfg.Server.AuthTokenEnv = "MCPFS_TOKEN"

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if cfg.Server.Auth == nil {
		t.Fatal("Auth is nil")
	}
	if cfg.Server.Auth.Mode != AuthModeBearer {
		t.Fatalf("Auth.Mode = %q, want %q", cfg.Server.Auth.Mode, AuthModeBearer)
	}
	if cfg.Server.Auth.TokenEnv != "MCPFS_TOKEN" {
		t.Fatalf("Auth.TokenEnv = %q, want %q", cfg.Server.Auth.TokenEnv, "MCPFS_TOKEN")
	}
}

func TestValidateLegacyRequireAuthRequiresTokenEnv(t *testing.T) {
	cfg := validConfig()
	cfg.Server.Transport = "http"
	cfg.Server.Auth = nil
	cfg.Server.RequireAuth = true
	cfg.Server.AuthTokenEnv = ""

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error")
	}
	assertErrorContains(t, err, "server.auth.token_env is required")
}

func TestValidateExplicitBearerFallsBackToLegacyTokenEnv(t *testing.T) {
	cfg := validConfig()
	cfg.Server.Transport = "http"
	cfg.Server.Auth = &AuthConfig{
		Mode: AuthModeBearer,
	}
	cfg.Server.AuthTokenEnv = "MCPFS_TOKEN"

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if cfg.Server.Auth.TokenEnv != "MCPFS_TOKEN" {
		t.Fatalf("Auth.TokenEnv = %q, want %q", cfg.Server.Auth.TokenEnv, "MCPFS_TOKEN")
	}
}

func TestValidateExplicitBearerRequiresTokenEnv(t *testing.T) {
	cfg := validConfig()
	cfg.Server.Transport = "http"
	cfg.Server.Auth = &AuthConfig{
		Mode: AuthModeBearer,
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error")
	}
	assertErrorContains(t, err, "server.auth.token_env is required")
}

func TestValidateOIDCRequiresIssuer(t *testing.T) {
	cfg := validOIDCConfig()
	cfg.Server.Auth.Issuer = ""

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error")
	}
	assertErrorContains(t, err, "server.auth.issuer is required")
}

func TestValidateOIDCRequiresAudience(t *testing.T) {
	cfg := validOIDCConfig()
	cfg.Server.Auth.Audience = ""

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error")
	}
	assertErrorContains(t, err, "server.auth.audience is required")
}

func TestValidateOIDCRequiresJWKSURL(t *testing.T) {
	cfg := validOIDCConfig()
	cfg.Server.Auth.JWKSURL = ""

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error")
	}
	assertErrorContains(t, err, "server.auth.jwks_url is required")
}

func TestValidateOIDCRequiresIdentityAllowlist(t *testing.T) {
	cfg := validOIDCConfig()
	cfg.Server.Auth.AllowedEmails = nil
	cfg.Server.Auth.AllowedSubjects = nil

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error")
	}
	assertErrorContains(t, err, "server.auth.allowed_emails or server.auth.allowed_subjects is required")
}

func TestValidateOIDCAllowsSubjectAllowlist(t *testing.T) {
	cfg := validOIDCConfig()
	cfg.Server.Auth.AllowedEmails = nil
	cfg.Server.Auth.AllowedSubjects = []string{"user-123"}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
}

func TestValidateRejectsHTTPPathWithoutLeadingSlash(t *testing.T) {
	cfg := validConfig()
	cfg.Server.Transport = "http"
	cfg.Server.Path = "mcp"

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error")
	}
	assertErrorContains(t, err, "server.path must start with /")
}

func TestValidateRejectsUnsupportedTransport(t *testing.T) {
	cfg := validConfig()
	cfg.Server.Transport = "websocket"

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error")
	}
	assertErrorContains(t, err, "unsupported server.transport")
}

func TestValidateRejectsDuplicateRootID(t *testing.T) {
	cfg := validConfig()
	cfg.Roots = append(cfg.Roots, cfg.Roots[0])

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error")
	}
	assertErrorContains(t, err, "duplicate root id")
}

func TestValidateRejectsInvalidRootMode(t *testing.T) {
	cfg := validConfig()
	cfg.Roots[0].Mode = "write"

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error")
	}
	assertErrorContains(t, err, "mode must be")
}

func TestValidateRejectsNegativeMaxFileBytes(t *testing.T) {
	cfg := validConfig()
	cfg.Roots[0].MaxFileBytes = -1

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error")
	}
	assertErrorContains(t, err, "max_file_bytes must be >= 0")

}

func TestValidateAllowsEmptyRoots(t *testing.T) {
	cfg := validConfig()
	cfg.Roots = nil

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
}

func TestValidateDefaultsCommandModeToDisabled(t *testing.T) {
	cfg := validConfig()

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if cfg.Commands.Mode != CommandModeDisabled {
		t.Fatalf("Commands.Mode = %q, want %q", cfg.Commands.Mode, CommandModeDisabled)
	}
}

func TestValidateAllowsPredefinedCommands(t *testing.T) {
	cfg := validConfig()
	cfg.Commands = CommandConfig{
		Mode: CommandModePredefined,
		Defaults: CommandDefaults{
			TimeoutSeconds: 30,
			MaxOutputBytes: 1024,
		},
		Items: []CommandItem{
			{
				ID:      "test",
				RootID:  "repo",
				Workdir: ".",
				Command: []string{"go", "test", "./..."},
			},
		},
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
}

func TestValidateRejectsInvalidCommandMode(t *testing.T) {
	cfg := validConfig()
	cfg.Commands.Mode = "wild"

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error")
	}
	assertErrorContains(t, err, "commands.mode must be")
}

func TestValidateRejectsDuplicateCommandID(t *testing.T) {
	cfg := validConfig()
	cfg.Commands = CommandConfig{
		Mode: CommandModePredefined,
		Items: []CommandItem{
			{ID: "test", RootID: "repo", Command: []string{"go", "test"}},
			{ID: "test", RootID: "repo", Command: []string{"go", "test"}},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error")
	}
	assertErrorContains(t, err, "duplicate command id")
}

func TestValidateRejectsCommandUnknownRoot(t *testing.T) {
	cfg := validConfig()
	cfg.Commands = CommandConfig{
		Mode: CommandModePredefined,
		Items: []CommandItem{
			{ID: "test", RootID: "missing", Command: []string{"go", "test"}},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error")
	}
	assertErrorContains(t, err, "does not match a configured root")
}

func TestValidateRejectsCommandWithoutArgv(t *testing.T) {
	cfg := validConfig()
	cfg.Commands = CommandConfig{
		Mode: CommandModePredefined,
		Items: []CommandItem{
			{ID: "test", RootID: "repo"},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error")
	}
	assertErrorContains(t, err, "command is required")
}

func TestValidateRejectsNegativeCommandDefaults(t *testing.T) {
	cfg := validConfig()
	cfg.Commands.Defaults.TimeoutSeconds = -1

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error")
	}
	assertErrorContains(t, err, "commands.defaults.timeout_seconds must be >= 0")
}

func TestDecodeEmbeddedGlobalConfig(t *testing.T) {
	cfg, err := Decode(embeddedGlobalConfig)
	if err != nil {
		t.Fatalf("Decode returned error: %v", err)
	}

	if cfg.Server.Name != "mcpfs" {
		t.Fatalf("Server.Name = %q, want mcpfs", cfg.Server.Name)
	}
	if cfg.Server.Transport != "stdio" {
		t.Fatalf("Server.Transport = %q, want stdio", cfg.Server.Transport)
	}
	if len(cfg.Roots) != 0 {
		t.Fatalf("len(Roots) = %d, want 0", len(cfg.Roots))
	}
	if cfg.Commands.Mode != CommandModeDisabled {
		t.Fatalf("Commands.Mode = %q, want %q", cfg.Commands.Mode, CommandModeDisabled)
	}
}

func TestLoadOrCreateGlobalWritesEmbeddedConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "mcpfs", GlobalConfigFileName)

	cfg, err := LoadOrCreateGlobal(configPath)
	if err != nil {
		t.Fatalf("LoadOrCreateGlobal returned error: %v", err)
	}

	if cfg.Server.Name != "mcpfs" {
		t.Fatalf("Server.Name = %q, want mcpfs", cfg.Server.Name)
	}

	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("config was not written: %v", err)
	}
}

func TestLoadOrCreateGlobalLoadsExistingConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "mcpfs", GlobalConfigFileName)

	data := []byte(`{
		"server": {
			"name": "custom-mcpfs",
			"version": "9.9.9",
			"transport": "stdio"
		},
		"roots": []
	}`)

	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadOrCreateGlobal(configPath)
	if err != nil {
		t.Fatalf("LoadOrCreateGlobal returned error: %v", err)
	}

	if cfg.Server.Name != "custom-mcpfs" {
		t.Fatalf("Server.Name = %q, want custom-mcpfs", cfg.Server.Name)
	}
}

func TestLoadOrCreateUsesExplicitConfigPath(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "explicit.json")

	data := []byte(`{
		"server": {
			"name": "explicit-mcpfs",
			"version": "1.2.3",
			"transport": "stdio"
		},
		"roots": []
	}`)

	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, resolved, err := LoadOrCreate(configPath)
	if err != nil {
		t.Fatalf("LoadOrCreate returned error: %v", err)
	}

	if resolved != configPath {
		t.Fatalf("resolved = %q, want %q", resolved, configPath)
	}
	if cfg.Server.Name != "explicit-mcpfs" {
		t.Fatalf("Server.Name = %q, want explicit-mcpfs", cfg.Server.Name)
	}
}

func validConfig() Config {
	return Config{
		Server: ServerConfig{
			Name:      "mcpfs",
			Version:   "0.2.0",
			Transport: "stdio",
		},
		Roots: []RootConfig{
			{
				ID:           "repo",
				Path:         ".",
				Mode:         ModeRead,
				Include:      []string{"**/*.go"},
				Exclude:      []string{"**/.git/**"},
				UseGitignore: true,
				MaxFileBytes: 262144,
			},
		},
	}
}

func validOIDCConfig() Config {
	cfg := validConfig()
	cfg.Server.Transport = "http"
	cfg.Server.Auth = &AuthConfig{
		Mode:          AuthModeOIDC,
		Issuer:        "https://issuer.example.com",
		Audience:      "mcpfs",
		JWKSURL:       "https://issuer.example.com/jwks",
		AllowedEmails: []string{"you@example.com"},
	}
	return cfg
}

func assertErrorContains(t *testing.T, err error, want string) {
	t.Helper()

	if !strings.Contains(err.Error(), want) {
		t.Fatalf("error = %q, want substring %q", err.Error(), want)
	}
}
