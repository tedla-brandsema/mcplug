package mcpfs

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/tedla-brandsema/mcpfs/internal/config"
	"github.com/tedla-brandsema/mcpfs/internal/upstream"
)

type stubUpstream struct {
	name       string
	callResult *mcp.CallToolResult
	callErr    error

	listCalls atomic.Int64
	lastTool  atomic.Value
}

func (s *stubUpstream) Name() string                  { return s.name }
func (s *stubUpstream) Connect(context.Context) error { return nil }
func (s *stubUpstream) Close() error                  { return nil }

func (s *stubUpstream) ListTools(context.Context) ([]*mcp.Tool, error) {
	s.listCalls.Add(1)
	return nil, errors.New("aggregator must not list tools")
}

func (s *stubUpstream) CallTool(_ context.Context, tool string, _ map[string]any) (*mcp.CallToolResult, error) {
	s.lastTool.Store(tool)
	if s.callErr != nil {
		return nil, s.callErr
	}
	if s.callResult != nil {
		return s.callResult, nil
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: "ok from " + s.name + "/" + tool}},
	}, nil
}

func testTool(name string) *mcp.Tool {
	return &mcp.Tool{
		Name:        name,
		Description: "test tool " + name,
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"text": {Type: "string"},
			},
		},
	}
}

func gatewayConfig(servers map[string]config.MCPServer) config.Config {
	return config.Config{
		Server:     config.ServerConfig{Name: "mcpfs", Version: "2.0.0", Transport: "stdio"},
		MCPServers: servers,
	}
}

// connectClient connects an in-memory MCP client to the aggregated server.
func connectClient(t *testing.T, server *Server) *mcp.ClientSession {
	t.Helper()

	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	if _, err := server.MCP.Connect(t.Context(), serverTransport, nil); err != nil {
		t.Fatalf("server connect: %v", err)
	}

	session, err := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "test"}, nil).
		Connect(t.Context(), clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })
	return session
}

func buildTestServer(t *testing.T, cfg config.Config, active []upstream.ActiveUpstream, logger *slog.Logger) *Server {
	t.Helper()
	server, err := BuildServer(cfg, &upstream.StartupResult{ActiveUpstreams: active}, logger)
	if err != nil {
		t.Fatalf("BuildServer returned error: %v", err)
	}
	return server
}

func TestBuildServerPrefixesAndProxiesTools(t *testing.T) {
	ups := &stubUpstream{name: "alpha"}
	cfg := gatewayConfig(map[string]config.MCPServer{"alpha": {Command: "x"}})

	server := buildTestServer(t, cfg, []upstream.ActiveUpstream{
		{Upstream: ups, Tools: []*mcp.Tool{testTool("echo")}},
	}, testLogger())

	session := connectClient(t, server)

	tools, err := session.ListTools(t.Context(), &mcp.ListToolsParams{})
	if err != nil {
		t.Fatalf("ListTools returned error: %v", err)
	}
	if len(tools.Tools) != 1 || tools.Tools[0].Name != "alpha_echo" {
		t.Fatalf("tools = %+v, want one tool alpha_echo", tools.Tools)
	}

	result, err := session.CallTool(t.Context(), &mcp.CallToolParams{
		Name:      "alpha_echo",
		Arguments: map[string]any{"text": "hi"},
	})
	if err != nil {
		t.Fatalf("CallTool returned error: %v", err)
	}
	if result.IsError {
		t.Fatalf("CallTool result is an error: %+v", result.Content)
	}
	if got := ups.lastTool.Load(); got != "echo" {
		t.Fatalf("upstream received tool %v, want echo (original name)", got)
	}
}

func TestBuildServerDoesNotListUpstreams(t *testing.T) {
	ups := &stubUpstream{name: "alpha"}
	cfg := gatewayConfig(map[string]config.MCPServer{"alpha": {Command: "x"}})

	buildTestServer(t, cfg, []upstream.ActiveUpstream{
		{Upstream: ups, Tools: []*mcp.Tool{testTool("echo")}},
	}, testLogger())

	if n := ups.listCalls.Load(); n != 0 {
		t.Fatalf("BuildServer listed tools %d times, want 0 (startup owns listing)", n)
	}
}

func TestBuildServerDeepCopiesTools(t *testing.T) {
	original := testTool("echo")
	ups := &stubUpstream{name: "alpha"}
	cfg := gatewayConfig(map[string]config.MCPServer{"alpha": {Command: "x"}})

	buildTestServer(t, cfg, []upstream.ActiveUpstream{
		{Upstream: ups, Tools: []*mcp.Tool{original}},
	}, testLogger())

	if original.Name != "echo" {
		t.Fatalf("upstream-owned tool was renamed to %q", original.Name)
	}
	if original.Description != "test tool echo" {
		t.Fatalf("upstream-owned tool description mutated: %q", original.Description)
	}
}

func TestBuildServerSameToolNameAcrossServers(t *testing.T) {
	cfg := gatewayConfig(map[string]config.MCPServer{
		"alpha": {Command: "x"},
		"beta":  {Command: "y"},
	})

	server := buildTestServer(t, cfg, []upstream.ActiveUpstream{
		{Upstream: &stubUpstream{name: "alpha"}, Tools: []*mcp.Tool{testTool("echo")}},
		{Upstream: &stubUpstream{name: "beta"}, Tools: []*mcp.Tool{testTool("echo")}},
	}, testLogger())

	session := connectClient(t, server)
	tools, err := session.ListTools(t.Context(), &mcp.ListToolsParams{})
	if err != nil {
		t.Fatalf("ListTools returned error: %v", err)
	}

	names := map[string]bool{}
	for _, tool := range tools.Tools {
		names[tool.Name] = true
	}
	if !names["alpha_echo"] || !names["beta_echo"] {
		t.Fatalf("tools = %v, want alpha_echo and beta_echo", names)
	}
}

func TestBuildServerAppliesIncludeFilterAndWarnsOnMissing(t *testing.T) {
	cfg := gatewayConfig(map[string]config.MCPServer{
		"alpha": {Command: "x", IncludeTools: []string{"keep", "ghost"}},
	})

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	server := buildTestServer(t, cfg, []upstream.ActiveUpstream{
		{Upstream: &stubUpstream{name: "alpha"}, Tools: []*mcp.Tool{testTool("keep"), testTool("drop")}},
	}, logger)

	session := connectClient(t, server)
	tools, err := session.ListTools(t.Context(), &mcp.ListToolsParams{})
	if err != nil {
		t.Fatalf("ListTools returned error: %v", err)
	}
	if len(tools.Tools) != 1 || tools.Tools[0].Name != "alpha_keep" {
		t.Fatalf("tools = %+v, want only alpha_keep", tools.Tools)
	}
	if !strings.Contains(buf.String(), "ghost") {
		t.Fatal("expected warning about includeTools entry the upstream did not report")
	}
}

func TestBuildServerAppliesExcludeFilter(t *testing.T) {
	cfg := gatewayConfig(map[string]config.MCPServer{
		"alpha": {Command: "x", ExcludeTools: []string{"drop"}},
	})

	server := buildTestServer(t, cfg, []upstream.ActiveUpstream{
		{Upstream: &stubUpstream{name: "alpha"}, Tools: []*mcp.Tool{testTool("keep"), testTool("drop")}},
	}, testLogger())

	session := connectClient(t, server)
	tools, err := session.ListTools(t.Context(), &mcp.ListToolsParams{})
	if err != nil {
		t.Fatalf("ListTools returned error: %v", err)
	}
	if len(tools.Tools) != 1 || tools.Tools[0].Name != "alpha_keep" {
		t.Fatalf("tools = %+v, want only alpha_keep", tools.Tools)
	}
}

func TestProxyPassesUpstreamToolErrorThrough(t *testing.T) {
	ups := &stubUpstream{
		name: "alpha",
		callResult: &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{&mcp.TextContent{Text: "upstream \"alpha\" timed out"}},
		},
	}
	cfg := gatewayConfig(map[string]config.MCPServer{"alpha": {Command: "x"}})

	server := buildTestServer(t, cfg, []upstream.ActiveUpstream{
		{Upstream: ups, Tools: []*mcp.Tool{testTool("echo")}},
	}, testLogger())

	session := connectClient(t, server)
	result, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: "alpha_echo"})
	if err != nil {
		t.Fatalf("CallTool returned protocol error: %v (want tool error)", err)
	}
	if !result.IsError {
		t.Fatal("result.IsError = false, want true")
	}
}

func TestProxyUpstreamGoErrorBecomesProtocolError(t *testing.T) {
	ups := &stubUpstream{name: "alpha", callErr: errors.New("gateway invariant broken")}
	cfg := gatewayConfig(map[string]config.MCPServer{"alpha": {Command: "x"}})

	server := buildTestServer(t, cfg, []upstream.ActiveUpstream{
		{Upstream: ups, Tools: []*mcp.Tool{testTool("echo")}},
	}, testLogger())

	session := connectClient(t, server)
	_, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: "alpha_echo"})
	if err == nil {
		t.Fatal("CallTool returned nil error, want protocol error")
	}
}

func TestUnknownExposedToolIsProtocolError(t *testing.T) {
	cfg := gatewayConfig(map[string]config.MCPServer{"alpha": {Command: "x"}})

	server := buildTestServer(t, cfg, []upstream.ActiveUpstream{
		{Upstream: &stubUpstream{name: "alpha"}, Tools: []*mcp.Tool{testTool("echo")}},
	}, testLogger())

	session := connectClient(t, server)
	_, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: "alpha_nope"})
	if err == nil {
		t.Fatal("CallTool returned nil error, want protocol error for unknown tool")
	}
}

func testLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}
