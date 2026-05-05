package mcpfs

import (
	"crypto/subtle"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type HTTPOptions struct {
	Path         string
	AuthTokenEnv string
	RequireAuth  bool
	Logger       *slog.Logger
}

func (s *Server) HTTPHandler(opts HTTPOptions) (http.Handler, error) {
	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}

	path := opts.Path
	if path == "" {
		path = "/mcp"
	}

	mcpHandler := mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server {
		return s.MCP
	}, nil)

	var protectedMCPHandler http.Handler = mcpHandler

	if opts.RequireAuth {
		if opts.AuthTokenEnv == "" {
			return nil, fmt.Errorf("auth token env var name is required")
		}

		token := os.Getenv(opts.AuthTokenEnv)
		if token == "" {
			return nil, fmt.Errorf("auth token env var %s is empty or unset", opts.AuthTokenEnv)
		}

		protectedMCPHandler = bearerAuth(token, logger, mcpHandler)
	} else {
		logger.Warn(
			"http auth disabled",
			"path", path,
			"warning", "mcpfs HTTP endpoint is unauthenticated",
		)
	}

	mux := http.NewServeMux()

	mux.Handle(path, protectedMCPHandler)

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok\n"))
	})

	return mux, nil
}

func bearerAuth(expectedToken string, logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		got := req.Header.Get("Authorization")
		const prefix = "Bearer "

		if !strings.HasPrefix(got, prefix) {
			logger.Warn(
				"http auth denied",
				"reason", "missing bearer token",
				"remote_addr", req.RemoteAddr,
				"path", req.URL.Path,
			)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		gotToken := strings.TrimPrefix(got, prefix)
		if subtle.ConstantTimeCompare([]byte(gotToken), []byte(expectedToken)) != 1 {
			logger.Warn(
				"http auth denied",
				"reason", "invalid bearer token",
				"remote_addr", req.RemoteAddr,
				"path", req.URL.Path,
			)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, req)
	})
}