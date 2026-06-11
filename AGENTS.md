# AGENTS.md

MCPlug is a pure MCP aggregating gateway (see README.md). Agents working here follow these rules. This file is authoritative for the repo.

## Non-negotiables

**Unit tests are first class.** Every behavior change ships with tests in the same change, at production level: table tests for pure logic, integration tests for lifecycle behavior (the self-exec MCP child pattern in `internal/upstream/supervisor_test.go` is the reference). Supervisor, config validation, sanitizer, filtering, and error-mapping changes are untested only if you can argue why in the PR. `go test ./...` and `go vet ./...` must pass; supervisor changes must also pass `go test -race ./internal/upstream/`.

**Logging is first class.** All logging goes through `log/slog` with structured fields — no `fmt.Print*` outside CLI user output. State transitions, upstream lifecycle events, and proxied calls (latency + failure category) are logged with stable field names (`upstream`, `tool`, `from`, `to`, `error`). Never log header values, env values, or anything matching the secret-key list in `internal/config/redact.go`; route any config echo through `RedactValue`.

## Architecture boundaries

- `internal/config` — schema, validation, sanitizer, redaction. No MCP SDK imports.
- `internal/upstream` — upstream clients, `StartAll`, supervisor. Owns upstream startup and the initial tool listing, exclusively.
- `internal/gateway` — aggregator, HTTP handler, ngrok. `BuildServer` consumes `StartupResult`; it must never create, start, or list upstreams.
- `internal/auth` — authenticators only.
- `cmd/plug` — CLI wiring, kept thin.

## Contracts to preserve

- `Upstream.CallTool`: expected upstream/runtime failures return `(CallToolResult{IsError: true}, nil)`; Go errors are reserved for gateway invariants. Documented on the interface — do not blur it.
- Tool names: always `<server>_<tool>`, sanitizer in `internal/config/sanitize.go` is the single source of truth.
- `optional` affects startup failure only; `disabled` means ignored; enabled servers are required by default.
- Commands run verbatim via `exec.CommandContext`, never through a shell. The SDK `CommandTransport` owns `Start`/`Close`; factories hand it unstarted `exec.Cmd`s.
- `plug ls` never starts the transport, HTTP listener, or ngrok tunnel.

## Workflow

- `gofmt -w .` before committing; CI checks format, vet, tests, build.
- Config files may contain secrets: write them 0600, keep the world-readable warning intact.
- Behavior changes update docs (`docs/`, README) and `CHANGELOG.md` in the same change; trust-boundary changes update `docs/security.md`.
- Deferred work goes in `TODO.md`, not code comments.
- Pre-0.5 MCPFS lives on `legacy/v1`; do not resurrect roots/commands/native tools.
