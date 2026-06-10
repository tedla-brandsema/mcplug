package gateway

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/tedla-brandsema/mcplug/internal/config"
	"github.com/tedla-brandsema/mcplug/internal/upstream"
)

// TestEndToEndHTTPUpstream exercises the full gateway path with a real HTTP
// upstream: StartAll -> BuildServer -> in-memory MCP client call.
func TestEndToEndHTTPUpstream(t *testing.T) {
	upstreamServer := mcp.NewServer(&mcp.Implementation{Name: "fake-upstream", Version: "test"}, nil)
	mcp.AddTool(upstreamServer, &mcp.Tool{
		Name:        "greet",
		Description: "greets the caller",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		Name string `json:"name"`
	}) (*mcp.CallToolResult, struct {
		Greeting string `json:"greeting"`
	}, error) {
		return nil, struct {
			Greeting string `json:"greeting"`
		}{Greeting: "hello " + args.Name}, nil
	})

	handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return upstreamServer }, nil)
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)

	cfg := gatewayConfig(map[string]config.MCPServer{
		"remote": {URL: ts.URL},
	})

	startup, err := upstream.StartAll(t.Context(), cfg, testLogger())
	if err != nil {
		t.Fatalf("StartAll returned error: %v", err)
	}
	t.Cleanup(startup.Close)

	server, err := BuildServer(cfg, startup, testLogger())
	if err != nil {
		t.Fatalf("BuildServer returned error: %v", err)
	}

	session := connectClient(t, server)

	tools, err := session.ListTools(t.Context(), &mcp.ListToolsParams{})
	if err != nil {
		t.Fatalf("ListTools returned error: %v", err)
	}
	if len(tools.Tools) != 1 || tools.Tools[0].Name != "remote_greet" {
		t.Fatalf("tools = %+v, want one tool remote_greet", tools.Tools)
	}

	result, err := session.CallTool(t.Context(), &mcp.CallToolParams{
		Name:      "remote_greet",
		Arguments: map[string]any{"name": "ted"},
	})
	if err != nil {
		t.Fatalf("CallTool returned error: %v", err)
	}
	if result.IsError {
		t.Fatalf("CallTool result is an error: %+v", result.Content)
	}
}
