package upstream

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/tedla-brandsema/mcpfs/internal/config"
)

// httpUpstream talks to a remote streamable-HTTP MCP server. Sessions are
// opened per call: simple, stateless, and resilient to upstream restarts.
type httpUpstream struct {
	name   string
	cfg    config.MCPServer
	client *http.Client
	logger *slog.Logger
}

func newHTTPUpstream(name string, cfg config.MCPServer, logger *slog.Logger) *httpUpstream {
	return &httpUpstream{
		name:   name,
		cfg:    cfg,
		client: &http.Client{Transport: &headerTransport{headers: cfg.Headers}},
		logger: logger,
	}
}

func (u *httpUpstream) Name() string { return u.name }

// Connect probes the upstream by listing its tools once.
func (u *httpUpstream) Connect(ctx context.Context) error {
	_, err := u.ListTools(ctx)
	return err
}

func (u *httpUpstream) ListTools(ctx context.Context) ([]*mcp.Tool, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	session, err := u.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer session.Close()

	result, err := session.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		return nil, err
	}
	return result.Tools, nil
}

func (u *httpUpstream) CallTool(ctx context.Context, tool string, args map[string]any) (*mcp.CallToolResult, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	session, err := u.connect(ctx)
	if err != nil {
		return callFailure(u.name, "unreachable", err), nil
	}
	defer session.Close()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      tool,
		Arguments: args,
	})
	if err != nil {
		return callFailure(u.name, "call failed", err), nil
	}
	return result, nil
}

func (u *httpUpstream) Close() error { return nil }

func (u *httpUpstream) connect(ctx context.Context) (*mcp.ClientSession, error) {
	return mcp.NewClient(clientImplementation(), nil).Connect(ctx, &mcp.StreamableClientTransport{
		Endpoint:             u.cfg.URL,
		HTTPClient:           u.client,
		DisableStandaloneSSE: true,
	}, nil)
}

// callFailure classifies an expected upstream call failure as a tool error.
// The error text never includes header or env values.
func callFailure(name, category string, err error) *mcp.CallToolResult {
	if errors.Is(err, context.DeadlineExceeded) {
		category = "timed out"
	}
	return toolError("upstream %q %s: %v", name, category, err)
}

// headerTransport injects configured per-server headers into every request.
type headerTransport struct {
	base    http.RoundTripper
	headers map[string]string
}

func (t *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	base := t.base
	if base == nil {
		base = http.DefaultTransport
	}
	if len(t.headers) == 0 {
		return base.RoundTrip(req)
	}

	clone := req.Clone(req.Context())
	for k, v := range t.headers {
		clone.Header.Set(k, v)
	}
	return base.RoundTrip(clone)
}
