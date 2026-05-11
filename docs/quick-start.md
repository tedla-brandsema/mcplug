# Quick start

Use this guide to run MCPFS locally with a read-only project root. This is the safest first setup because it avoids network exposure, file writes, and command execution.

## Before you begin

You need:

- Go installed.
- A local clone of MCPFS.
- A project directory that you are comfortable exposing read-only to a trusted MCP client.

## Build MCPFS

From the MCPFS repository root, run:

```bash
go test ./...
go build -o ./bin/mcpfs ./cmd/mcpfs
```

## Create project-local metadata config

From the project you want to expose, run:

```bash
/path/to/mcpfs/bin/mcpfs init
```

This writes `.mcpfs/project.cfg.json` for project overview detection. It does not expose the project by itself.

## Add a read-only project root

Add the current directory to the default global MCPFS config:

```bash
/path/to/mcpfs/bin/mcpfs project add
```

The added root uses read mode by default.

List configured roots:

```bash
/path/to/mcpfs/bin/mcpfs project ls
```

Expected output is a small table that includes the root id, `read` mode, and project path.

## Start MCPFS

Run MCPFS with the default STDIO transport:

```bash
/path/to/mcpfs/bin/mcpfs
```

Connect your local MCP client to that command.

## Check the result

From your MCP client, call `fs_roots`.

Expected result:

- The configured root is listed.
- The root mode is `read`.
- Absolute host paths are not exposed by the tool response.

Then call `project_overview` for the configured root.

Expected result:

- MCPFS returns a bounded project summary.
- Git status and recent commits are included when Git is available.
- Files ignored by `.gitignore` or explicit exclude rules are not included.

## Clean up

To remove the root from the default global config, run:

```bash
/path/to/mcpfs/bin/mcpfs project rm
```

If you no longer want project-local overview rules, remove:

```text
.mcpfs/project.cfg.json
```

## Next steps

- Read [Security](security.md) before enabling writes, HTTP, ngrok, or commands.
- Use [Configure local read/write access](how-to/configure-local-read-write.md) when you need writes.
- Use [Add predefined commands](how-to/add-predefined-commands.md) when you want MCPFS to run known development commands.
- Browse [Examples](../examples/) for complete scenarios.
