# Configure local read-only access

Use this guide when you want a trusted local MCP client to inspect a project without writes or command execution.

## Before you begin

Build MCPFS and choose a project directory to expose read-only.

## Configure the root

Add the current project to the default global config:

```bash
mcpfs project add
```

The added root uses `mode: "read"` by default.

## Run the server

```bash
mcpfs
```

Connect your local MCP client over STDIO.

## Verify the configuration

Call `fs_roots` from your MCP client.

Expected result:

- the root appears;
- the mode is `read`;
- no command execution tools are registered when `commands.mode` is `disabled`.

Call `project_overview` to confirm MCPFS can inspect the project.

## Troubleshoot

If the root is missing, confirm the client uses the same MCPFS config you updated.

If file reads fail, check include rules, exclude rules, `.gitignore`, root id, and file size limits.

## Next steps

- Read [Security](../security.md).
- Try the [local read-only example](../../examples/local-read-only/).
