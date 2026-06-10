package mcpfs

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/tedla-brandsema/mcpfs/internal/config"
	"github.com/tedla-brandsema/mcpfs/internal/upstream"
)

// route maps an exposed tool name back to its upstream and original name.
// Keeping the map explicit keeps the naming policy swappable.
type route struct {
	upstream upstream.Upstream
	tool     string
}

// BuildServer assembles the aggregated MCP server from already-started
// upstreams. It consumes the startup listing only: it never creates, starts,
// or lists upstreams itself.
//
// The tool list is a startup snapshot. A future version can refresh it live
// via per-upstream ToolListChangedHandler plus AddTool/RemoveTools on this
// server (HTTP upstreams would then need DisableStandaloneSSE unset).
func BuildServer(cfg config.Config, startup *upstream.StartupResult, logger *slog.Logger) (*Server, error) {
	if logger == nil {
		logger = slog.Default()
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    cfg.Server.Name,
		Version: cfg.Server.Version,
	}, nil)

	routes := make(map[string]route)

	for _, active := range startup.ActiveUpstreams {
		name := active.Upstream.Name()

		srvCfg, ok := cfg.MCPServers[name]
		if !ok {
			return nil, fmt.Errorf("active upstream %q has no config entry", name)
		}

		prefix, err := config.SanitizeServerName(name)
		if err != nil {
			return nil, fmt.Errorf("upstream %q: %w", name, err)
		}

		tools, missing := filterTools(active.Tools, srvCfg.IncludeTools, srvCfg.ExcludeTools)
		for _, m := range missing {
			logger.Warn("includeTools names a tool the upstream did not report", "upstream", name, "tool", m)
		}

		for _, tool := range tools {
			exposed := prefix + "_" + tool.Name
			if existing, ok := routes[exposed]; ok {
				return nil, fmt.Errorf("exposed tool name %q from upstream %q collides with upstream %q", exposed, name, existing.upstream.Name())
			}
			routes[exposed] = route{upstream: active.Upstream, tool: tool.Name}

			clone, err := cloneTool(tool)
			if err != nil {
				return nil, fmt.Errorf("clone tool %q from upstream %q: %w", tool.Name, name, err)
			}
			clone.Name = exposed

			server.AddTool(clone, proxyHandler(logger, active.Upstream, tool.Name))
		}
	}

	logger.Info("aggregated upstream tools", "upstreams", len(startup.ActiveUpstreams), "tools", len(routes))

	return &Server{MCP: server}, nil
}

// cloneTool deep-copies a tool definition via JSON so renaming the clone can
// never mutate or alias upstream-owned schemas, slices, or maps.
func cloneTool(tool *mcp.Tool) (*mcp.Tool, error) {
	data, err := json.Marshal(tool)
	if err != nil {
		return nil, err
	}
	var clone mcp.Tool
	if err := json.Unmarshal(data, &clone); err != nil {
		return nil, err
	}
	return &clone, nil
}

// proxyHandler forwards a tool call to its upstream. Per the upstream
// CallTool contract, expected upstream failures arrive as tool-error results
// and pass through; a Go error from the upstream is a gateway invariant
// violation and surfaces as a protocol error. A malformed client argument
// payload is also a protocol error.
func proxyHandler(logger *slog.Logger, ups upstream.Upstream, tool string) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args map[string]any
		if raw := req.Params.Arguments; len(raw) > 0 {
			if err := json.Unmarshal(raw, &args); err != nil {
				return nil, fmt.Errorf("malformed tool arguments: %w", err)
			}
		}

		start := time.Now()
		result, err := ups.CallTool(ctx, tool, args)
		elapsed := time.Since(start)

		category := "ok"
		switch {
		case err != nil:
			category = "internal_error"
		case result.IsError:
			category = "tool_error"
		}

		logger.Info("proxied tool call",
			"upstream", ups.Name(),
			"tool", tool,
			"duration_ms", elapsed.Milliseconds(),
			"category", category,
		)

		if err != nil {
			return nil, fmt.Errorf("upstream %q: %w", ups.Name(), err)
		}
		return result, nil
	}
}
