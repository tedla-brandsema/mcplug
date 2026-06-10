# Security

MCPFS is a power tool.

It spawns the MCP servers you configure and exposes their tools — potentially to remote clients. Treat access to an MCPFS endpoint like access to everything its configured upstream servers can do.

## High-risk configuration

The following settings intentionally grant powerful capabilities:

* upstream servers with write, delete, or execute tools (MCPFS does not sandbox them; stdio children run with MCPFS's own OS privileges)
* HTTP or ngrok transports exposed outside your local machine
* `auth.mode: "none"` on network transports

## Recommendations

* Configure only upstream MCP servers you trust; they run as your user.
* Connect only MCP clients you trust: every client gets every aggregated tool.
* Use `includeTools`/`excludeTools` to narrow what each upstream exposes.
* Use bearer or OIDC auth for HTTP transports; never expose `auth.mode: "none"` to untrusted networks.
* Keep config files private (`chmod 600`): `headers` and `env` values may contain secrets. MCPFS never logs these values and warns when such a config is world-readable.
* Review config files before starting the server; commands run verbatim via `exec` (never through a shell).

See [docs/security.md](docs/security.md) for the full trust-boundary discussion.

## Reporting vulnerabilities

Please report security issues privately through GitHub Security Advisories if available, or contact the maintainer directly.

Do not open public issues for vulnerabilities that could put users at risk.
