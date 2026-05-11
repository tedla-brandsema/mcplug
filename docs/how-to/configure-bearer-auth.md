# Configure bearer authentication

Use this guide when MCPFS runs over HTTP and clients can share a static bearer token.

## Before you begin

You need an MCPFS HTTP config and a secure way to provide an environment variable to the server process.

## Generate a token

For local development, generate a high-entropy token:

```bash
export MCPFS_TOKEN="$(openssl rand -hex 32)"
```

Do not commit real tokens.

## Configure the server

Set HTTP transport and bearer auth:

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

## Run the server

```bash
mcpfs -config /path/to/mcpfs.cfg.json
```

## Verify the configuration

Check the health endpoint:

```bash
curl http://127.0.0.1:8080/healthz
```

Then verify that unauthenticated MCP requests are rejected and authenticated MCP requests reach the MCP handler.

## Troubleshoot

If requests fail unexpectedly, check that:

- `MCPFS_TOKEN` is set in the server process environment;
- the client sends the expected bearer token;
- the client uses the configured HTTP path;
- the server is bound to the address you expect.

## Next steps

- Read [Transports](../transports.md).
- Read [Security](../security.md).
- Try the [HTTP bearer example](../../examples/http-bearer/).
