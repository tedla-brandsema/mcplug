package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateAcceptsStdioEntry(t *testing.T) {
	cfg := validConfig()
	cfg.MCPServers = map[string]MCPServer{
		"filesystem": {
			Command: "npx",
			Args:    []string{"-y", "@modelcontextprotocol/server-filesystem", "/tmp"},
			Env:     map[string]string{"DEBUG": "1"},
			Cwd:     "/tmp",
		},
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
}

func TestValidateAcceptsURLEntry(t *testing.T) {
	cfg := validConfig()
	cfg.MCPServers = map[string]MCPServer{
		"remote": {
			URL:     "https://example.com/mcp",
			Headers: map[string]string{"Authorization": "Bearer secret"},
		},
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
}

func TestValidateRejectsCommandAndURL(t *testing.T) {
	cfg := validConfig()
	cfg.MCPServers = map[string]MCPServer{
		"both": {Command: "npx", URL: "https://example.com/mcp"},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error")
	}
	assertErrorContains(t, err, "mutually exclusive")
}

func TestValidateRejectsNeitherCommandNorURL(t *testing.T) {
	cfg := validConfig()
	cfg.MCPServers = map[string]MCPServer{
		"neither": {},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error")
	}
	assertErrorContains(t, err, "exactly one of command or url")
}

func TestValidateRejectsNonHTTPURLScheme(t *testing.T) {
	cfg := validConfig()
	cfg.MCPServers = map[string]MCPServer{
		"bad": {URL: "ftp://example.com/mcp"},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error")
	}
	assertErrorContains(t, err, "scheme must be http or https")
}

func TestValidateRejectsIncludeAndExcludeTools(t *testing.T) {
	cfg := validConfig()
	cfg.MCPServers = map[string]MCPServer{
		"filtered": {
			Command:      "npx",
			IncludeTools: []string{"read_file"},
			ExcludeTools: []string{"write_file"},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error")
	}
	assertErrorContains(t, err, "includeTools and excludeTools are mutually exclusive")
}

func TestValidateRejectsHeadersOnStdioEntry(t *testing.T) {
	cfg := validConfig()
	cfg.MCPServers = map[string]MCPServer{
		"stdio": {
			Command: "npx",
			Headers: map[string]string{"Authorization": "Bearer x"},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error")
	}
	assertErrorContains(t, err, "headers only apply to url upstreams")
}

func TestValidateRejectsStdioFieldsOnURLEntry(t *testing.T) {
	for field, srv := range map[string]MCPServer{
		"args": {URL: "https://example.com/mcp", Args: []string{"-y"}},
		"env":  {URL: "https://example.com/mcp", Env: map[string]string{"A": "B"}},
		"cwd":  {URL: "https://example.com/mcp", Cwd: "/tmp"},
	} {
		cfg := validConfig()
		cfg.MCPServers = map[string]MCPServer{"remote": srv}

		err := cfg.Validate()
		if err == nil {
			t.Fatalf("%s: Validate returned nil error", field)
		}
		assertErrorContains(t, err, "only appl")
	}
}

func TestValidateRejectsArgsWithoutCommand(t *testing.T) {
	cfg := validConfig()
	cfg.MCPServers = map[string]MCPServer{
		"argsonly": {Args: []string{"-y"}},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error")
	}
	assertErrorContains(t, err, "exactly one of command or url")
}

func TestValidateStructurallyValidatesDisabledEntries(t *testing.T) {
	cfg := validConfig()
	cfg.MCPServers = map[string]MCPServer{
		"broken": {Disabled: true},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error for structurally invalid disabled entry")
	}
	assertErrorContains(t, err, "exactly one of command or url")
}

func TestValidateAllowsDisabledEntryWithMissingBinary(t *testing.T) {
	cfg := validConfig()
	cfg.MCPServers = map[string]MCPServer{
		"future": {Command: "/no/such/binary/anywhere", Disabled: true},
	}

	// No command-availability check: structurally valid is enough.
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
}

func TestValidateParsesOptionalFlag(t *testing.T) {
	cfg, err := Decode([]byte(`{
		"server": {"name": "mcplug", "version": "0.5.0", "transport": "stdio"},
		"mcpServers": {
			"maybe": {"command": "npx", "optional": true}
		}
	}`))
	if err != nil {
		t.Fatalf("Decode returned error: %v", err)
	}

	if !cfg.MCPServers["maybe"].Optional {
		t.Fatal("Optional = false, want true")
	}
}

func TestValidateRejectsSanitizedPrefixCollision(t *testing.T) {
	cfg := validConfig()
	cfg.MCPServers = map[string]MCPServer{
		"my.server": {Command: "a"},
		"my_server": {Command: "b"},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error")
	}
	assertErrorContains(t, err, "sanitize to tool prefix")
}

func TestValidateRejectsEmptySanitizedName(t *testing.T) {
	cfg := validConfig()
	cfg.MCPServers = map[string]MCPServer{
		"...": {Command: "a"},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error")
	}
	assertErrorContains(t, err, "empty tool prefix")
}

func TestValidateAllowsEmptyMCPServers(t *testing.T) {
	cfg := validConfig()
	cfg.MCPServers = nil

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
}

func TestSanitizeServerName(t *testing.T) {
	tests := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{in: "filesystem", want: "filesystem"},
		{in: "my-server", want: "my-server"},
		{in: "my.server", want: "my_server"},
		{in: "my..server", want: "my_server"},
		{in: "My_Server_2", want: "My_Server_2"},
		{in: "a__b", want: "a_b"},
		{in: "_leading", want: "leading"},
		{in: "trailing_", want: "trailing"},
		{in: "ünïcode", want: "n_code"},
		{in: "über", want: "ber"},
		{in: "1password", want: "server_1password"},
		{in: "-dash", want: "server_-dash"},
		{in: "...", wantErr: true},
		{in: "___", wantErr: true},
		{in: "", wantErr: true},
	}

	for _, tc := range tests {
		got, err := SanitizeServerName(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("SanitizeServerName(%q) = %q, want error", tc.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("SanitizeServerName(%q) returned error: %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("SanitizeServerName(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestRedactValue(t *testing.T) {
	tests := []struct {
		key   string
		value string
		want  string
	}{
		{"Authorization", "Bearer abc", RedactedValue},
		{"X-Api-Key", "abc", RedactedValue},
		{"API_KEY", "abc", RedactedValue},
		{"MY_SECRET", "abc", RedactedValue},
		{"GITHUB_TOKEN", "abc", RedactedValue},
		{"DB_PASSWORD", "abc", RedactedValue},
		{"DEBUG", "1", "1"},
	}

	for _, tc := range tests {
		if got := RedactValue(tc.key, tc.value); got != tc.want {
			t.Errorf("RedactValue(%q, %q) = %q, want %q", tc.key, tc.value, got, tc.want)
		}
	}
}

func TestWorldReadableWarning(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.json")
	if err := os.WriteFile(path, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	withSecrets := validConfig()
	withSecrets.MCPServers = map[string]MCPServer{
		"remote": {URL: "https://example.com/mcp", Headers: map[string]string{"Authorization": "x"}},
	}

	if w := WorldReadableWarning(path, withSecrets); w == "" {
		t.Fatal("expected warning for world-readable config with headers")
	}

	noSecrets := validConfig()
	if w := WorldReadableWarning(path, noSecrets); w != "" {
		t.Fatalf("unexpected warning: %q", w)
	}

	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatal(err)
	}
	if w := WorldReadableWarning(path, withSecrets); w != "" {
		t.Fatalf("unexpected warning for 0600 config: %q", w)
	}
}
