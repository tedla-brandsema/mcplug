package auth

import (
	"context"
	"crypto/subtle"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type BearerAuthenticator struct {
	token string
}

func NewBearerFromEnv(envName string) (*BearerAuthenticator, error) {
	if envName == "" {
		return nil, fmt.Errorf("auth token env var name is required")
	}

	token := os.Getenv(envName)
	if token == "" {
		return nil, fmt.Errorf("auth token env var %s is empty or unset", envName)
	}

	return NewBearer(token), nil
}

func NewBearer(token string) *BearerAuthenticator {
	return &BearerAuthenticator{token: token}
}

func (a *BearerAuthenticator) Authenticate(ctx context.Context, req *http.Request) (*Principal, error) {
	_ = ctx

	got := req.Header.Get("Authorization")
	const prefix = "Bearer "

	if !strings.HasPrefix(got, prefix) {
		return nil, unauthorized("missing bearer token")
	}

	gotToken := strings.TrimPrefix(got, prefix)
	if subtle.ConstantTimeCompare([]byte(gotToken), []byte(a.token)) != 1 {
		return nil, unauthorized("invalid bearer token")
	}

	return &Principal{
		Subject: "bearer",
	}, nil
}
