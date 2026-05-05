package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/tedla-brandsema/mcpfs/internal/auth"
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

	server, err := mcpfs.NewServer(cfg, logger)
	if err != nil {
		logger.Error("create server", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()

	switch cfg.Server.Transport {
	case "stdio":
		logger.Info("starting mcpfs", "transport", "stdio", "roots", len(cfg.Roots))

		if err := server.MCP.Run(ctx, &mcp.StdioTransport{}); err != nil {
			logger.Error("run stdio server", "error", err)
			os.Exit(1)
		}

	case "http", "http_ngrok":
		authenticator, err := auth.New(auth.Config{
			Mode:     string(cfg.Server.Auth.Mode),
			TokenEnv: cfg.Server.Auth.TokenEnv,
		})
		if err != nil {
			logger.Error("create authenticator", "error", err)
			os.Exit(1)
		}

		handler, err := server.HTTPHandler(mcpfs.HTTPOptions{
			Path:          cfg.Server.Path,
			Authenticator: authenticator,
			Logger:        logger,
		})
		if err != nil {
			logger.Error("create http handler", "error", err)
			os.Exit(1)
		}

		httpServer := &http.Server{
			Addr:    cfg.Server.Addr,
			Handler: handler,
		}

		if cfg.Server.Transport == "http_ngrok" {
			mcpURL, err := mcpfs.StartNgrok(ctx, mcpfs.NgrokOptions{
				Addr:   cfg.Server.Addr,
				Path:   cfg.Server.Path,
				URL:    cfg.Server.NgrokURL,
				Logger: logger,
			})
			if err != nil {
				logger.Error("start ngrok", "error", err)
				os.Exit(1)
			}

			logger.Info("mcpfs public endpoint ready", "mcp_url", mcpURL)
		}

		logger.Info(
			"starting mcpfs",
			"transport", cfg.Server.Transport,
			"addr", cfg.Server.Addr,
			"path", cfg.Server.Path,
			"auth_mode", cfg.Server.Auth.Mode,
			"roots", len(cfg.Roots),
		)

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("run http server", "error", err)
			os.Exit(1)
		}

	default:
		logger.Error("unsupported transport", "transport", cfg.Server.Transport)
		os.Exit(1)
	}
}
