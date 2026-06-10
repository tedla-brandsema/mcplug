# Transports

MCPFS serves the aggregated MCP endpoint over one of three transports, configured in `server.transport`. Choose the narrowest transport that works for your client.

## `stdio` (default)

For local MCP clients that spawn the server themselves (Claude Desktop, IDEs, the MCP Inspector).

```json
{"server": {"name": "mcpfs", "version": "2.0.0", "transport": "stdio"}}
```

Client config (Claude Desktop style):

```json
{
  "mcpServers": {
    "mcpfs": {"command": "/path/to/bin/mcpfs"}
  }
}
```

## `http`

Streamable HTTP on a local address. `GET /healthz` reports liveness.

```json
{
  "server": {
    "name": "mcpfs",
    "version": "2.0.0",
    "transport": "http",
    "addr": "127.0.0.1:8080",
    "path": "/mcp",
    "auth": {"mode": "bearer", "token_env": "MCPFS_TOKEN"}
  }
}
```

Keep `addr` on localhost unless you have a reverse proxy with TLS and auth in front.

## `http_ngrok`

Same HTTP server plus an embedded ngrok tunnel. MCPFS logs the public MCP URL at startup; add it as a connector in remote clients (e.g. ChatGPT). Requires ngrok credentials in the environment (`NGROK_AUTHTOKEN`). `ngrok_url` optionally pins a reserved domain.

Always combine `http_ngrok` with `bearer` or `oidc` auth — see [security](security.md).

## Authentication modes (`server.auth`)

### `none`

No authentication. Only acceptable for localhost.

### `bearer`

Shared token compared in constant time. The token is read from the environment variable named by `token_env`, never from the config file.

```json
{"auth": {"mode": "bearer", "token_env": "MCPFS_TOKEN"}}
```

```bash
MCPFS_TOKEN=$(openssl rand -hex 32) ./bin/mcpfs
```

Clients send `Authorization: Bearer <token>`.

### `oidc`

Validates JWTs against an identity provider: issuer, audience, expiry, and a subject/email allowlist, with JWKS fetched from `jwks_url` (cached 5 minutes).

```json
{
  "auth": {
    "mode": "oidc",
    "issuer": "https://issuer.example.com",
    "audience": "mcpfs",
    "jwks_url": "https://issuer.example.com/jwks",
    "allowed_emails": ["you@example.com"]
  }
}
```

`allowed_subjects` may be used instead of (or alongside) `allowed_emails`; at least one of the two lists is required.

## Failure semantics at the endpoint

- Missing/invalid credentials → HTTP 401.
- Auth backend failure (e.g. JWKS unreachable) → HTTP 500.
- Upstream server failures (crash, restart, timeout) surface as MCP **tool errors** on the affected tools; the endpoint itself stays up.
