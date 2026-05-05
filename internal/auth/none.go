package auth

import (
	"context"
	"net/http"
)

type NoneAuthenticator struct{}

func None() NoneAuthenticator {
	return NoneAuthenticator{}
}

func (NoneAuthenticator) Authenticate(ctx context.Context, req *http.Request) (*Principal, error) {
	_ = ctx
	_ = req

	return &Principal{}, nil
}
