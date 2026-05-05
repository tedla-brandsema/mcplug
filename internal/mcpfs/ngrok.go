package mcpfs

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"golang.ngrok.com/ngrok/v2"
)

type NgrokOptions struct {
	Addr   string
	Path   string
	URL    string
	Logger *slog.Logger
}

func StartNgrok(ctx context.Context, opts NgrokOptions) (string, error) {
	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}

	addr := opts.Addr
	if addr == "" {
		addr = "127.0.0.1:8080"
	}

	upstream := "http://" + addr

	endpointOptions := []ngrok.EndpointOption{}
	if opts.URL != "" {
		endpointOptions = append(endpointOptions, ngrok.WithURL(opts.URL))
	}

	fwd, err := ngrok.Forward(
		ctx,
		ngrok.WithUpstream(upstream),
		endpointOptions...,
	)
	if err != nil {
		return "", fmt.Errorf("start ngrok forwarder: %w", err)
	}

	publicURL := strings.TrimRight(fwd.URL().String(), "/")

	mcpURL := publicURL
	if opts.Path != "" {
		mcpURL += "/" + strings.TrimLeft(opts.Path, "/")
	}

	logger.Info(
		"ngrok tunnel started",
		"public_url", publicURL,
		"mcp_url", mcpURL,
		"upstream", upstream,
	)

	return mcpURL, nil
}