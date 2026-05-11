# Local read-only example

This example shows how to expose a local project root to a trusted MCP client without writes or command execution.

## Scenario

You want an MCP client to inspect project files, search code, and read Git metadata.

## User goal

Run MCPFS locally with one read-only root.

## Files

- `README.md` — this guide.
- `mcpfs.cfg.json` — example read-only STDIO config.

## Command flow

Build MCPFS from the repository root:

```bash
go build -o ./bin/mcpfs ./cmd/mcpfs
```

Run this example:

```bash
./bin/mcpfs -config examples/local-read-only/mcpfs.cfg.json
```

Connect a local MCP client over STDIO.

## Expected output

From the MCP client:

- `fs_roots` lists the configured root in `read` mode.
- `project_overview` returns a bounded project summary.
- `fs_read` can read allowed files.
- `fs_write` is rejected because the root is read-only.

## Security notes

This is the safest starting mode. It does not expose a network port, does not allow writes, and does not register command execution tools.

## Related docs

- [Quick start](../../docs/quick-start.md)
- [Security](../../docs/security.md)
- [Configure local read-only access](../../docs/how-to/configure-local-read-only.md)
