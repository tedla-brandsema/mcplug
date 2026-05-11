# Local read/write example

This example shows how to let a trusted local MCP client write files inside a narrow configured root.

## Scenario

You want an MCP client to edit files in a project while keeping writes scoped to that project.

## User goal

Run MCPFS locally with one `read_write` root.

## Files

- `README.md` — this guide.
- `mcpfs.cfg.json` — example local read/write STDIO config.

## Command flow

Start from a clean Git working tree, then run:

```bash
go build -o ./bin/mcpfs ./cmd/mcpfs
./bin/mcpfs -config examples/local-read-write/mcpfs.cfg.json
```

Connect a trusted local MCP client over STDIO.

## Expected output

From the MCP client:

- `fs_roots` lists the configured root in `read_write` mode.
- `fs_write` can create a file inside the root.
- writes outside the root are rejected.
- excluded and ignored paths are rejected.

After a write, check Git status:

```bash
git status --porcelain=v1 -b
```

## Security notes

Use `read_write` only for roots you are comfortable modifying. Keep the root narrow and review Git status after write operations.

## Related docs

- [Security](../../docs/security.md)
- [Configure local read/write access](../../docs/how-to/configure-local-read-write.md)
- [Configuration](../../docs/configuration.md)
