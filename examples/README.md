# MCPFS examples

These examples show common MCPFS operating modes. Start with local read-only access, then enable higher-risk capabilities only when you need them.

## Choose an example

| Example | Use when | Risk level |
| --- | --- | --- |
| [Local read-only](local-read-only/) | You want a local MCP client to inspect a project. | Lowest |
| [Local read/write](local-read-write/) | You want a trusted local client to modify files inside a narrow root. | Medium |
| [Predefined commands](predefined-commands/) | You want a client to run a small allowlist of known commands. | Medium |
| [HTTP bearer](http-bearer/) | You need HTTP transport with a shared token. | Medium |
| [OIDC](oidc/) | You need HTTP transport with identity-provider-backed JWT validation. | Medium to high |
| [ngrok development](ngrok-dev/) | You need a short-lived public development tunnel. | High |

## Safety notes

- Keep roots narrow.
- Use `mode: "read"` unless writes are required.
- Keep `commands.mode: "disabled"` unless command execution is required.
- Prefer predefined commands over unguarded commands.
- Require bearer or OIDC auth for HTTP and tunnels.
- Do not commit real tokens, private paths, or credentials.

## Related docs

- [Quick start](../docs/quick-start.md)
- [Security](../docs/security.md)
- [Configuration](../docs/configuration.md)
- [Transports](../docs/transports.md)
- [Commands](../docs/commands.md)
