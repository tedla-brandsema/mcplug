# Configure OIDC authentication

Use this guide when MCPFS runs over HTTP and clients authenticate with JWTs from an external identity provider.

## Before you begin

You need:

- an identity provider that publishes JWKS;
- the expected issuer;
- the expected audience;
- at least one allowed email or subject;
- an MCP client that can send a valid bearer JWT.

## Configure the server

Set HTTP transport and OIDC auth:

```json
"server": {
  "name": "mcpfs",
  "version": "0.4.0",
  "transport": "http",
  "addr": "127.0.0.1:8080",
  "path": "/mcp",
  "auth": {
    "mode": "oidc",
    "issuer": "https://issuer.example.com",
    "audience": "mcpfs",
    "jwks_url": "https://issuer.example.com/.well-known/jwks.json",
    "allowed_subjects": ["user-or-client-subject-id"]
  }
}
```

Use `allowed_emails` instead of `allowed_subjects` when email allowlisting is the right identity boundary.

## Run the server

```bash
mcpfs -config /path/to/mcpfs.cfg.json
```

## Verify the configuration

Test these cases before production use:

1. A request without a JWT is rejected.
2. A request with an invalid JWT is rejected.
3. A request with the wrong issuer is rejected.
4. A request with the wrong audience is rejected.
5. A request with a valid JWT and allowed identity reaches the MCP handler.

## Troubleshoot

If valid tokens fail, check issuer, audience, JWKS URL, token expiry, clock skew, and the allowed identity list.

## Next steps

- Read [Security](../security.md).
- Read [Authentication model](../advanced/authentication-model.md).
- Try the [OIDC example](../../examples/oidc/).
