# OIDC example

This example shows how to configure MCPFS as an HTTP MCP server that validates JWTs from an external identity provider.

## Scenario

You need HTTP access to MCPFS, but you want identity-provider-backed authentication instead of a shared static token.

## User goal

Run MCPFS with HTTP transport and OIDC/JWT validation.

## Files

- `README.md` — this guide.
- `mcpfs.cfg.json` — example OIDC config with placeholder issuer, audience, JWKS, and identity allowlist values.

## Command flow

Build MCPFS:

```bash
go build -o ./bin/mcpfs ./cmd/mcpfs
```

Edit `examples/oidc/mcpfs.cfg.json` and replace the placeholder values:

- `issuer`
- `audience`
- `jwks_url`
- `allowed_subjects` or `allowed_emails`

Run the server:

```bash
./bin/mcpfs -config examples/oidc/mcpfs.cfg.json
```

Configure your MCP client to connect to:

```text
http://127.0.0.1:8080/mcp
```

## Expected output

- Requests without a JWT are rejected.
- Requests with an invalid JWT are rejected.
- Requests with the wrong issuer, audience, or identity are rejected.
- Requests with a valid JWT and allowed identity reach the MCP handler.

## Security notes

OIDC configuration is security-critical. Validate issuer, audience, signature, expiry, and identity allowlists. Do not use placeholder values outside local testing.

## Related docs

- [Configure OIDC authentication](../../docs/how-to/configure-oidc.md)
- [Authentication model](../../docs/advanced/authentication-model.md)
- [Security](../../docs/security.md)
