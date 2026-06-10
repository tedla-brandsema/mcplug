package upstream

import (
	"context"
	"fmt"
	"log/slog"
	"sort"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/tedla-brandsema/mcplug/internal/config"
)

// ActiveUpstream is a started upstream together with its startup tool
// listing. Listing happens exactly once, here; the aggregator consumes this
// snapshot and never lists on its own.
type ActiveUpstream struct {
	Upstream Upstream
	Tools    []*mcp.Tool
}

type SkippedUpstream struct {
	Name string
	Err  error
}

type StartupResult struct {
	ActiveUpstreams []ActiveUpstream
	SkippedOptional []SkippedUpstream
}

// Close closes every active upstream.
func (r *StartupResult) Close() {
	for _, a := range r.ActiveUpstreams {
		_ = a.Upstream.Close()
	}
}

// newUpstream is variable so tests can substitute stub upstreams.
var newUpstream = New

// StartAll is the single owner of upstream startup: it starts every enabled
// mcpServers entry and performs the initial tool listing. A required upstream
// failing aborts startup (already-started upstreams are closed); an optional
// upstream failing is logged and skipped — its tools stay absent until MCPlug
// restarts. Disabled entries are ignored entirely.
func StartAll(ctx context.Context, cfg config.Config, logger *slog.Logger) (*StartupResult, error) {
	if logger == nil {
		logger = slog.Default()
	}

	names := make([]string, 0, len(cfg.MCPServers))
	for name, srv := range cfg.MCPServers {
		if srv.Disabled {
			logger.Info("upstream disabled", "upstream", name)
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)

	result := &StartupResult{}
	for _, name := range names {
		srvCfg := cfg.MCPServers[name]

		active, err := startOne(ctx, name, srvCfg, logger)
		if err != nil {
			if srvCfg.Optional {
				logger.Warn("optional upstream failed at startup, skipping", "upstream", name, "error", err)
				result.SkippedOptional = append(result.SkippedOptional, SkippedUpstream{Name: name, Err: err})
				continue
			}
			result.Close()
			return nil, fmt.Errorf("required upstream %q failed at startup: %w", name, err)
		}

		logger.Info("upstream started", "upstream", name, "tools", len(active.Tools))
		result.ActiveUpstreams = append(result.ActiveUpstreams, active)
	}

	return result, nil
}

func startOne(ctx context.Context, name string, srvCfg config.MCPServer, logger *slog.Logger) (ActiveUpstream, error) {
	ups := newUpstream(name, srvCfg, logger)

	if err := ups.Connect(ctx); err != nil {
		_ = ups.Close()
		return ActiveUpstream{}, err
	}

	tools, err := ups.ListTools(ctx)
	if err != nil {
		_ = ups.Close()
		return ActiveUpstream{}, fmt.Errorf("list tools: %w", err)
	}

	return ActiveUpstream{Upstream: ups, Tools: tools}, nil
}
