# HTTP bearer example

This example shows how to run MCPFS over HTTP with a bearer token.

## Scenario

You need a local or controlled network HTTP endpoint for an MCP client.

## User goal

Run MCPFS with HTTP transport and bearer authentication.

## Files

- `README.md` — this guide.
- `mcpfs.cfg.json` — example HTTP bearer config.
- `env.example` — example environment variable name.

## Command flow

Build MCPFS:

```bash
go build -o ./bin/mcpfs ./cmd/mcpfs
```

Set a development token:

```bash
export MCPFS_TOKEN="$(openssl rand -hex 32)"
```

Run the server:

```bash
./bin/mcpfs -config examples/http-bearer/mcpfs.cfg.json
```

Check health:

```bash
curl http://127.0.0.1:8080/healthz
```

Configure your MCP client to connect to:

```text
http://127.0.0.1:8080/mcp
```

## Expected output

- Health check succeeds.
- Requests without the expected token are rejected.
- Requests with the expected token reach the MCP handler.

## Security notes

Do not commit real tokens. Use TLS or a trusted reverse proxy for remote HTTP. Keep roots read-only unless writes are required.

## Related docs

- [Transports](../../docs/transports.md)
- [Security](../../docs/security.md)
- [Configure bearer authentication](../../docs/how-to/configure-bearer-auth.md)
