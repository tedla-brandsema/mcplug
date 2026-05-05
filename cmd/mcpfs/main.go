package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/tedla-brandsema/mcpfs/internal/config"
	"github.com/tedla-brandsema/mcpfs/internal/mcpfs"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.json", "path to mcpfs config file")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{}))

	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Error("load config", "error", err)
		os.Exit(1)
	}

	svc, err := mcpfs.NewService(cfg, logger)
	if err != nil {
		logger.Error("create service", "error", err)
		os.Exit(1)
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    cfg.Server.Name,
		Version: cfg.Server.Version,
	}, nil)

	mcpfs.RegisterTools(server, svc)

	switch cfg.Server.Transport {
	case "stdio":
		logger.Info("starting mcpfs", "transport", "stdio", "roots", len(cfg.Roots))

		if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
			logger.Error("run server", "error", err)
			os.Exit(1)
		}

	default:
		logger.Error("unsupported transport", "transport", cfg.Server.Transport)
		os.Exit(1)
	}
}