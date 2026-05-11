# MCPFS

MCPFS is a Model Context Protocol server that gives trusted MCP clients controlled access to local project files, Git metadata, project summaries, optional writes, and optional command execution.

## What MCPFS is

MCPFS runs on your machine or in an environment you control. It exposes explicitly configured project roots to MCP clients through filesystem, Git, project overview, and command tools.

Use it when an MCP client needs live project context without uploading a repository or manually copying files into chat.

## Who should use it

MCPFS is for developers and tool builders who want to connect MCP clients to local or self-hosted project workspaces.

It is a good fit when you want to:

- inspect files and project trees from an MCP client;
- read bounded files and line ranges;
- search source code;
- inspect Git status, diffs, commits, and blame;
- get compact project summaries;
- optionally write files inside explicitly writable roots;
- optionally expose predefined development commands.

## Why it exists

AI coding tools work better when they can see the actual project state. MCPFS provides that context through bounded, configured access instead of broad filesystem access or copied snippets.

## Core features

- JSON-configured filesystem roots.
- Read-only roots by default.
- Per-root `read_write` opt-in for file writes.
- `.gitignore`, include, and exclude rule support.
- Root escape and symlink escape protection.
- Bounded file reads, writes, listings, trees, and search output.
- Git status, diff, log, show, and blame inspection.
- Compact project overview tool.
- Command execution modes: `disabled`, `predefined`, and `unguarded`.
- STDIO transport for local MCP clients.
- Streamable HTTP transport for remote MCP clients.
- HTTP auth modes: `none`, `bearer`, and `oidc`.
- Optional embedded ngrok tunnel for development testing.

## Security warning

MCPFS is a power tool.

When configured with writable roots or command execution, connected MCP clients can modify files and run programs on your machine. Treat access to an MCPFS server like access to your terminal.

Start with local read-only roots. Enable writes, HTTP exposure, ngrok, predefined commands, or unguarded command execution only after you understand the trust boundary.

Do not expose MCPFS to untrusted networks without authentication, narrow roots, and a reviewed configuration.

Read the [security guide](docs/security.md) before using write access, HTTP transport, OIDC, ngrok, or command execution.

## Install

Clone, test, and build from source:

```bash
git clone https://github.com/tedla-brandsema/mcpfs.git
cd mcpfs

go test ./...
go build -o ./bin/mcpfs ./cmd/mcpfs
```

The module path is:

```text
github.com/tedla-brandsema/mcpfs
```

## Quick start

Create a project-local config in the current directory:

```bash
./bin/mcpfs init
```

Add the current directory as a read-only root in the default global MCPFS config:

```bash
./bin/mcpfs project add
```

List configured roots:

```bash
./bin/mcpfs project ls
```

Start MCPFS with the default STDIO transport:

```bash
./bin/mcpfs
```

Connect a local MCP client or MCP Inspector to the `./bin/mcpfs` command.

For a complete first-run path, see the [quick start](docs/quick-start.md).

## Examples

Canonical examples live in this repository:

- [Local read-only](examples/local-read-only/) — inspect a project without writes or command execution.
- [Local read/write](examples/local-read-write/) — enable controlled writes inside a narrow root.
- [Predefined commands](examples/predefined-commands/) — expose a small allowlist of development commands.
- [HTTP bearer](examples/http-bearer/) — run HTTP transport with a bearer token.
- [OIDC](examples/oidc/) — validate JWTs from an external identity provider.
- [ngrok development](examples/ngrok-dev/) — test remote connectivity through a short-lived tunnel.

See [examples](examples/) for the full list.

## Documentation

Start here:

- [Documentation index](docs/index.md)
- [Quick start](docs/quick-start.md)
- [Security](docs/security.md)
- [Architecture](docs/architecture.md)
- [Configuration](docs/configuration.md)
- [Transports](docs/transports.md)
- [Commands](docs/commands.md)
- [Client recipes](docs/client-recipes.md)
- [Release policy](docs/release-policy.md)

Reference docs:

- [CLI reference](docs/reference/cli.md)
- [Configuration schema](docs/reference/config-schema.md)
- [MCP tools](docs/reference/tools.md)
- [Errors](docs/reference/errors.md)

## Maturity and compatibility

MCPFS is currently **Beta**.

It is useful and tested, but pre-v1 configuration, transport settings, command settings, tool details, and documentation structure may still change. Security-sensitive changes are documented in the changelog and release notes.

See the [release policy](docs/release-policy.md) for the compatibility boundary and v1 readiness criteria.

## Contributing

Contributions are welcome. Start with [CONTRIBUTING.md](CONTRIBUTING.md).

Security-sensitive changes should include documentation updates and tests where practical.

## License

MCPFS is licensed under the terms in [LICENSE](LICENSE).
