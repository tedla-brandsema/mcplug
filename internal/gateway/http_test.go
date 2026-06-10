package gateway

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tedla-brandsema/mcplug/internal/auth"
)

func TestAuthenticateHTTPAllowsAuthorizedRequest(t *testing.T) {
	authenticator := staticAuthenticator{
		principal: &auth.Principal{
			Subject: "user-123",
		},
	}

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusNoContent)
	})

	handler := authenticateHTTP(authenticator, discardLogger(), next)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if !nextCalled {
		t.Fatal("next handler was not called")
	}
}

func TestAuthenticateHTTPMapsUnauthorizedErrorTo401(t *testing.T) {
	authenticator := staticAuthenticator{
		err: &auth.UnauthorizedError{Reason: "invalid token"},
	}

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusNoContent)
	})

	handler := authenticateHTTP(authenticator, discardLogger(), next)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if nextCalled {
		t.Fatal("next handler was called")
	}
}

func TestAuthenticateHTTPMapsBackendErrorTo500(t *testing.T) {
	authenticator := staticAuthenticator{
		err: errors.New("load jwks: backend unavailable"),
	}

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusNoContent)
	})

	handler := authenticateHTTP(authenticator, discardLogger(), next)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	if nextCalled {
		t.Fatal("next handler was called")
	}
}

func TestHTTPHandlerHealthzIsUnauthenticated(t *testing.T) {
	server := &Server{}

	authenticator := staticAuthenticator{
		err: &auth.UnauthorizedError{Reason: "should not be called for healthz"},
	}

	handler, err := server.HTTPHandler(HTTPOptions{
		Path:          "/mcp",
		Authenticator: authenticator,
		Logger:        discardLogger(),
	})
	if err != nil {
		t.Fatalf("HTTPHandler returned error: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != "ok\n" {
		t.Fatalf("body = %q, want %q", rec.Body.String(), "ok\n")
	}
}

type staticAuthenticator struct {
	principal *auth.Principal
	err       error
}

func (a staticAuthenticator) Authenticate(ctx context.Context, req *http.Request) (*auth.Principal, error) {
	_ = ctx
	_ = req

	if a.err != nil {
		return nil, a.err
	}
	if a.principal != nil {
		return a.principal, nil
	}
	return &auth.Principal{}, nil
}

func discardLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}
