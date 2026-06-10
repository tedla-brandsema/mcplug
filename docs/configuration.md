# Configuration

MCPFS reads one JSON config file containing the gateway settings (`server`) and the upstream MCP servers (`mcpServers`).

Resolution order:

1. `-config <path>` CLI flag, if given.
2. Otherwise the global config at `<user config dir>/mcpfs/mcpfs.cfg.json` (e.g. `~/.config/mcpfs/mcpfs.cfg.json`), created from an embedded default if missing.

`mcpfs init` writes a starter config with disabled example entries. Config files are created with mode 0600 because `headers` and `env` values may contain secrets; MCPFS warns at startup when a config containing such values is world-readable.

## `server`

Unchanged from MCPFS v1.

| Field | Values | Notes |
| --- | --- | --- |
| `name` | string | Required. MCP server name presented to clients. |
| `version` | string | Required. |
| `transport` | `stdio` (default), `http`, `http_ngrok` | See [transports](transports.md). |
| `addr` | host:port | HTTP only; defaults to `127.0.0.1:8080`. |
| `path` | `/mcp` (default) | HTTP only; must start with `/`. |
| `auth` | object | HTTP only; see [transports](transports.md). Modes: `none`, `bearer`, `oidc`. |
| `ngrok_url` | string | `http_ngrok` only; optional fixed ngrok domain. |

## `mcpServers`

A map of server name → entry. The shape is **compatible with** the `mcpServers` convention used by Claude Desktop and Cursor: `command`/`args`/`env` entries paste in verbatim. The remaining fields are MCPFS extensions.

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/path"],
      "env": {"DEBUG": "1"},
      "cwd": "/path",
      "excludeTools": ["delete_file"]
    },
    "remote": {
      "url": "https://example.com/mcp",
      "headers": {"Authorization": "Bearer ..."},
      "optional": true
    },
    "later": {
      "command": "uvx",
      "args": ["mcp-server-git"],
      "disabled": true
    }
  }
}
```

### Fields

| Field | Applies to | Description |
| --- | --- | --- |
| `command` | stdio | Executable to spawn. Run verbatim via `exec`, never through a shell. Exactly one of `command`/`url` is required. |
| `args` | stdio | Argument vector. |
| `env` | stdio | Extra environment variables, merged over the inherited environment (config wins). Values are never logged. |
| `cwd` | stdio | Working directory for the child. *(MCPFS extension)* |
| `url` | HTTP | Streamable-HTTP MCP endpoint (`http`/`https`). *(MCPFS extension)* |
| `headers` | HTTP | Headers added to every request (e.g. `Authorization`). Values are never logged. *(MCPFS extension)* |
| `disabled` | both | Entry is ignored entirely. Still validated structurally. *(MCPFS extension)* |
| `optional` | both | Changes **startup-failure** behavior only: a failing optional server is logged and skipped instead of aborting startup. Its tools stay absent until MCPFS restarts. A successfully started optional server is supervised and restarted like any other. *(MCPFS extension)* |
| `includeTools` | both | Allowlist of original tool names to expose. Mutually exclusive with `excludeTools`. Unknown names produce a startup warning. *(MCPFS extension)* |
| `excludeTools` | both | Denylist of original tool names to hide. *(MCPFS extension)* |

### Validation rules

- Exactly one of `command` / `url` per entry.
- `url` must be `http://` or `https://`.
- `headers` only with `url`; `args`/`env`/`cwd` only with `command`.
- `includeTools` and `excludeTools` are mutually exclusive.
- Server names must be non-empty and must not collide after sanitization (see below).
- Disabled entries get the same structural validation; no entry is ever checked for command availability or network reachability at validation time.
- An empty `mcpServers` map is valid; MCPFS starts with a warning and exposes zero tools.

## Tool naming

Every upstream tool is exposed as `<server>_<tool>`. The server name is sanitized to the tool-name-safe alphabet:

- characters outside `[A-Za-z0-9_-]` become `_`;
- runs of `_` collapse; leading/trailing `_` are trimmed;
- a name that sanitizes to nothing is a config error;
- a sanitized name not starting with a letter is prefixed with `server_`.

Names are always prefixed, so adding or removing a server never renames the tools of another server.

## Lifecycle

- **Startup:** all enabled servers are started eagerly and their tools listed once. A required server failing aborts startup; optional failures are skipped.
- **Supervision:** stdio children that exit unexpectedly are restarted with exponential backoff (0.5s–30s, reset after 60s healthy). While a server restarts, its tool calls return tool errors ("upstream restarting") rather than protocol failures.
- **Timeouts:** upstream connect/list/call operations are bounded at 60 seconds; a timeout surfaces as a tool error.
- **Snapshot:** the aggregated tool list is fixed at startup. Restart MCPFS to pick up upstream tool changes.
- **Shutdown:** SIGINT/SIGTERM stops the transport first, then terminates all children gracefully (stdin close → SIGTERM → kill).
