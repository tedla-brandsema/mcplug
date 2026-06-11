package main

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func writeConfig(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cfg.json")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func lsLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

func TestLsRequiredFailureExitsNonZero(t *testing.T) {
	path := writeConfig(t, `{
		"server": {"name": "mcplug", "version": "0.5.0", "transport": "stdio"},
		"mcpServers": {
			"broken": {"command": "/no/such/binary/anywhere"}
		}
	}`)

	if code := runLs([]string{"-config", path}, lsLogger()); code != 1 {
		t.Fatalf("runLs exit code = %d, want 1", code)
	}
}

func TestLsOptionalFailureExitsZero(t *testing.T) {
	path := writeConfig(t, `{
		"server": {"name": "mcplug", "version": "0.5.0", "transport": "stdio"},
		"mcpServers": {
			"broken": {"command": "/no/such/binary/anywhere", "optional": true}
		}
	}`)

	if code := runLs([]string{"-config", path}, lsLogger()); code != 0 {
		t.Fatalf("runLs exit code = %d, want 0", code)
	}
}

func TestLsDisabledOnlyExitsZero(t *testing.T) {
	path := writeConfig(t, `{
		"server": {"name": "mcplug", "version": "0.5.0", "transport": "stdio"},
		"mcpServers": {
			"off": {"command": "/no/such/binary/anywhere", "disabled": true}
		}
	}`)

	if code := runLs([]string{"-config", path}, lsLogger()); code != 0 {
		t.Fatalf("runLs exit code = %d, want 0", code)
	}
}

func TestLsEmptyMCPServersExitsZero(t *testing.T) {
	path := writeConfig(t, `{
		"server": {"name": "mcplug", "version": "0.5.0", "transport": "stdio"},
		"mcpServers": {}
	}`)

	if code := runLs([]string{"-config", path}, lsLogger()); code != 0 {
		t.Fatalf("runLs exit code = %d, want 0", code)
	}
}

func TestLsInvalidConfigExitsNonZero(t *testing.T) {
	path := writeConfig(t, `{
		"server": {"name": "mcplug", "version": "0.5.0", "transport": "stdio"},
		"mcpServers": {
			"bad": {"command": "x", "url": "https://example.com/mcp"}
		}
	}`)

	if code := runLs([]string{"-config", path}, lsLogger()); code != 1 {
		t.Fatalf("runLs exit code = %d, want 1", code)
	}
}
