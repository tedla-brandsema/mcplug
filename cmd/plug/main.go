package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/tedla-brandsema/mcplug/internal/auth"
	"github.com/tedla-brandsema/mcplug/internal/config"
	"github.com/tedla-brandsema/mcplug/internal/gateway"
	"github.com/tedla-brandsema/mcplug/internal/upstream"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{}))

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "init":
			os.Exit(runInit(os.Args[2:], logger))
		case "ls":
			os.Exit(runLs(os.Args[2:], logger))
		case "help", "-h", "--help":
			printUsage()
			os.Exit(0)
		}
	}

	var configPath string
	flag.Usage = func() {
		printUsage()
		flag.PrintDefaults()
	}
	flag.StringVar(&configPath, "config", "", "path to config file; defaults to the global user config")
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	startup, err := upstream.StartAll(ctx, cfg, logger)
	if err != nil {
		logger.Error("start upstreams", "error", err)
		os.Exit(1)
	}
	defer startup.Close()

	for _, skipped := range startup.SkippedOptional {
		logger.Warn("optional upstream skipped; restart plug after fixing it",
			"upstream", skipped.Name, "error", skipped.Err)
	}

	server, err := gateway.BuildServer(cfg, startup, logger)
	if err != nil {
		logger.Error("create server", "error", err)
		os.Exit(1)
	}

	switch cfg.Server.Transport {
	case "stdio":
		logger.Info("starting plug", "transport", "stdio")

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

		handler, err := server.HTTPHandler(gateway.HTTPOptions{
			Path:          cfg.Server.Path,
			Authenticator: authenticator,
			Logger:        logger,
			Tunneled:      cfg.Server.Transport == "http_ngrok",
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
			mcpURL, err := gateway.StartNgrok(ctx, gateway.NgrokOptions{
				Addr:   cfg.Server.Addr,
				Path:   cfg.Server.Path,
				URL:    cfg.Server.NgrokURL,
				Logger: logger,
			})
			if err != nil {
				logger.Error("start ngrok", "error", err)
				os.Exit(1)
			}

			logger.Info("plug public endpoint ready", "mcp_url", mcpURL)
		}

		logger.Info(
			"starting plug",
			"transport", cfg.Server.Transport,
			"addr", cfg.Server.Addr,
			"path", cfg.Server.Path,
			"auth_mode", cfg.Server.Auth.Mode,
		)

		// Shut the listener down before the deferred startup.Close()
		// terminates the upstream children.
		go func() {
			<-ctx.Done()
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_ = httpServer.Shutdown(shutdownCtx)
		}()

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("run http server", "error", err)
			os.Exit(1)
		}

	default:
		logger.Error("unsupported transport", "transport", cfg.Server.Transport)
		os.Exit(1)
	}
}

func printUsage() {
	os.Stderr.WriteString(`plug - MCPlug, an MCP aggregating gateway

Usage:
  plug [-config path]   run the gateway with the configured transport
  plug init [-path p]   write a starter config with example mcpServers
  plug ls [-config p]   probe configured mcpServers and list their tools
  plug help             show this help

`)
}
