package upstream

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/tedla-brandsema/mcpfs/internal/config"
)

type echoArgs struct {
	Text string `json:"text"`
}

type echoResult struct {
	Echo string `json:"echo"`
}

func newFakeUpstreamServer() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "fake-upstream", Version: "test"}, nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "echo",
		Description: "echoes text back",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args echoArgs) (*mcp.CallToolResult, echoResult, error) {
		return nil, echoResult{Echo: args.Text}, nil
	})
	return server
}

// startFakeHTTPUpstream serves a real MCP server over streamable HTTP and
// records the headers of every request it receives.
func startFakeHTTPUpstream(t *testing.T) (*httptest.Server, *headerRecorder) {
	t.Helper()

	server := newFakeUpstreamServer()
	handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, nil)

	rec := &headerRecorder{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec.record(r.Header)
		handler.ServeHTTP(w, r)
	}))
	t.Cleanup(ts.Close)
	return ts, rec
}

type headerRecorder struct {
	mu      sync.Mutex
	headers []http.Header
}

func (r *headerRecorder) record(h http.Header) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.headers = append(r.headers, h.Clone())
}

func (r *headerRecorder) all() []http.Header {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.headers
}

func TestHTTPUpstreamListAndCall(t *testing.T) {
	ts, _ := startFakeHTTPUpstream(t)

	u := New("remote", config.MCPServer{URL: ts.URL}, testLogger(t))
	ctx := context.Background()

	if err := u.Connect(ctx); err != nil {
		t.Fatalf("Connect returned error: %v", err)
	}

	tools, err := u.ListTools(ctx)
	if err != nil {
		t.Fatalf("ListTools returned error: %v", err)
	}
	if len(tools) != 1 || tools[0].Name != "echo" {
		t.Fatalf("tools = %+v, want one tool named echo", tools)
	}

	result, err := u.CallTool(ctx, "echo", map[string]any{"text": "hi"})
	if err != nil {
		t.Fatalf("CallTool returned error: %v", err)
	}
	if result.IsError {
		t.Fatalf("CallTool result is an error: %+v", result.Content)
	}
}

func TestHTTPUpstreamInjectsConfiguredHeaders(t *testing.T) {
	ts, rec := startFakeHTTPUpstream(t)

	u := New("remote", config.MCPServer{
		URL:     ts.URL,
		Headers: map[string]string{"X-Api-Key": "sekrit", "X-Custom": "yes"},
	}, testLogger(t))

	if _, err := u.ListTools(context.Background()); err != nil {
		t.Fatalf("ListTools returned error: %v", err)
	}

	headers := rec.all()
	if len(headers) == 0 {
		t.Fatal("no requests recorded")
	}
	for _, h := range headers {
		if h.Get("X-Api-Key") != "sekrit" {
			t.Fatalf("X-Api-Key = %q, want sekrit", h.Get("X-Api-Key"))
		}
		if h.Get("X-Custom") != "yes" {
			t.Fatalf("X-Custom = %q, want yes", h.Get("X-Custom"))
		}
	}
}

func TestHTTPUpstreamUnreachableCallReturnsToolError(t *testing.T) {
	u := New("gone", config.MCPServer{URL: "http://127.0.0.1:1/mcp"}, testLogger(t))

	result, err := u.CallTool(context.Background(), "echo", nil)
	if err != nil {
		t.Fatalf("CallTool returned protocol error: %v (want tool error)", err)
	}
	if !result.IsError {
		t.Fatal("CallTool result.IsError = false, want true")
	}
}

func TestHTTPUpstreamTimeoutReturnsToolError(t *testing.T) {
	hang := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	t.Cleanup(hang.Close)

	old := timeout
	timeout = 200 * time.Millisecond
	t.Cleanup(func() { timeout = old })

	u := New("slow", config.MCPServer{URL: hang.URL}, testLogger(t))

	result, err := u.CallTool(context.Background(), "echo", nil)
	if err != nil {
		t.Fatalf("CallTool returned protocol error: %v (want tool error)", err)
	}
	if !result.IsError {
		t.Fatal("CallTool result.IsError = false, want true")
	}
	if text := resultText(result); text == "" {
		t.Fatal("tool error has no text content")
	}
}

func resultText(result *mcp.CallToolResult) string {
	for _, c := range result.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			return tc.Text
		}
	}
	return ""
}
