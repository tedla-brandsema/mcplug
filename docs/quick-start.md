# Quick start

This guide runs MCPlug locally over stdio, aggregating the reference filesystem and git servers. It needs no network exposure.

## Before you begin

- Go 1.25+ to build MCPlug.
- `npx` (Node.js) for the reference filesystem server.
- `uvx` (uv) for the reference git server — optional.

## 1. Build

```bash
git clone https://github.com/tedla-brandsema/mcplug.git
cd mcplug
go build -o ./bin/plug ./cmd/plug
```

## 2. Create a config

```bash
./bin/plug init
```

This writes a starter config (mode 0600) to your user config directory (e.g. `~/.config/mcplug/mcplug.cfg.json`) with example entries that are disabled. Edit it:

```json
{
  "server": {
    "name": "mcplug",
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
./bin/plug ls
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
./bin/plug
```

Connect any local MCP client to the `./bin/plug` command, or inspect interactively:

```bash
bunx @modelcontextprotocol/inspector ./bin/plug
```

## Next steps

- Restrict tools with `includeTools`/`excludeTools` — see [configuration](configuration.md).
- Serve over HTTP or expose remotely with ngrok — see [transports](transports.md) and read [security](security.md) first.
