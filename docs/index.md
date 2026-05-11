# MCPFS documentation

MCPFS is a Model Context Protocol server that gives trusted MCP clients controlled access to local project files, Git metadata, project summaries, optional writes, and optional command execution.

Use these docs to install MCPFS, choose a safe configuration, connect clients, and understand the security boundary before enabling powerful features.

## Recommended reading order

1. [Quick start](quick-start.md) — run MCPFS locally with a safe read-only setup.
2. [Security](security.md) — understand trust boundaries before enabling writes, HTTP, ngrok, or command execution.
3. [Configuration](configuration.md) — configure roots, access modes, auth, transports, and commands.
4. [Transports](transports.md) — choose STDIO, HTTP, bearer auth, OIDC, or ngrok development transport.
5. [Commands](commands.md) — configure predefined commands or understand unguarded command execution.
6. [Examples](../examples/) — run scenario-focused examples from a fresh clone.
7. [Release policy](release-policy.md) — understand maturity, compatibility, and v1 readiness.

## Choose your path

### Local-only users

Start with:

- [Quick start](quick-start.md)
- [Configure local read-only access](how-to/configure-local-read-only.md)
- [Configure local read/write access](how-to/configure-local-read-write.md)

Local-only STDIO is the recommended starting point because it avoids network exposure.

### Remote HTTP users

Start with:

- [Security](security.md)
- [Transports](transports.md)
- [Configure bearer authentication](how-to/configure-bearer-auth.md)
- [Configure OIDC authentication](how-to/configure-oidc.md)
- [Production hardening](advanced/production-hardening.md)

Do not expose MCPFS over HTTP without reviewing roots, write access, command modes, and authentication.

### Command users

Start with:

- [Commands](commands.md)
- [Add predefined commands](how-to/add-predefined-commands.md)
- [Command execution](advanced/command-execution.md)

Prefer `commands.mode: "predefined"` over `commands.mode: "unguarded"`.

### Client integrators

Start with:

- [Client recipes](client-recipes.md)
- [Transports](transports.md)
- [MCP tools reference](reference/tools.md)

## Reference

- [CLI reference](reference/cli.md)
- [Configuration schema](reference/config-schema.md)
- [MCP tools reference](reference/tools.md)
- [Error reference](reference/errors.md)

## Maintainers and contributors

- [Architecture](architecture.md)
- [Security](security.md)
- [Release policy](release-policy.md)
- [Contributing guide](../CONTRIBUTING.md)
