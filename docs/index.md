# MCPFS documentation

MCPFS is an MCP aggregating gateway: it runs/connects to the MCP servers configured under `mcpServers` and exposes all of their tools through one MCP endpoint, over stdio, localhost HTTP, or HTTP + ngrok with bearer/OIDC auth.

- [Quick start](quick-start.md) — first run with the reference filesystem and git servers.
- [Configuration](configuration.md) — full `server` and `mcpServers` field reference.
- [Transports](transports.md) — stdio, HTTP, ngrok, and authentication modes.
- [Security](security.md) — trust boundaries, secret handling, exposure guidance. Read this before enabling HTTP or ngrok.

Documentation for MCPFS v1 (the native filesystem/git/command tool server) lives on the `legacy/v1` branch. Migration notes are in the [README](../README.md#migration-from-v1).
