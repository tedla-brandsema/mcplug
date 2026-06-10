// Package upstream connects MCPlug to the MCP servers configured under
// mcpServers: stdio child processes and remote streamable-HTTP endpoints.
package upstream

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/tedla-brandsema/mcplug/internal/config"
)

// DefaultTimeout bounds upstream connect, list, and call operations.
const DefaultTimeout = 60 * time.Second

// timeout is variable so tests can shorten it.
var timeout = DefaultTimeout

// Upstream is a connected MCP server whose tools MCPlug aggregates.
//
// CallTool follows the gateway error-mapping contract: expected
// upstream/runtime failures (upstream restarting or unavailable, timeout,
// child process exited, HTTP upstream unreachable, upstream-reported tool
// error) return (*mcp.CallToolResult{IsError: true}, nil). Go errors are
// reserved for gateway-side invariants, malformed local state, or impossible
// internal conditions.
type Upstream interface {
	Name() string
	Connect(ctx context.Context) error
	ListTools(ctx context.Context) ([]*mcp.Tool, error)
	CallTool(ctx context.Context, tool string, args map[string]any) (*mcp.CallToolResult, error)
	Close() error
}

// New builds an upstream for one mcpServers entry: url entries become HTTP
// upstreams, command entries become stdio upstreams.
func New(name string, cfg config.MCPServer, logger *slog.Logger) Upstream {
	if logger == nil {
		logger = slog.Default()
	}
	if cfg.IsHTTP() {
		return newHTTPUpstream(name, cfg, logger)
	}
	// stdio children get restart supervision; HTTP upstreams are stateless
	// per-call sessions and need none.
	return newSupervisor(name, cfg, logger)
}

func clientImplementation() *mcp.Implementation {
	return &mcp.Implementation{
		Name:    "mcplug",
		Version: "v2",
	}
}

// toolError wraps an expected upstream failure as a tool error result, per
// the CallTool contract.
func toolError(format string, args ...any) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(format, args...)},
		},
	}
}

func withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, timeout)
}
