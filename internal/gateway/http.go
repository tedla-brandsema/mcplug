package gateway

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/tedla-brandsema/mcplug/internal/auth"
)

type HTTPOptions struct {
	Path          string
	Authenticator auth.Authenticator
	Logger        *slog.Logger

	// Tunneled indicates requests reach this loopback listener through a
	// public tunnel (e.g. ngrok), so the Host header legitimately differs
	// from the loopback address and the SDK's DNS-rebinding protection
	// must be disabled.
	Tunneled bool
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
	}, &mcp.StreamableHTTPOptions{
		DisableLocalhostProtection: opts.Tunneled,
	})

	authenticator := opts.Authenticator
	if authenticator == nil {
		authenticator = auth.None()
		logger.Warn(
			"http auth disabled",
			"path", path,
			"warning", "plug HTTP endpoint is unauthenticated",
		)
	}

	mux := http.NewServeMux()

	mux.Handle(path, authenticateHTTP(authenticator, logger, mcpHandler))

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok\n"))
	})

	return mux, nil
}

func authenticateHTTP(authenticator auth.Authenticator, logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		principal, err := authenticator.Authenticate(req.Context(), req)
		if err != nil {
			var unauthorized *auth.UnauthorizedError
			status := http.StatusInternalServerError
			if errors.As(err, &unauthorized) {
				status = http.StatusUnauthorized
			}

			logger.Warn(
				"http auth denied",
				"error", err,
				"remote_addr", req.RemoteAddr,
				"path", req.URL.Path,
			)
			http.Error(w, http.StatusText(status), status)
			return
		}

		_ = principal

		next.ServeHTTP(w, req)
	})
}
