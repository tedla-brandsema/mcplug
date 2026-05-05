package auth

import (
	"context"
	"fmt"
	"net/http"
)

type Config struct {
	Mode     string
	TokenEnv string
}

type Principal struct {
	Subject string
	Email   string
	Issuer  string
	Claims  map[string]any
}

type Authenticator interface {
	Authenticate(ctx context.Context, req *http.Request) (*Principal, error)
}

type UnauthorizedError struct {
	Reason string
}

func (e *UnauthorizedError) Error() string {
	if e.Reason == "" {
		return "unauthorized"
	}
	return "unauthorized: " + e.Reason
}

func New(cfg Config) (Authenticator, error) {
	switch cfg.Mode {
	case "", "none":
		return None(), nil

	case "bearer":
		return NewBearerFromEnv(cfg.TokenEnv)

	default:
		return nil, fmt.Errorf("unsupported auth mode %q", cfg.Mode)
	}
}

func unauthorized(reason string) error {
	return &UnauthorizedError{Reason: reason}
}
