# Production hardening

Use this guide before running MCPFS in a production-like environment or exposing it to remote clients.

## Start from least privilege

Use the narrowest practical configuration:

- narrow roots;
- `mode: "read"` by default;
- `commands.mode: "disabled"` by default;
- local STDIO when possible;
- HTTP auth for any network transport.

## Review roots

Avoid configuring broad roots such as a home directory or a parent workspace that contains unrelated projects.

For each root, verify:

- it contains only files the client should access;
- include and exclude rules match your intent;
- `.gitignore` is enabled when useful;
- sensitive files are excluded;
- `max_file_bytes` is appropriate.

## Review writes

Use `read_write` only when the client must modify files.

Before enabling writes:

- start from a clean Git working tree;
- scope the root narrowly;
- test a small write;
- verify writes outside the root fail;
- review `git status` after writes.

## Review command execution

Prefer this order:

1. `commands.mode: "disabled"`
2. `commands.mode: "predefined"`
3. `commands.mode: "unguarded"` only in controlled environments

Do not expose unguarded command execution to untrusted or remote clients.

## Harden HTTP

For remote HTTP:

- use bearer or OIDC auth;
- use TLS or a trusted reverse proxy;
- bind only to the required interface;
- avoid logging secrets;
- rotate bearer tokens when needed;
- validate OIDC issuer, audience, expiry, signature, and identity allowlists.

## Review tunnels

Use ngrok only for short-lived development unless you have an explicit hardening plan.

Avoid tunnels with:

- no auth;
- writable roots;
- unguarded commands;
- long-lived public URLs.

## Operational checklist

Before relying on MCPFS:

- Run `go test ./...`.
- Review `CHANGELOG.md` for security-sensitive changes.
- Review `SECURITY.md` and `docs/security.md`.
- Verify the active config file.
- Verify roots and modes.
- Verify command mode.
- Verify HTTP auth.
- Test allowed and rejected access.
- Store secrets outside Git.
