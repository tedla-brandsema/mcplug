package auth

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwk"
)

type CachingJWKS struct {
	url    string
	client *http.Client
	ttl    time.Duration

	mu        sync.Mutex
	keySet    jwk.Set
	fetchedAt time.Time
}

func NewCachingJWKS(url string) *CachingJWKS {
	return &CachingJWKS{
		url: url,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		ttl: 5 * time.Minute,
	}
}

func (c *CachingJWKS) KeySet(ctx context.Context) (jwk.Set, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.keySet != nil && time.Since(c.fetchedAt) < c.ttl {
		return c.keySet, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
	if err != nil {
		return nil, fmt.Errorf("create jwks request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch jwks: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fetch jwks: unexpected status %s", resp.Status)
	}

	keySet, err := jwk.ParseReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse jwks: %w", err)
	}

	c.keySet = keySet
	c.fetchedAt = time.Now()

	return c.keySet, nil
}
