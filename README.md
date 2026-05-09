# mcpfs

`mcpfs` is a Model Context Protocol server for exposing local project context and controlled project operations to MCP clients.

It lets an MCP client work with explicitly configured project roots without uploading files or manually copying snippets into chat. It is designed for live developer context: list files, view bounded project trees, read files and line ranges, search source, inspect Git status and diffs, review commits, inspect blame, get compact project overviews, optionally write files, and optionally run configured development commands.

The core idea is simple:

```text
configured roots
  → safe path resolver
  → .gitignore-aware matcher
  → filesystem tools
  → git inspection tools
  → project overview tools
  → command tools
  → MCP
```

## Warning

`mcpfs` is a power tool.

When configured with write access or command execution, it can modify files and run programs on your machine. Treat access to an MCPFS server like access to your terminal.

Only run it in environments you control. Only connect clients you trust. Review your configuration before starting the server.

You are responsible for what you expose.

## Features

* Explicit JSON-configured filesystem roots.
* Global config bootstrap from an embedded default config.
* Project-local config support through `.mcpfs/project.cfg.json`.
* Read access by default.
* Write access through explicit per-root `read_write` opt-in.
* Command execution through explicit `commands.mode` opt-in.
* `.gitignore` support.
* Additional include/exclude glob rules.
* Root escape protection.
* Symlink escape protection, including writable-path parent checks.
* File size limits for reads and writes.
* Timeout and output limits for command execution.
* Bounded directory listing and tree output.
* Structured JSON logging.
* STDIO transport for local MCP hosts and MCP Inspector.
* Streamable HTTP transport for remote MCP clients.
* Optional embedded ngrok tunnel for quick remote development.
* HTTP auth modes: `none`, `bearer`, and `oidc`.
* OIDC/JWT validation using JWKS, issuer, audience, expiry, not-before, and identity allowlists.
* CLI project setup and root management:

  * `mcpfs init`
  * `mcpfs project add`
  * `mcpfs project rm`
  * `mcpfs project ls`

## MCP tools

Filesystem tools:

| Tool              | Description                                                                                               |
| ----------------- | --------------------------------------------------------------------------------------------------------- |
| `fs_roots`        | List configured filesystem roots and their modes. Does not expose absolute host paths.                    |
| `fs_list`         | List files under a configured root. Honors explicit excludes and `.gitignore` rules.                      |
| `fs_tree`         | Return a bounded tree view with structured entries and compact text output.                               |
| `fs_read`         | Read a bounded file from a configured root.                                                               |
| `fs_read_lines`   | Read a 1-based inclusive line range from a file.                                                          |
| `fs_write`        | Create or replace a file under a configured `read_write` root. Honors excludes, `.gitignore`, and limits. |
| `fs_search`       | Search text files using a case-sensitive substring query.                                                 |
| `fs_search_regex` | Search text files using a regular expression query.                                                       |
| `fs_stat`         | Return metadata for a file or directory.                                                                  |

Git tools:

| Tool         | Description                                                                                                         |
| ------------ | ------------------------------------------------------------------------------------------------------------------- |
| `git_status` | Return `git status --porcelain=v1 -b` as structured JSON.                                                           |
| `git_diff`   | Return a diff for the whole root or a specific path. Supports staged diffs and synthetic diffs for untracked files. |
| `git_blame`  | Return read-only blame information for a file, optionally scoped to a 1-based inclusive line range.                 |
| `git_show`   | Return metadata and bounded patch output for a single commit, optionally scoped to a path.                          |
| `git_log`    | Return recent commit history, optionally scoped to a path.                                                          |

Project tools:

| Tool               | Description                                                                                      |
| ------------------ | ------------------------------------------------------------------------------------------------ |
| `project_overview` | Return a compact project summary: tree, important files, counts, git status, and recent commits. |

Command tools:

| Tool       | Description                                                                                       |
| ---------- | ------------------------------------------------------------------------------------------------- |
| `cmd_list` | List configured command IDs available through MCPFS command execution.                            |
| `cmd_run`  | Run a predefined command by configured command ID with fixed argv, timeout, and output limits.    |

`cmd_list` and `cmd_run` are registered when `commands.mode` is `predefined` or `unguarded`. `cmd_exec` for arbitrary command execution is planned for `unguarded` mode, but is not implemented yet.

## Permission model

`mcpfs` exposes only the roots and capabilities you configure.

Tool paths are always interpreted relative to a configured root. `mcpfs` rejects:

* absolute paths
* `..` root escapes
* symlink escapes
* explicitly excluded paths
* `.gitignore` ignored paths
* files larger than the configured root limit

Filesystem write access is opt-in per root. A root with:

```json
"mode": "read"
```

can be inspected but not written. A root with:

```json
"mode": "read_write"
```

allows `fs_write`, subject to the same root boundary, include/exclude, `.gitignore`, symlink, and size-limit checks.

Command execution is controlled by `commands.mode`:

| Mode         | Behavior |
| ------------ | -------- |
| `disabled`   | No command execution tools are registered. This is the default. |
| `predefined` | Registers `cmd_list` and `cmd_run`. Only configured command IDs can run. |
| `unguarded`  | Currently behaves like `predefined`. Future `cmd_exec` support will allow arbitrary command execution. Treat this mode like terminal access. |

Predefined commands use fixed argv arrays. No shell interpolation is used by `cmd_run`. Commands run from root-scoped working directories and return structured stdout, stderr, exit code, duration, timeout, and truncation metadata.

The HTTP transports support three auth modes:

* `none` — no HTTP authentication. Useful only for local-only setups or short-lived development tunnels.
* `bearer` — static bearer token loaded from an environment variable.
* `oidc` — JWT/OIDC validation using a configured issuer, audience, JWKS URL, and identity allowlist.

Do not expose `mcpfs` to the public internet without understanding what roots you configured, which roots are writable, which commands are available, and which auth mode is active.

## Installation

```bash
git clone https://github.com/tedla-brandsema/mcpfs.git
cd mcpfs

go test ./...
go build -o ./bin/mcpfs ./cmd/mcpfs
```

## Quick start

Create a project-local config in the current directory:

```bash
mcpfs init
```

Or initialize a project elsewhere:

```bash
mcpfs init -path /path/to/project
```

Add the current directory to the default global MCPFS config:

```bash
mcpfs project add
```

Add a specific project directory:

```bash
mcpfs project add -path /path/to/project
```

List configured project roots:

```bash
mcpfs project ls
```

Run the server with the default global config:

```bash
mcpfs
```

The default global config is created automatically when missing at:

```text
os.UserConfigDir()/mcpfs/mcpfs.cfg.json
```

The embedded default global config starts with STDIO transport and no roots:

```json
{
  "server": {
    "name": "mcpfs",
    "version": "0.3.0",
    "transport": "stdio",
    "auth": {
      "mode": "none"
    }
  },
  "roots": [],
  "commands": {
    "mode": "disabled"
  }
}
```

## CLI commands

### `mcpfs`

Run the MCP server.

```bash
mcpfs
mcpfs -config /path/to/mcpfs.cfg.json
```

If `-config` is omitted, `mcpfs` loads or creates the default global config at `os.UserConfigDir()/mcpfs/mcpfs.cfg.json`.

If `-config` is provided, only that explicit path is loaded.

### `mcpfs init`

Create `.mcpfs/project.cfg.json` for a project using the embedded default project registry config.

```bash
mcpfs init
mcpfs init -path /path/to/project
```

Flags:

| Flag    | Description                                           |
| ------- | ----------------------------------------------------- |
| `-path` | Project directory. Defaults to the current directory. |

`mcpfs init` does not add the project to any MCPFS server config. It only writes the project-local config if it does not already exist.

### `mcpfs project add`

Add a project root to an MCPFS config.

```bash
mcpfs project add
mcpfs project add -path /path/to/project
mcpfs project add -id my-project
mcpfs project add -cfg /path/to/mcpfs.cfg.json
mcpfs project add -path /path/to/project -id my-project -cfg /path/to/mcpfs.cfg.json
```

Flags:

| Flag    | Description                                                      |
| ------- | ---------------------------------------------------------------- |
| `-path` | Project directory. Defaults to the current directory.            |
| `-id`   | Root id to add. Defaults to the project directory name.          |
| `-cfg`  | MCPFS config path to update. Defaults to the global user config. |

The added root uses read mode, `**/*` includes, the standard sensitive-file excludes, `.gitignore` support, and the default max file size.

To allow writes, change the root's `mode` from `read` to `read_write` in the target config.

### `mcpfs project rm`

Remove a project root from an MCPFS config by root id.

```bash
mcpfs project rm
mcpfs project rm -path /path/to/project
mcpfs project rm -id my-project
mcpfs project rm -cfg /path/to/mcpfs.cfg.json
mcpfs project rm -path /path/to/project -id my-project -cfg /path/to/mcpfs.cfg.json
```

Flags:

| Flag    | Description                                                                          |
| ------- | ------------------------------------------------------------------------------------ |
| `-path` | Project directory used to derive the default root id. Defaults to current directory. |
| `-id`   | Root id to remove. Defaults to the project directory name.                           |
| `-cfg`  | MCPFS config path to update. Defaults to the global user config.                     |

`project rm` removes by id. If the project was added with a custom id, pass that id explicitly.

### `mcpfs project ls`

List project roots configured in an MCPFS config.

```bash
mcpfs project ls
mcpfs project ls -cfg /path/to/mcpfs.cfg.json
```

Flags:

| Flag   | Description                                                    |
| ------ | -------------------------------------------------------------- |
| `-cfg` | MCPFS config path to read. Defaults to the global user config. |

Output is a simple table with root id, mode, and path.

## STDIO usage

STDIO is useful for local MCP clients and MCP Inspector.

Run with the default global config:

```bash
mcpfs
```

Run with an explicit config:

```bash
./bin/mcpfs -config config.example.json
```

Or run directly:

```bash
go run ./cmd/mcpfs -config config.example.json
```

### MCP Inspector

```bash
bunx @modelcontextprotocol/inspector
```

Using the default global config:

```text
Transport Type:
STDIO

Command:
/path/to/mcpfs/bin/mcpfs
```

Using an explicit config:

```text
Transport Type:
STDIO

Command:
/path/to/mcpfs/bin/mcpfs

Arguments:
-config
/path/to/mcpfs/config.example.json
```

Or with `go run`:

```text
Transport Type:
STDIO

Command:
go

Arguments:
run
/path/to/mcpfs/cmd/mcpfs
-config
/path/to/mcpfs/config.example.json
```

## HTTP usage

HTTP transport is useful when the MCP client needs to connect to a network endpoint.

Example config using bearer auth:

```json
{
  "server": {
    "name": "mcpfs",
    "version": "0.3.0",
    "transport": "http",
    "addr": "127.0.0.1:8080",
    "path": "/mcp",
    "auth": {
      "mode": "bearer",
      "token_env": "MCPFS_TOKEN"
    }
  },
  "roots": [
    {
      "id": "project",
      "path": "/path/to/project",
      "mode": "read",
      "include": ["**/*"],
      "exclude": [
        "**/.git/**",
        "**/.env",
        "**/.env.*",
        "**/*secret*",
        "**/*credential*",
        "**/*.pem",
        "**/*.key"
      ],
      "use_gitignore": true,
      "max_file_bytes": 262144
    }
  ],
  "commands": {
    "mode": "disabled"
  }
}
```

Run:

```bash
export MCPFS_TOKEN="$(openssl rand -hex 32)"
./bin/mcpfs -config config.http.example.json
```

Health check:

```bash
curl http://127.0.0.1:8080/healthz
```

Authenticated MCP endpoint:

```bash
curl -i \
  -H "Authorization: Bearer $MCPFS_TOKEN" \
  http://127.0.0.1:8080/mcp
```

A plain `GET /mcp` may return a protocol-level MCP response such as `405 Method Not Allowed`. That is expected; it means the request reached the MCP handler.

## OIDC/JWT auth

`mcpfs` can run as an OIDC/JWT-protected MCP resource server.

It does not implement an OAuth authorization server. Instead, it validates bearer JWTs issued by an external identity provider such as Google Identity Platform, Firebase, Auth0, WorkOS, Zitadel, Keycloak, Dex, or Azure Entra ID.

Example config:

```json
{
  "server": {
    "name": "mcpfs",
    "version": "0.3.0",
    "transport": "http",
    "addr": "127.0.0.1:8080",
    "path": "/mcp",
    "auth": {
      "mode": "oidc",
      "issuer": "https://accounts.google.com",
      "audience": "mcpfs",
      "jwks_url": "https://www.googleapis.com/oauth2/v3/certs",
      "allowed_emails": ["you@example.com"]
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

OIDC validation checks:

* JWT signature using the configured JWKS URL.
* `iss`
* `aud`
* `exp`
* `nbf`, when present
* `allowed_emails` and/or `allowed_subjects`

At least one of `allowed_emails` or `allowed_subjects` must be configured.

Example using subject allowlisting instead of email allowlisting:

```json
"auth": {
  "mode": "oidc",
  "issuer": "https://issuer.example.com",
  "audience": "mcpfs",
  "jwks_url": "https://issuer.example.com/.well-known/jwks.json",
  "allowed_subjects": ["user-or-client-subject-id"]
}
```

## Embedded ngrok usage

`mcpfs` can start an ngrok tunnel automatically for quick remote development.

First, install/configure an ngrok account and create an authtoken.

Run:

```bash
export NGROK_AUTHTOKEN="<your-ngrok-authtoken>"
./scripts/run-ngrok.sh
```

The app logs a public MCP URL:

```text
https://example.ngrok-free.dev/mcp
```

Use that URL in your remote MCP client.

For short-lived development with ChatGPT Developer Mode, `config.ngrok.example.json` uses:

```json
"auth": {
  "mode": "none"
}
```

Only do this while actively testing, and only with carefully scoped roots and commands.

You can also run ngrok with bearer or OIDC auth by changing the `auth` block in the ngrok config.

## Configuration

There are two config layers:

| Config file               | Purpose                                                                                                     |
| ------------------------- | ----------------------------------------------------------------------------------------------------------- |
| `mcpfs.cfg.json`          | Server settings, configured roots, and configured commands. The default global path is `os.UserConfigDir()/mcpfs/mcpfs.cfg.json`. |
| `.mcpfs/project.cfg.json` | Project-local overview detection rules for a configured root.                                               |

### Global MCPFS config

The global config controls server settings, roots, and command execution.

Root config fields:

| Field            | Description                                                                   |
| ---------------- | ----------------------------------------------------------------------------- |
| `id`             | Stable root identifier used by MCP tools.                                     |
| `path`           | Local filesystem path to expose.                                              |
| `mode`           | `read` or `read_write`. `read_write` enables `fs_write` for that root.        |
| `include`        | Glob allowlist. Empty means all files not excluded are allowed.               |
| `exclude`        | Glob denylist.                                                                |
| `use_gitignore`  | Apply `.gitignore` rules as an additional deny/allow layer.                   |
| `max_file_bytes` | Maximum readable or writable file size. Defaults to `262144` when set to `0`. |

Command config fields:

| Field                                | Description |
| ------------------------------------ | ----------- |
| `commands.mode`                      | `disabled`, `predefined`, or `unguarded`. Defaults to `disabled`. |
| `commands.defaults.timeout_seconds`  | Default command timeout. Defaults to `60` when unset. |
| `commands.defaults.max_output_bytes` | Default combined stdout/stderr output limit. Defaults to `65536` when unset. |
| `commands.items[].id`                | Stable command id used by `cmd_run`. |
| `commands.items[].description`       | Optional human-readable description. |
| `commands.items[].root_id`           | Root id used to scope the command working directory. |
| `commands.items[].workdir`           | Relative working directory inside the root. Defaults to `.`. |
| `commands.items[].command`           | Fixed argv array to execute. The first item is the executable. |
| `commands.items[].timeout_seconds`   | Optional per-command timeout override. |
| `commands.items[].max_output_bytes`  | Optional per-command output limit override. |

Example predefined commands:

```json
"commands": {
  "mode": "predefined",
  "defaults": {
    "timeout_seconds": 60,
    "max_output_bytes": 65536
  },
  "items": [
    {
      "id": "test",
      "description": "Run all Go tests",
      "root_id": "mcpfs",
      "workdir": ".",
      "command": ["go", "test", "./..."],
      "timeout_seconds": 120
    },
    {
      "id": "fmt-go",
      "description": "Format Go source",
      "root_id": "mcpfs",
      "workdir": ".",
      "command": ["gofmt", "-w", "cmd", "internal"]
    }
  ]
}
```

Server config fields:

| Field                   | Description                                                           |
| ----------------------- | --------------------------------------------------------------------- |
| `name`                  | MCP server name.                                                      |
| `version`               | MCP server version.                                                   |
| `transport`             | `stdio`, `http`, or `http_ngrok`.                                     |
| `addr`                  | HTTP bind address. Defaults to `127.0.0.1:8080` for HTTP transports.  |
| `path`                  | MCP HTTP path. Defaults to `/mcp`.                                    |
| `auth.mode`             | HTTP auth mode: `none`, `bearer`, or `oidc`.                          |
| `auth.token_env`        | Environment variable containing the bearer token when using `bearer`. |
| `auth.issuer`           | Expected JWT issuer when using `oidc`.                                |
| `auth.audience`         | Expected JWT audience when using `oidc`.                              |
| `auth.jwks_url`         | JWKS URL used to verify JWT signatures when using `oidc`.             |
| `auth.allowed_emails`   | Optional email allowlist for OIDC-authenticated users.                |
| `auth.allowed_subjects` | Optional subject allowlist for OIDC-authenticated users.              |
| `ngrok_url`             | Optional reserved ngrok URL/domain.                                   |

Legacy `require_auth` and `auth_token_env` fields are still accepted for compatibility, but new configs should use `auth.mode`.

### Project-local config

A project-local config lives at:

```text
.mcpfs/project.cfg.json
```

It customizes project overview detection rules for that configured root. When present, it is merged with the global/user project registry for that root only. This prevents project-local rules from leaking across multi-root workspaces.

Default project-local config is embedded in the binary and can be written with:

```bash
mcpfs init
```

Project-local config shape:

```json
{
  "project": {
    "important_files": ["README.md", "TODO.md", "go.mod"],
    "source_extensions": [".go", ".ts", ".py"],
    "test_patterns": ["*_test.go", "*test*", "*spec*"],
    "documentation_extensions": [".md", ".rst", ".adoc", ".txt"],
    "documentation_files": ["README", "LICENSE", "COPYING"],
    "configuration_extensions": [".json", ".yaml", ".yml", ".toml", ".ini", ".env", ".xml"],
    "configuration_files": ["Dockerfile", "Makefile", "go.mod", "go.sum", "package.json"]
  }
}
```

Rule lists in the project-local file override the corresponding global/user project registry list when non-empty. Empty local lists inherit the base list.

### Auth examples

No auth:

```json
"auth": {
  "mode": "none"
}
```

Bearer auth:

```json
"auth": {
  "mode": "bearer",
  "token_env": "MCPFS_TOKEN"
}
```

OIDC auth:

```json
"auth": {
  "mode": "oidc",
  "issuer": "https://accounts.google.com",
  "audience": "mcpfs",
  "jwks_url": "https://www.googleapis.com/oauth2/v3/certs",
  "allowed_emails": ["you@example.com"]
}
```

## Example MCP workflows

Ask for the configured roots:

```json
{}
```

Call `fs_roots`.

List project files:

```json
{
  "root_id": "project",
  "path": ".",
  "recursive": false
}
```

Call `fs_list`.

View a bounded project tree:

```json
{
  "root_id": "project",
  "path": ".",
  "max_depth": 3,
  "max_entries": 300,
  "include_files": true
}
```

Call `fs_tree`.

Read a file:

```json
{
  "root_id": "project",
  "path": "internal/core/resolve.go"
}
```

Call `fs_read`.

Read a line range:

```json
{
  "root_id": "project",
  "path": "internal/core/resolve.go",
  "start_line": 1,
  "end_line": 80
}
```

Call `fs_read_lines`.

Write a file under a writable root:

```json
{
  "root_id": "project",
  "path": "notes/example.md",
  "content": "# Example\n",
  "create_dirs": true
}
```

Call `fs_write`. The root must be configured with `"mode": "read_write"`.

List configured commands:

```json
{}
```

Call `cmd_list`.

Run a predefined command:

```json
{
  "id": "test"
}
```

Call `cmd_run`.

Run a predefined command with tighter limits:

```json
{
  "id": "test",
  "timeout_seconds": 30,
  "max_output_bytes": 32768
}
```

Call `cmd_run`.

Search code with a substring:

```json
{
  "root_id": "project",
  "query": "ResolveInsideRoot",
  "glob": "**/*.go"
}
```

Call `fs_search`.

Search code with a regex:

```json
{
  "root_id": "project",
  "query": "func\\s+Resolve",
  "glob": "**/*.go",
  "case_sensitive": true,
  "max_results": 50
}
```

Call `fs_search_regex`.

Get a project overview:

```json
{
  "root_id": "project",
  "path": ".",
  "max_depth": 3,
  "max_entries": 500,
  "recent_commits": 5
}
```

Call `project_overview`.

Check working tree state:

```json
{
  "root_id": "project"
}
```

Call `git_status`.

Inspect a changed file:

```json
{
  "root_id": "project",
  "path": "internal/service/git/diff.go"
}
```

Call `git_diff`.

Inspect blame for a line range:

```json
{
  "root_id": "project",
  "path": "internal/service/project/overview.go",
  "start_line": 1,
  "end_line": 80
}
```

Call `git_blame`.

Inspect a commit:

```json
{
  "root_id": "project",
  "rev": "HEAD",
  "max_bytes": 65536
}
```

Call `git_show`.

Review history:

```json
{
  "root_id": "project",
  "limit": 5
}
```

Call `git_log`.

## Development

Run tests:

```bash
go test ./...
```

Build:

```bash
go build -o ./bin/mcpfs ./cmd/mcpfs
```

Run with STDIO and default global config:

```bash
./bin/mcpfs
```

Run with STDIO and explicit config:

```bash
./bin/mcpfs -config config.example.json
```

Run with HTTP bearer auth:

```bash
export MCPFS_TOKEN="$(openssl rand -hex 32)"
./bin/mcpfs -config config.http.example.json
```

Run with HTTP OIDC auth:

```bash
./bin/mcpfs -config config.oidc.example.json
```

Run with embedded ngrok:

```bash
export NGROK_AUTHTOKEN="<your-ngrok-authtoken>"
./scripts/run-ngrok.sh
```

## Status

`mcpfs` currently focuses on filesystem, Git, project-context, and predefined command execution tools. Roots are read-only by default. File writes are available through `fs_write` only for roots explicitly configured with `mode: "read_write"`. Command execution is disabled by default and available through `cmd_list`/`cmd_run` when `commands.mode` is `predefined` or `unguarded`.
