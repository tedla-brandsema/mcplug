# Quick start

This guide runs MCPFS locally over stdio, aggregating the reference filesystem and git servers. It needs no network exposure.

## Before you begin

- Go 1.25+ to build MCPFS.
- `npx` (Node.js) for the reference filesystem server.
- `uvx` (uv) for the reference git server — optional.

## 1. Build

```bash
git clone https://github.com/tedla-brandsema/mcpfs.git
cd mcpfs
go build -o ./bin/mcpfs ./cmd/mcpfs
```

## 2. Create a config

```bash
./bin/mcpfs init
```

This writes a starter config (mode 0600) to your user config directory (e.g. `~/.config/mcpfs/mcpfs.cfg.json`) with example entries that are disabled. Edit it:

```json
{
  "server": {
    "name": "mcpfs",
    "version": "2.0.0",
    "transport": "stdio"
  },
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/absolute/path/to/project"]
    },
    "git": {
      "command": "uvx",
      "args": ["mcp-server-git", "--repository", "/absolute/path/to/project"]
    }
  }
}
```

## 3. Smoke-test

```bash
./bin/mcpfs ls
```

`ls` probes every configured server and prints its tools (exposed name and original name) without starting any transport, tunnel, or listener. It exits non-zero if a required server fails.

Expected output shape:

```text
filesystem (stdio) — running, 14 tool(s)
  filesystem_read_file  (read_file)
  ...
```

## 4. Run

```bash
./bin/mcpfs
```

Connect any local MCP client to the `./bin/mcpfs` command, or inspect interactively:

```bash
bunx @modelcontextprotocol/inspector ./bin/mcpfs
```

## Next steps

- Restrict tools with `includeTools`/`excludeTools` — see [configuration](configuration.md).
- Serve over HTTP or expose remotely with ngrok — see [transports](transports.md) and read [security](security.md) first.
