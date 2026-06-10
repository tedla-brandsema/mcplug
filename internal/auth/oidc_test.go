package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

const (
	testIssuer   = "https://issuer.example.com"
	testAudience = "mcplug"
	testSubject  = "user-123"
	testEmail    = "you@example.com"
	testKeyID    = "test-key"
)

func TestOIDCAuthenticateValidToken(t *testing.T) {
	env := newOIDCTestEnv(t)

	authenticator := env.authenticator(t, OIDCConfig{
		Issuer:        testIssuer,
		Audience:      testAudience,
		JWKSURL:       env.jwksURL,
		AllowedEmails: []string{testEmail},
	})

	token := env.signToken(t, tokenClaims{
		Issuer:     testIssuer,
		Subject:    testSubject,
		Audience:   testAudience,
		Email:      testEmail,
		Expiration: env.now.Add(time.Hour),
	})

	principal, err := authenticator.Authenticate(context.Background(), bearerRequest(token))
	if err != nil {
		t.Fatalf("Authenticate returned error: %v", err)
	}

	if principal.Subject != testSubject {
		t.Fatalf("Subject = %q, want %q", principal.Subject, testSubject)
	}
	if principal.Email != testEmail {
		t.Fatalf("Email = %q, want %q", principal.Email, testEmail)
	}
	if principal.Issuer != testIssuer {
		t.Fatalf("Issuer = %q, want %q", principal.Issuer, testIssuer)
	}
}

func TestOIDCAuthenticateMissingBearerToken(t *testing.T) {
	env := newOIDCTestEnv(t)

	authenticator := env.authenticator(t, OIDCConfig{
		Issuer:        testIssuer,
		Audience:      testAudience,
		JWKSURL:       env.jwksURL,
		AllowedEmails: []string{testEmail},
	})

	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)

	_, err := authenticator.Authenticate(context.Background(), req)
	assertUnauthorized(t, err)
}

func TestOIDCAuthenticateWrongIssuer(t *testing.T) {
	env := newOIDCTestEnv(t)

	authenticator := env.authenticator(t, OIDCConfig{
		Issuer:        testIssuer,
		Audience:      testAudience,
		JWKSURL:       env.jwksURL,
		AllowedEmails: []string{testEmail},
	})

	token := env.signToken(t, tokenClaims{
		Issuer:     "https://wrong-issuer.example.com",
		Subject:    testSubject,
		Audience:   testAudience,
		Email:      testEmail,
		Expiration: env.now.Add(time.Hour),
	})

	_, err := authenticator.Authenticate(context.Background(), bearerRequest(token))
	assertUnauthorized(t, err)
}

func TestOIDCAuthenticateWrongAudience(t *testing.T) {
	env := newOIDCTestEnv(t)

	authenticator := env.authenticator(t, OIDCConfig{
		Issuer:        testIssuer,
		Audience:      testAudience,
		JWKSURL:       env.jwksURL,
		AllowedEmails: []string{testEmail},
	})

	token := env.signToken(t, tokenClaims{
		Issuer:     testIssuer,
		Subject:    testSubject,
		Audience:   "other-audience",
		Email:      testEmail,
		Expiration: env.now.Add(time.Hour),
	})

	_, err := authenticator.Authenticate(context.Background(), bearerRequest(token))
	assertUnauthorized(t, err)
}

func TestOIDCAuthenticateExpiredToken(t *testing.T) {
	env := newOIDCTestEnv(t)

	authenticator := env.authenticator(t, OIDCConfig{
		Issuer:        testIssuer,
		Audience:      testAudience,
		JWKSURL:       env.jwksURL,
		AllowedEmails: []string{testEmail},
	})

	token := env.signToken(t, tokenClaims{
		Issuer:     testIssuer,
		Subject:    testSubject,
		Audience:   testAudience,
		Email:      testEmail,
		Expiration: env.now.Add(-time.Minute),
	})

	_, err := authenticator.Authenticate(context.Background(), bearerRequest(token))
	assertUnauthorized(t, err)
}

func TestOIDCAuthenticateNotBeforeInFuture(t *testing.T) {
	env := newOIDCTestEnv(t)

	authenticator := env.authenticator(t, OIDCConfig{
		Issuer:        testIssuer,
		Audience:      testAudience,
		JWKSURL:       env.jwksURL,
		AllowedEmails: []string{testEmail},
	})

	token := env.signToken(t, tokenClaims{
		Issuer:     testIssuer,
		Subject:    testSubject,
		Audience:   testAudience,
		Email:      testEmail,
		Expiration: env.now.Add(time.Hour),
		NotBefore:  env.now.Add(time.Minute),
	})

	_, err := authenticator.Authenticate(context.Background(), bearerRequest(token))
	assertUnauthorized(t, err)
}

func TestOIDCAuthenticateDisallowedIdentity(t *testing.T) {
	env := newOIDCTestEnv(t)

	authenticator := env.authenticator(t, OIDCConfig{
		Issuer:        testIssuer,
		Audience:      testAudience,
		JWKSURL:       env.jwksURL,
		AllowedEmails: []string{"someone-else@example.com"},
	})

	token := env.signToken(t, tokenClaims{
		Issuer:     testIssuer,
		Subject:    testSubject,
		Audience:   testAudience,
		Email:      testEmail,
		Expiration: env.now.Add(time.Hour),
	})

	_, err := authenticator.Authenticate(context.Background(), bearerRequest(token))
	assertUnauthorized(t, err)
}

func TestOIDCAuthenticateAllowedSubjectWithoutEmail(t *testing.T) {
	env := newOIDCTestEnv(t)

	authenticator := env.authenticator(t, OIDCConfig{
		Issuer:          testIssuer,
		Audience:        testAudience,
		JWKSURL:         env.jwksURL,
		AllowedSubjects: []string{testSubject},
	})

	token := env.signToken(t, tokenClaims{
		Issuer:     testIssuer,
		Subject:    testSubject,
		Audience:   testAudience,
		Expiration: env.now.Add(time.Hour),
	})

	principal, err := authenticator.Authenticate(context.Background(), bearerRequest(token))
	if err != nil {
		t.Fatalf("Authenticate returned error: %v", err)
	}

	if principal.Subject != testSubject {
		t.Fatalf("Subject = %q, want %q", principal.Subject, testSubject)
	}
	if principal.Email != "" {
		t.Fatalf("Email = %q, want empty", principal.Email)
	}
}

func TestNewOIDCRequiresIdentityAllowlist(t *testing.T) {
	_, err := NewOIDC(OIDCConfig{
		Issuer:   testIssuer,
		Audience: testAudience,
		JWKSURL:  "https://issuer.example.com/jwks",
	})
	if err == nil {
		t.Fatal("NewOIDC returned nil error")
	}
}

type oidcTestEnv struct {
	privateKey *rsa.PrivateKey
	publicJWK  jwk.Key
	jwksURL    string
	now        time.Time
}

func newOIDCTestEnv(t *testing.T) *oidcTestEnv {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}

	publicJWK, err := jwk.Import(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("import public jwk: %v", err)
	}

	if err := publicJWK.Set(jwk.KeyIDKey, testKeyID); err != nil {
		t.Fatalf("set public jwk kid: %v", err)
	}
	if err := publicJWK.Set(jwk.AlgorithmKey, jwa.RS256()); err != nil {
		t.Fatalf("set public jwk alg: %v", err)
	}

	keySet := jwk.NewSet()
	keySet.AddKey(publicJWK)

	jwksJSON, err := json.Marshal(keySet)
	if err != nil {
		t.Fatalf("marshal jwks: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jwksJSON)
	}))
	t.Cleanup(server.Close)

	return &oidcTestEnv{
		privateKey: privateKey,
		publicJWK:  publicJWK,
		jwksURL:    server.URL,
		now:        time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC),
	}
}

func (e *oidcTestEnv) authenticator(t *testing.T, cfg OIDCConfig) *OIDCAuthenticator {
	t.Helper()

	authenticator, err := NewOIDC(cfg)
	if err != nil {
		t.Fatalf("NewOIDC returned error: %v", err)
	}

	authenticator.now = func() time.Time {
		return e.now
	}

	return authenticator
}

type tokenClaims struct {
	Issuer     string
	Subject    string
	Audience   string
	Email      string
	Expiration time.Time
	NotBefore  time.Time
}

func (e *oidcTestEnv) signToken(t *testing.T, claims tokenClaims) string {
	t.Helper()

	builder := jwt.NewBuilder().
		Issuer(claims.Issuer).
		Subject(claims.Subject).
		Audience([]string{claims.Audience}).
		Expiration(claims.Expiration)

	if !claims.NotBefore.IsZero() {
		builder = builder.NotBefore(claims.NotBefore)
	}
	if claims.Email != "" {
		builder = builder.Claim("email", claims.Email)
	}

	token, err := builder.Build()
	if err != nil {
		t.Fatalf("build jwt: %v", err)
	}

	privateJWK, err := jwk.Import(e.privateKey)
	if err != nil {
		t.Fatalf("import private jwk: %v", err)
	}
	if err := privateJWK.Set(jwk.KeyIDKey, testKeyID); err != nil {
		t.Fatalf("set private jwk kid: %v", err)
	}
	if err := privateJWK.Set(jwk.AlgorithmKey, jwa.RS256()); err != nil {
		t.Fatalf("set private jwk alg: %v", err)
	}

	signed, err := jwt.Sign(token, jwt.WithKey(jwa.RS256(), privateJWK))
	if err != nil {
		t.Fatalf("sign jwt: %v", err)
	}

	return string(signed)
}

func bearerRequest(token string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

func assertUnauthorized(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var unauthorizedErr *UnauthorizedError
	if !errors.As(err, &unauthorizedErr) {
		t.Fatalf("error type = %T, want *UnauthorizedError: %v", err, err)
	}
}
