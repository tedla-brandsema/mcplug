# Configuration schema reference

This page summarizes the MCPFS JSON configuration fields. See [Configuration](../configuration.md) for usage guidance.

## Top-level object

| Field | Type | Description |
| --- | --- | --- |
| `server` | object | Server, transport, and auth settings. |
| `roots` | array | Configured filesystem roots. |
| `commands` | object | Optional command execution settings. |

## `server`

| Field | Type | Accepted values | Description |
| --- | --- | --- | --- |
| `name` | string | any string | MCP server name. |
| `version` | string | semantic version string | MCP server version. |
| `transport` | string | `stdio`, `http`, `http_ngrok` | Transport mode. |
| `addr` | string | host:port | HTTP bind address. Defaults to `127.0.0.1:8080` for HTTP transports. |
| `path` | string | URL path | MCP HTTP path. Defaults to `/mcp`. |
| `auth` | object | see below | HTTP auth settings. |
| `ngrok_url` | string | URL or empty string | Optional reserved ngrok URL or domain. |

## `server.auth`

| Field | Type | Accepted values | Description |
| --- | --- | --- | --- |
| `mode` | string | `none`, `bearer`, `oidc` | HTTP auth mode. |
| `token_env` | string | environment variable name | Environment variable containing the bearer token. |
| `issuer` | string | issuer URL | Expected JWT issuer for OIDC. |
| `audience` | string | audience string | Expected JWT audience for OIDC. |
| `jwks_url` | string | URL | JWKS URL used to verify JWT signatures. |
| `allowed_emails` | array | email strings | Optional OIDC email allowlist. |
| `allowed_subjects` | array | subject strings | Optional OIDC subject allowlist. |

For OIDC, configure at least one of `allowed_emails` or `allowed_subjects`.

## `roots[]`

| Field | Type | Accepted values | Description |
| --- | --- | --- | --- |
| `id` | string | stable root id | Identifier used by MCP tools. |
| `path` | string | filesystem path | Local path to expose. |
| `mode` | string | `read`, `read_write` | Access mode. |
| `include` | array | glob strings | Allowlist patterns. Empty means all non-excluded files. |
| `exclude` | array | glob strings | Denylist patterns. |
| `use_gitignore` | boolean | `true`, `false` | Whether to apply `.gitignore` rules. |
| `max_file_bytes` | number | positive integer | Maximum readable or writable file size. Defaults to `262144` when set to `0`. |

## `commands`

| Field | Type | Accepted values | Description |
| --- | --- | --- | --- |
| `mode` | string | `disabled`, `predefined`, `unguarded` | Command execution mode. Defaults to `disabled`. |
| `defaults` | object | see below | Default command limits. |
| `items` | array | command objects | Predefined commands. |

## `commands.defaults`

| Field | Type | Description |
| --- | --- | --- |
| `timeout_seconds` | number | Default command timeout. Defaults to `60` when unset. |
| `max_output_bytes` | number | Default combined stdout/stderr output limit. Defaults to `65536` when unset. |

## `commands.items[]`

| Field | Type | Description |
| --- | --- | --- |
| `id` | string | Stable command id used by `cmd_run`. |
| `description` | string | Optional human-readable description. |
| `root_id` | string | Root id used to scope the command working directory. |
| `workdir` | string | Relative working directory inside the root. Defaults to `.`. |
| `command` | array | Fixed argv array to execute. The first item is the executable. |
| `timeout_seconds` | number | Optional per-command timeout override. |
| `max_output_bytes` | number | Optional per-command output limit override. |

## Project-local config

Project-local config lives at `.mcpfs/project.cfg.json` and controls project overview detection rules.

Common project-local fields include important files, source extensions, test patterns, documentation extensions, documentation files, configuration extensions, and configuration files.

Create a project-local config with:

```bash
mcpfs init
```

## Compatibility notes

Legacy `require_auth` and `auth_token_env` fields are still accepted for compatibility, but new configs should use `auth.mode` and `auth.token_env`.
