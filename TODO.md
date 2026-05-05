- [ ] Authenticate wraps JWKS loading failures as a plain error:
```go
return nil, fmt.Errorf("load jwks: %w", err)
```
The HTTP layer currently maps non-`UnauthorizedError` auth failures to `500`. That is defensible for “auth backend unavailable.” Token problems still become `401`.