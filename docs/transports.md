# Transports

MCPFS supports local and HTTP-based transports. Choose the narrowest transport that works for your client.

## Transport options

| Transport | Use case |
| --- | --- |
| `stdio` | Local MCP clients and MCP Inspector. Recommended first. |
| `http` | Network-accessible MCP clients. Requires careful auth and exposure review. |
| `http_ngrok` | Short-lived development tunnels. Treat as higher risk. |

## STDIO

STDIO is the recommended first transport because it does not open a network listener.

Run with the default global config:

```bash
mcpfs
```

Run with an explicit config:

```bash
mcpfs -config /path/to/mcpfs.cfg.json
```

Use STDIO for local MCP hosts that can start MCPFS as a child process.

## HTTP

HTTP transport is useful when an MCP client needs a URL endpoint.

Minimal HTTP server settings:

```json
"server": {
  "name": "mcpfs",
  "version": "0.4.0",
  "transport": "http",
  "addr": "127.0.0.1:8080",
  "path": "/mcp",
  "auth": {
    "mode": "bearer",
    "token_env": "MCPFS_TOKEN"
  }
}
```

Start the server:

```bash
export MCPFS_TOKEN="$(openssl rand -hex 32)"
mcpfs -config /path/to/mcpfs.cfg.json
```

Check the health endpoint:

```bash
curl http://127.0.0.1:8080/healthz
```

A plain `GET /mcp` may return a protocol-level MCP response such as `405 Method Not Allowed`. That can still mean the request reached the MCP handler.

## Bearer auth

Bearer auth checks an `Authorization: Bearer ...` header against a token loaded from an environment variable.

Example request:

```bash
curl -i \
  -H "Authorization: Bearer $MCPFS_TOKEN" \
  http://127.0.0.1:8080/mcp
```

Use bearer auth for simple trusted deployments. Use TLS for remote access.

## OIDC/JWT auth

OIDC auth validates bearer JWTs from an external identity provider.

Required settings:

- `issuer`
- `audience`
- `jwks_url`
- `allowed_emails` or `allowed_subjects`

MCPFS validates JWT signature, issuer, audience, expiry, not-before time when present, and identity allowlists.

Use OIDC when you want identity-provider-backed authentication instead of a static shared token.

## ngrok development tunnels

`http_ngrok` starts MCPFS with an embedded ngrok tunnel for remote development testing.

Use ngrok only for short-lived development unless you have reviewed and hardened the deployment.

Safer defaults for ngrok examples:

- read-only roots;
- bearer or OIDC auth;
- no unguarded command execution;
- short-lived tokens;
- stop the tunnel when testing ends.

Avoid combining ngrok with `auth.mode: "none"`, `read_write` roots, or `commands.mode: "unguarded"` unless you fully accept the risk in a controlled environment.

## Related docs

- [Security](security.md)
- [Configuration](configuration.md)
- [Configure bearer authentication](how-to/configure-bearer-auth.md)
- [Configure OIDC authentication](how-to/configure-oidc.md)
- [ngrok development](advanced/ngrok-development.md)
