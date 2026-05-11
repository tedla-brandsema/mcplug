# Configuration

MCPFS uses JSON configuration to define the server, roots, authentication, transports, and optional commands.

There are two configuration layers:

| Config file | Purpose |
| --- | --- |
| `mcpfs.cfg.json` | Server settings, configured roots, and configured commands. The default global path is `os.UserConfigDir()/mcpfs/mcpfs.cfg.json`. |
| `.mcpfs/project.cfg.json` | Project-local overview detection rules for a configured root. |

## Global MCPFS config

The global config controls runtime behavior.

Minimal read-only STDIO config:

```json
{
  "server": {
    "name": "mcpfs",
    "version": "0.4.0",
    "transport": "stdio",
    "auth": {
      "mode": "none"
    }
  },
  "roots": [
    {
      "id": "project",
      "path": "/path/to/project",
      "mode": "read",
      "include": ["**/*"],
      "exclude": ["**/.git/**", "**/.env", "**/.env.*"],
      "use_gitignore": true,
      "max_file_bytes": 262144
    }
  ],
  "commands": {
    "mode": "disabled"
  }
}
```

## Server settings

Common server fields:

| Field | Description |
| --- | --- |
| `name` | MCP server name. |
| `version` | MCP server version. |
| `transport` | `stdio`, `http`, or `http_ngrok`. |
| `addr` | HTTP bind address. Defaults to `127.0.0.1:8080` for HTTP transports. |
| `path` | MCP HTTP path. Defaults to `/mcp`. |
| `auth` | HTTP auth configuration. |
| `ngrok_url` | Optional reserved ngrok URL or domain. |

## Roots

Each root defines a local directory MCPFS can expose.

| Field | Description |
| --- | --- |
| `id` | Stable root identifier used by MCP tools. |
| `path` | Local filesystem path to expose. |
| `mode` | `read` or `read_write`. |
| `include` | Glob allowlist. Empty means all files not excluded are allowed. |
| `exclude` | Glob denylist. |
| `use_gitignore` | Apply `.gitignore` rules as an additional filter. |
| `max_file_bytes` | Maximum readable or writable file size. Defaults to `262144` when set to `0`. |

Use narrow roots. Avoid exposing home directories, credential directories, or broad workspaces.

## Access modes

`mode: "read"` allows inspection but not writes.

`mode: "read_write"` enables `fs_write` for that root. Writes still honor root boundaries, symlink checks, include/exclude rules, `.gitignore`, and file size limits.

## Authentication settings

HTTP transports support these auth modes:

| Mode | Use case |
| --- | --- |
| `none` | Local-only HTTP or short-lived controlled development. Do not use on untrusted networks. |
| `bearer` | Static bearer token loaded from an environment variable. |
| `oidc` | JWT/OIDC validation through issuer, audience, JWKS, and identity allowlists. |

Bearer example:

```json
"auth": {
  "mode": "bearer",
  "token_env": "MCPFS_TOKEN"
}
```

OIDC example:

```json
"auth": {
  "mode": "oidc",
  "issuer": "https://issuer.example.com",
  "audience": "mcpfs",
  "jwks_url": "https://issuer.example.com/.well-known/jwks.json",
  "allowed_subjects": ["user-or-client-subject-id"]
}
```

At least one of `allowed_emails` or `allowed_subjects` must be configured for OIDC.

## Commands

Command execution is configured separately from root access.

| Field | Description |
| --- | --- |
| `commands.mode` | `disabled`, `predefined`, or `unguarded`. Defaults to `disabled`. |
| `commands.defaults.timeout_seconds` | Default command timeout. Defaults to `60` when unset. |
| `commands.defaults.max_output_bytes` | Default combined stdout/stderr output limit. Defaults to `65536` when unset. |
| `commands.items[].id` | Stable command id used by `cmd_run`. |
| `commands.items[].description` | Optional human-readable description. |
| `commands.items[].root_id` | Root id used to scope the command working directory. |
| `commands.items[].workdir` | Relative working directory inside the root. Defaults to `.`. |
| `commands.items[].command` | Fixed argv array to execute. The first item is the executable. |
| `commands.items[].timeout_seconds` | Optional per-command timeout override. |
| `commands.items[].max_output_bytes` | Optional per-command output limit override. |

See [Commands](commands.md) for security guidance and examples.

## Project-local config

A project-local config lives at:

```text
.mcpfs/project.cfg.json
```

It customizes project overview detection rules for a configured root. When present, it is merged with the global or user project registry for that root only.

Create it with:

```bash
mcpfs init
```

## Reference

For exact fields, accepted values, defaults, and compatibility notes, see [Configuration schema](reference/config-schema.md).
