# Security

MCPlug is security-sensitive software: it executes the programs you configure and can expose their capabilities to remote clients. Read this page before enabling HTTP or ngrok.

## Trust boundaries

- **MCPlug does not sandbox upstream servers.** stdio children run with the same OS privileges as the MCPlug process. A malicious or compromised upstream server can do anything your user account can. Configure only servers you trust.
- **Connected clients get every aggregated tool.** If the filesystem server you configure can write files, every client that can reach the MCPlug endpoint can write files. Use `includeTools`/`excludeTools` to narrow what is exposed.
- **Commands are executed directly** with `exec` (argv as configured), never through a shell. There is no shell-expansion or injection surface in the config itself.

## Secrets

- `mcpServers.*.env` and `mcpServers.*.headers` values may contain credentials. MCPlug never logs these values, and redacts values whose keys look secret-like (`token`, `authorization`, `api_key`, `apikey`, `secret`, `password`, `bearer`) in any config output.
- Config files created by MCPlug use mode 0600. MCPlug warns at startup when a config containing `headers`/`env` values is world-readable; fix with `chmod 600`.
- Bearer tokens for MCPlug's own auth are read from the environment (`auth.token_env`), never stored in the config.

## Exposure guidance

In order of increasing risk:

1. **`stdio`** — no network surface. Default and safest.
2. **`http` on localhost** — reachable by local processes. Use bearer auth if other users share the machine.
3. **`http` on a non-localhost address** — only behind a TLS-terminating reverse proxy with auth.
4. **`http_ngrok`** — publicly reachable. **Always** configure `bearer` or `oidc` auth; the example configs in this repository do. Treat the printed public URL as sensitive.

## Operational notes

- A crashing upstream is restarted with backoff; during restart its tools return tool errors. This is not an auth bypass — the endpoint auth is unaffected.
- `plug ls` probes upstreams (it spawns the configured commands) but never starts the HTTP listener or tunnel.
- Upstream calls are bounded by a 60-second timeout to prevent a hung upstream from pinning client connections indefinitely.

## Reporting

See [SECURITY.md](../SECURITY.md) in the repository root for vulnerability reporting.
