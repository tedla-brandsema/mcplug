package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwt"
)

type OIDCConfig struct {
	Issuer          string
	Audience        string
	JWKSURL         string
	AllowedEmails   []string
	AllowedSubjects []string
}

type OIDCAuthenticator struct {
	issuer          string
	audience        string
	allowedEmails   map[string]struct{}
	allowedSubjects map[string]struct{}
	jwks            *CachingJWKS
	now             func() time.Time
}

func NewOIDC(cfg OIDCConfig) (*OIDCAuthenticator, error) {
	if cfg.Issuer == "" {
		return nil, fmt.Errorf("oidc issuer is required")
	}
	if cfg.Audience == "" {
		return nil, fmt.Errorf("oidc audience is required")
	}
	if cfg.JWKSURL == "" {
		return nil, fmt.Errorf("oidc jwks url is required")
	}
	if len(cfg.AllowedEmails) == 0 && len(cfg.AllowedSubjects) == 0 {
		return nil, fmt.Errorf("oidc allowed emails or allowed subjects is required")
	}

	return &OIDCAuthenticator{
		issuer:          cfg.Issuer,
		audience:        cfg.Audience,
		allowedEmails:   stringSet(cfg.AllowedEmails),
		allowedSubjects: stringSet(cfg.AllowedSubjects),
		jwks:            NewCachingJWKS(cfg.JWKSURL),
		now:             time.Now,
	}, nil
}

func (a *OIDCAuthenticator) Authenticate(ctx context.Context, req *http.Request) (*Principal, error) {
	raw, err := bearerToken(req)
	if err != nil {
		return nil, err
	}

	keySet, err := a.jwks.KeySet(ctx)
	if err != nil {
		return nil, fmt.Errorf("load jwks: %w", err)
	}

	token, err := jwt.Parse(
		[]byte(raw),
		jwt.WithKeySet(keySet),
		jwt.WithValidate(false),
	)
	if err != nil {
		return nil, unauthorized("invalid oidc token")
	}

	if err := a.validateToken(token); err != nil {
		return nil, err
	}

	subject, _ := token.Subject()
	issuer, _ := token.Issuer()
	email, _ := claimString(token, "email")

	return &Principal{
		Subject: subject,
		Email:   email,
		Issuer:  issuer,
		Claims: map[string]any{
			"sub":   subject,
			"iss":   issuer,
			"email": email,
		},
	}, nil
}

func (a *OIDCAuthenticator) validateToken(token jwt.Token) error {
	now := a.now()

	issuer, ok := token.Issuer()
	if !ok || issuer != a.issuer {
		return unauthorized("invalid issuer")
	}

	subject, ok := token.Subject()
	if !ok || subject == "" {
		return unauthorized("missing subject")
	}

	if !hasAudience(token, a.audience) {
		return unauthorized("invalid audience")
	}

	expiration, ok := token.Expiration()
	if !ok || expiration.IsZero() {
		return unauthorized("missing expiration")
	}
	if !now.Before(expiration) {
		return unauthorized("token expired")
	}

	notBefore, ok := token.NotBefore()
	if ok && !notBefore.IsZero() && now.Before(notBefore) {
		return unauthorized("token not valid yet")
	}

	email, _ := claimString(token, "email")
	if !a.identityAllowed(subject, email) {
		return unauthorized("identity not allowed")
	}

	return nil
}

func hasAudience(token jwt.Token, audience string) bool {
	audiences, ok := token.Audience()
	if !ok {
		return false
	}

	for _, aud := range audiences {
		if aud == audience {
			return true
		}
	}
	return false
}

func claimString(token jwt.Token, name string) (string, bool) {
	var value string
	if err := token.Get(name, &value); err != nil {
		return "", false
	}
	return value, true
}

func (a *OIDCAuthenticator) identityAllowed(subject, email string) bool {
	if _, ok := a.allowedSubjects[subject]; ok {
		return true
	}
	if email != "" {
		if _, ok := a.allowedEmails[email]; ok {
			return true
		}
	}
	return false
}

func bearerToken(req *http.Request) (string, error) {
	got := req.Header.Get("Authorization")
	const prefix = "Bearer "

	if !strings.HasPrefix(got, prefix) {
		return "", unauthorized("missing bearer token")
	}

	token := strings.TrimSpace(strings.TrimPrefix(got, prefix))
	if token == "" {
		return "", unauthorized("missing bearer token")
	}

	return token, nil
}

func stringSet(values []string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		set[value] = struct{}{}
	}
	return set
}
