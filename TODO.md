- [ ] Decide how OIDC JWKS/backend failures should map to HTTP responses.
  Currently `Authenticate` wraps JWKS loading failures as a plain error:

  ```go
  return nil, fmt.Errorf("load jwks: %w", err)
  ```
  The HTTP layer currently maps non-`UnauthorizedError` auth failures to `500`. That may be correct for an unavailable auth backend, while malformed/invalid tokens remain `401`.