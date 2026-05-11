# ngrok development example

This example shows how to expose MCPFS through a short-lived ngrok tunnel for development testing.

## Scenario

You need a temporary public URL so a remote MCP client can reach a local MCPFS server during development.

## User goal

Run MCPFS with `http_ngrok`, a read-only root, and bearer authentication.

## Files

- `README.md` — this guide.
- `mcpfs.cfg.json` — safer ngrok development config.

## Command flow

Build MCPFS:

```bash
go build -o ./bin/mcpfs ./cmd/mcpfs
```

Set development tokens in your shell or secret manager:

```bash
export NGROK_AUTHTOKEN="<your-ngrok-authtoken>"
export MCPFS_TOKEN="$(openssl rand -hex 32)"
```

Run the server:

```bash
./bin/mcpfs -config examples/ngrok-dev/mcpfs.cfg.json
```

Copy the public MCP URL from the logs and configure your remote MCP client to use it.

## Expected output

- MCPFS starts an HTTP server locally.
- ngrok creates a public tunnel.
- Requests without the bearer token are rejected.
- Requests with the bearer token reach the MCP handler.

## Security notes

Use this example for development only. Stop the tunnel when testing ends.

This canonical example uses a read-only root and bearer auth. Avoid public tunnels with `auth.mode: "none"`, `read_write` roots, or `commands.mode: "unguarded"` unless you fully accept the risk in a controlled environment.

## Related docs

- [ngrok development](../../docs/advanced/ngrok-development.md)
- [Transports](../../docs/transports.md)
- [Security](../../docs/security.md)
