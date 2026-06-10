package upstream

import (
	"context"
	"errors"
	"log/slog"
	"sync/atomic"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/tedla-brandsema/mcpfs/internal/config"
)

// StubUpstream is a configurable in-memory Upstream for tests.
type StubUpstream struct {
	UpstreamName string
	Tools        []*mcp.Tool
	ConnectErr   error
	ListErr      error
	CallResult   *mcp.CallToolResult
	CallErr      error

	ListCalls atomic.Int64
	CallCalls atomic.Int64
	Closed    atomic.Bool
}

func (s *StubUpstream) Name() string                  { return s.UpstreamName }
func (s *StubUpstream) Connect(context.Context) error { return s.ConnectErr }
func (s *StubUpstream) Close() error                  { s.Closed.Store(true); return nil }

func (s *StubUpstream) ListTools(context.Context) ([]*mcp.Tool, error) {
	s.ListCalls.Add(1)
	return s.Tools, s.ListErr
}

func (s *StubUpstream) CallTool(context.Context, string, map[string]any) (*mcp.CallToolResult, error) {
	s.CallCalls.Add(1)
	return s.CallResult, s.CallErr
}

// withStubFactory substitutes the upstream factory for the test's duration.
func withStubFactory(t *testing.T, stubs map[string]*StubUpstream) {
	t.Helper()
	old := newUpstream
	newUpstream = func(name string, _ config.MCPServer, _ *slog.Logger) Upstream {
		return stubs[name]
	}
	t.Cleanup(func() { newUpstream = old })
}

func TestStartAllCollectsActiveUpstreams(t *testing.T) {
	stubs := map[string]*StubUpstream{
		"a": {UpstreamName: "a", Tools: []*mcp.Tool{{Name: "t1"}}},
		"b": {UpstreamName: "b", Tools: []*mcp.Tool{{Name: "t2"}}},
	}
	withStubFactory(t, stubs)

	cfg := config.Config{MCPServers: map[string]config.MCPServer{
		"a": {Command: "x"},
		"b": {Command: "y"},
	}}

	result, err := StartAll(t.Context(), cfg, testLogger(t))
	if err != nil {
		t.Fatalf("StartAll returned error: %v", err)
	}

	if len(result.ActiveUpstreams) != 2 {
		t.Fatalf("len(ActiveUpstreams) = %d, want 2", len(result.ActiveUpstreams))
	}
	if len(result.SkippedOptional) != 0 {
		t.Fatalf("len(SkippedOptional) = %d, want 0", len(result.SkippedOptional))
	}
	for _, a := range result.ActiveUpstreams {
		if len(a.Tools) != 1 {
			t.Fatalf("upstream %q tools = %d, want 1", a.Upstream.Name(), len(a.Tools))
		}
	}
}

func TestStartAllRequiredFailureAbortsAndClosesStarted(t *testing.T) {
	stubs := map[string]*StubUpstream{
		"a": {UpstreamName: "a"},
		"b": {UpstreamName: "b", ConnectErr: errors.New("boom")},
	}
	withStubFactory(t, stubs)

	cfg := config.Config{MCPServers: map[string]config.MCPServer{
		"a": {Command: "x"},
		"b": {Command: "y"},
	}}

	_, err := StartAll(t.Context(), cfg, testLogger(t))
	if err == nil {
		t.Fatal("StartAll returned nil error")
	}
	if !stubs["a"].Closed.Load() {
		t.Fatal("previously started upstream was not closed")
	}
	if !stubs["b"].Closed.Load() {
		t.Fatal("failed upstream was not closed")
	}
}

func TestStartAllOptionalFailureIsSkipped(t *testing.T) {
	stubs := map[string]*StubUpstream{
		"a": {UpstreamName: "a"},
		"b": {UpstreamName: "b", ConnectErr: errors.New("boom")},
	}
	withStubFactory(t, stubs)

	cfg := config.Config{MCPServers: map[string]config.MCPServer{
		"a": {Command: "x"},
		"b": {Command: "y", Optional: true},
	}}

	result, err := StartAll(t.Context(), cfg, testLogger(t))
	if err != nil {
		t.Fatalf("StartAll returned error: %v", err)
	}

	if len(result.ActiveUpstreams) != 1 || result.ActiveUpstreams[0].Upstream.Name() != "a" {
		t.Fatalf("ActiveUpstreams = %+v, want only a", result.ActiveUpstreams)
	}
	if len(result.SkippedOptional) != 1 || result.SkippedOptional[0].Name != "b" {
		t.Fatalf("SkippedOptional = %+v, want b", result.SkippedOptional)
	}
}

func TestStartAllOptionalListFailureIsSkipped(t *testing.T) {
	stubs := map[string]*StubUpstream{
		"a": {UpstreamName: "a", ListErr: errors.New("listing broke")},
	}
	withStubFactory(t, stubs)

	cfg := config.Config{MCPServers: map[string]config.MCPServer{
		"a": {Command: "x", Optional: true},
	}}

	result, err := StartAll(t.Context(), cfg, testLogger(t))
	if err != nil {
		t.Fatalf("StartAll returned error: %v", err)
	}
	if len(result.SkippedOptional) != 1 {
		t.Fatalf("SkippedOptional = %+v, want one entry", result.SkippedOptional)
	}
	if !stubs["a"].Closed.Load() {
		t.Fatal("upstream that failed listing was not closed")
	}
}

func TestStartAllIgnoresDisabledEntries(t *testing.T) {
	stubs := map[string]*StubUpstream{
		"on":  {UpstreamName: "on"},
		"off": {UpstreamName: "off", ConnectErr: errors.New("must never be started")},
	}
	withStubFactory(t, stubs)

	cfg := config.Config{MCPServers: map[string]config.MCPServer{
		"on":  {Command: "x"},
		"off": {Command: "y", Disabled: true},
	}}

	result, err := StartAll(t.Context(), cfg, testLogger(t))
	if err != nil {
		t.Fatalf("StartAll returned error: %v", err)
	}
	if len(result.ActiveUpstreams) != 1 || result.ActiveUpstreams[0].Upstream.Name() != "on" {
		t.Fatalf("ActiveUpstreams = %+v, want only on", result.ActiveUpstreams)
	}
}
