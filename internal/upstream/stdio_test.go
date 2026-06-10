package upstream

import (
	"log/slog"
	"slices"
	"testing"

	"github.com/tedla-brandsema/mcplug/internal/config"
)

func testLogger(t *testing.T) *slog.Logger {
	t.Helper()
	return slog.New(slog.DiscardHandler)
}

func configStdio() config.MCPServer {
	return config.MCPServer{Command: "true"}
}

func TestMergeEnvConfigWins(t *testing.T) {
	base := []string{"PATH=/usr/bin", "HOME=/home/ted", "DEBUG=0"}
	merged := mergeEnv(base, map[string]string{"DEBUG": "1", "API_TOKEN": "x"})

	// Base entries stay first; config entries are appended in sorted key
	// order, so exec resolves duplicates to the config value.
	want := []string{"PATH=/usr/bin", "HOME=/home/ted", "DEBUG=0", "API_TOKEN=x", "DEBUG=1"}
	if !slices.Equal(merged, want) {
		t.Fatalf("mergeEnv = %v, want %v", merged, want)
	}
}

func TestMergeEnvNoExtra(t *testing.T) {
	base := []string{"PATH=/usr/bin"}
	if got := mergeEnv(base, nil); !slices.Equal(got, base) {
		t.Fatalf("mergeEnv = %v, want %v", got, base)
	}
}

func TestStdioUpstreamCallBeforeConnectReturnsToolError(t *testing.T) {
	u := newStdioUpstream("child", configStdio(), testLogger(t))

	result, err := u.CallTool(t.Context(), "echo", nil)
	if err != nil {
		t.Fatalf("CallTool returned protocol error: %v (want tool error)", err)
	}
	if !result.IsError {
		t.Fatal("CallTool result.IsError = false, want true")
	}
}

func TestStdioUpstreamCallAfterCloseReturnsToolError(t *testing.T) {
	u := newStdioUpstream("child", configStdio(), testLogger(t))
	if err := u.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}

	result, err := u.CallTool(t.Context(), "echo", nil)
	if err != nil {
		t.Fatalf("CallTool returned protocol error: %v (want tool error)", err)
	}
	if !result.IsError {
		t.Fatal("CallTool result.IsError = false, want true")
	}
}
