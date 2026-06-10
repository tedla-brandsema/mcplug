# MCPFS

MCPFS is an MCP (Model Context Protocol) aggregating gateway. It launches and connects to the MCP servers you configure — reference servers, community servers, remote endpoints — and exposes all of their tools through a single MCP endpoint, locally over stdio or HTTP, and remotely through an embedded ngrok tunnel with bearer or OIDC authentication.

## What MCPFS is

The MCP ecosystem has excellent servers (filesystem, git, fetch, and many more), but most of them speak stdio only: no HTTP transport, no auth, no public URL. Remote MCP clients such as ChatGPT connectors need exactly that.

MCPFS fills the gap. You describe your servers in a Claude/Cursor-compatible `mcpServers` map, and MCPFS:

- spawns stdio entries as supervised child processes (restarted with backoff if they crash);
- connects to `url` entries over streamable HTTP;
- aggregates every tool into one MCP server, each tool prefixed with its server name (`filesystem_read_file`, `git_status`, …);
- serves the aggregate over stdio, localhost HTTP, or HTTP + ngrok;
- authenticates remote clients with bearer tokens or OIDC.

MCPFS implements no tools of its own. (Version 1 was a filesystem/git MCP server with native tools; see [Migration from v1](#migration-from-v1).)

## Quick start

Build:

```bash
git clone https://github.com/tedla-brandsema/mcpfs.git
cd mcpfs
go test ./...
go build -o ./bin/mcpfs ./cmd/mcpfs
```

Create a starter config (written with mode 0600 to your user config directory, e.g. `~/.config/mcpfs/mcpfs.cfg.json`):

```bash
./bin/mcpfs init
```

Edit it to add servers. A minimal local setup with the reference filesystem and git servers:

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
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/home/you/projects/myproject"]
    },
    "git": {
      "command": "uvx",
      "args": ["mcp-server-git", "--repository", "/home/you/projects/myproject"]
    }
  }
}
```

Smoke-test the config — this probes every server and lists its tools without starting any transport:

```bash
./bin/mcpfs ls
```

Run the gateway:

```bash
./bin/mcpfs
```

Inspect it interactively with the MCP Inspector:

```bash
bunx @modelcontextprotocol/inspector ./bin/mcpfs
```

For a complete first-run path, see the [quick start](docs/quick-start.md).

## Configuration

One JSON file holds the gateway settings (`server`) and the upstream servers (`mcpServers`). The `mcpServers` shape is compatible with the convention used by Claude Desktop and Cursor, so config snippets from server READMEs paste in as-is. MCPFS adds extensions: `url`, `headers`, `disabled`, `optional`, `cwd`, `includeTools`, and `excludeTools`.

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/path"],
      "env": {"DEBUG": "1"},
      "excludeTools": ["delete_file"]
    },
    "remote": {
      "url": "https://example.com/mcp",
      "headers": {"Authorization": "Bearer ..."},
      "optional": true
    }
  }
}
```

Key behaviors:

- **Commands run verbatim** via `exec`, never through a shell.
- **Enabled servers are required by default**: if one fails at startup, MCPFS refuses to start. Mark a server `"optional": true` to log and skip it instead (its tools stay absent until you restart MCPFS). `"disabled": true` ignores an entry entirely.
- **Tool names are stable**: every tool is exposed as `<server>_<tool>` with the server name sanitized to `[A-Za-z0-9_-]`.
- **The tool list is a startup snapshot**; restart MCPFS to pick up upstream tool changes.

See [docs/configuration.md](docs/configuration.md) for the full field reference.

## Remote access (ChatGPT and other remote clients)

Switch the transport to `http_ngrok` and require auth:

```json
{
  "server": {
    "name": "mcpfs",
    "version": "2.0.0",
    "transport": "http_ngrok",
    "addr": "127.0.0.1:8080",
    "path": "/mcp",
    "auth": {"mode": "bearer", "token_env": "MCPFS_TOKEN"}
  }
}
```

MCPFS prints the public MCP URL at startup; add it as a connector in your remote client. See [docs/transports.md](docs/transports.md) and the [examples](examples/).

## Security

MCPFS does **not** sandbox upstream servers: stdio children run with the same OS privileges as MCPFS itself. Configure only servers you trust, keep config files private (`headers` and `env` may contain secrets; MCPFS warns when such a config is world-readable and never logs those values), and never expose the HTTP transport publicly without bearer or OIDC auth. Read [docs/security.md](docs/security.md) before enabling `http` or `http_ngrok`.

## CLI

| Command | Purpose |
| --- | --- |
| `mcpfs [-config path]` | Run the gateway with the configured transport |
| `mcpfs init [-path p]` | Write a starter config (existing files untouched) |
| `mcpfs ls [-config p]` | Probe all configured servers and list their tools; exits non-zero if a required server fails |

## Examples

- [HTTP bearer](examples/http-bearer/) — localhost HTTP transport with a bearer token.
- [OIDC](examples/oidc/) — HTTP transport validating JWTs from an identity provider.
- [ngrok development](examples/ngrok-dev/) — public development tunnel for remote clients.

## Migration from v1

MCPFS v1 was an MCP server with native filesystem (`fs_*`), git (`git_*`), project-overview, and command-execution tools configured through `roots` and `commands`. Version 2 removes all native tools: MCPFS is now purely a gateway, and the ecosystem servers provide the tools.

A v1 read-only root:

```json
{"roots": [{"id": "project", "path": "/home/you/projects/myproject", "mode": "read"}]}
```

becomes a reference-server entry in v2:

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/home/you/projects/myproject"]
    },
    "git": {
      "command": "uvx",
      "args": ["mcp-server-git", "--repository", "/home/you/projects/myproject"]
    }
  }
}
```

Notes:

- The reference filesystem server has no read-only mode comparable to v1 roots; use `includeTools` to restrict it (for example, keep only the `read_*`, `list_*`, and `search_*` tools).
- v1's `commands` execution has no v2 equivalent by design.
- The transport, auth, and ngrok configuration (`server` block) is unchanged.
- v1 is preserved on the `legacy/v1` branch.

## Maturity and compatibility

MCPFS v2 is currently **Beta**. Configuration fields and CLI behavior may still change; breaking changes are documented in the [changelog](CHANGELOG.md).

## Contributing

Contributions are welcome. Start with [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MCPFS is licensed under the terms in [LICENSE](LICENSE).
