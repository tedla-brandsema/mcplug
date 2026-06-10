package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"sort"

	"github.com/tedla-brandsema/mcplug/internal/config"
	"github.com/tedla-brandsema/mcplug/internal/gateway"
	"github.com/tedla-brandsema/mcplug/internal/upstream"
)

// runLs is the config smoke test: it validates the config and probes every
// mcpServers entry, printing its tools. It never starts the MCPlug transport,
// HTTP listener, or ngrok tunnel. A required upstream failure exits non-zero;
// optional failures are shown but do not fail the command.
func runLs(args []string, logger *slog.Logger) int {
	fs := flag.NewFlagSet("ls", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var configPath string
	fs.StringVar(&configPath, "config", "", "path to config file; defaults to the global user config")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}

	cfg, resolvedConfigPath, err := config.LoadOrCreate(configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	fmt.Printf("config: %s\n", resolvedConfigPath)
	if len(cfg.MCPServers) == 0 {
		fmt.Println("no mcpServers configured")
		return 0
	}

	names := make([]string, 0, len(cfg.MCPServers))
	for name := range cfg.MCPServers {
		names = append(names, name)
	}
	sort.Strings(names)

	requiredFailure := false
	for _, name := range names {
		srvCfg := cfg.MCPServers[name]
		if !lsServer(name, srvCfg, logger) && !srvCfg.Optional {
			requiredFailure = true
		}
	}

	if requiredFailure {
		return 1
	}
	return 0
}

// lsServer probes one entry and reports whether it is healthy (disabled
// entries count as healthy).
func lsServer(name string, srvCfg config.MCPServer, logger *slog.Logger) bool {
	kind := "stdio"
	if srvCfg.IsHTTP() {
		kind = "http"
	}

	if srvCfg.Disabled {
		fmt.Printf("\n%s (%s) — disabled\n", name, kind)
		return true
	}

	ups := upstream.New(name, srvCfg, logger)
	defer ups.Close()

	ctx := context.Background()
	if err := ups.Connect(ctx); err != nil {
		printFailure(name, kind, srvCfg.Optional, err)
		return false
	}

	tools, err := ups.ListTools(ctx)
	if err != nil {
		printFailure(name, kind, srvCfg.Optional, err)
		return false
	}

	prefix, err := config.SanitizeServerName(name)
	if err != nil {
		printFailure(name, kind, srvCfg.Optional, err)
		return false
	}

	kept, missing := gateway.FilterTools(tools, srvCfg.IncludeTools, srvCfg.ExcludeTools)
	keptNames := make(map[string]struct{}, len(kept))
	for _, t := range kept {
		keptNames[t.Name] = struct{}{}
	}

	fmt.Printf("\n%s (%s) — running, %d tool(s)\n", name, kind, len(kept))
	for _, t := range tools {
		if _, ok := keptNames[t.Name]; ok {
			fmt.Printf("  %s_%s  (%s)\n", prefix, t.Name, t.Name)
		} else {
			fmt.Printf("  -          (%s) [filtered out]\n", t.Name)
		}
	}
	for _, m := range missing {
		fmt.Printf("  !          (%s) [in includeTools but not reported by upstream]\n", m)
	}

	return true
}

func printFailure(name, kind string, optional bool, err error) {
	status := "failed"
	if optional {
		status = "failed (optional, skipped)"
	}
	fmt.Printf("\n%s (%s) — %s: %v\n", name, kind, status, err)
}
