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
	"github.com/tedla-brandsema/mcpfs/internal/upstream"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{}))

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "init":
			os.Exit(runInit(os.Args[2:], logger))
		}
	}

	var configPath string
	flag.StringVar(&configPath, "config", "", "path to mcpfs config file; defaults to the global user config")
	flag.Parse()

	cfg, resolvedConfigPath, err := config.LoadOrCreate(configPath)
	if err != nil {
		logger.Error("load config", "error", err)
		os.Exit(1)
	}

	logger.Info("loaded config", "path", resolvedConfigPath)

	if warning := config.WorldReadableWarning(resolvedConfigPath, cfg); warning != "" {
		logger.Warn(warning)
	}
	if len(cfg.MCPServers) == 0 {
		logger.Warn("no mcpServers configured; the gateway will expose zero tools")
	}

	ctx := context.Background()

	startup, err := upstream.StartAll(ctx, cfg, logger)
	if err != nil {
		logger.Error("start upstreams", "error", err)
		os.Exit(1)
	}
	defer startup.Close()

	server, err := mcpfs.BuildServer(cfg, startup, logger)
	if err != nil {
		logger.Error("create server", "error", err)
		os.Exit(1)
	}

	switch cfg.Server.Transport {
	case "stdio":
		logger.Info("starting mcpfs", "transport", "stdio")

		if err := server.MCP.Run(ctx, &mcp.StdioTransport{}); err != nil {
			logger.Error("run stdio server", "error", err)
			os.Exit(1)
		}

	case "http", "http_ngrok":
		authenticator, err := auth.New(auth.Config{
			Mode:            string(cfg.Server.Auth.Mode),
			TokenEnv:        cfg.Server.Auth.TokenEnv,
			Issuer:          cfg.Server.Auth.Issuer,
			Audience:        cfg.Server.Auth.Audience,
			JWKSURL:         cfg.Server.Auth.JWKSURL,
			AllowedEmails:   cfg.Server.Auth.AllowedEmails,
			AllowedSubjects: cfg.Server.Auth.AllowedSubjects,
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
