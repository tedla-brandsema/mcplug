# Configure local read/write access

Use this guide when a trusted local MCP client needs to write files inside a narrow project root.

## Before you begin

Start with a clean Git working tree. Choose the narrowest root that needs write access.

## Configure the root

In your MCPFS config, set the root mode to `read_write`:

```json
"mode": "read_write"
```

Keep include and exclude rules narrow. Avoid directories with secrets.

## Run the server

```bash
mcpfs -config /path/to/mcpfs.cfg.json
```

## Verify the configuration

From your MCP client, write a small test file inside the root with `fs_write`.

Then check Git status:

```bash
git status --porcelain=v1 -b
```

Expected result:

- the test file appears as an expected change;
- writes outside the configured root are rejected;
- writes to ignored or excluded paths are rejected.

## Clean up

Delete the test file or revert the change:

```bash
git checkout -- path/to/test-file
```

## Troubleshoot

If writes fail, check root mode, root id, include/exclude rules, `.gitignore`, path boundaries, symlinks, and file size limits.

## Next steps

- Read [Security](../security.md).
- Try the [local read/write example](../../examples/local-read-write/).
