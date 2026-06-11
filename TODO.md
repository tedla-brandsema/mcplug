# MCPlug TODO

MCPlug is a pure MCP aggregating gateway. The old MCPFS TODO list (native fs/git/command tools) is preserved on the `legacy/v1` branch and no longer applies.

## Deferred from the gateway plan

* [ ] Managed runtime download: auto-install `bun`/`uv` into the MCPlug data dir when a config references `npx`/`bunx`/`uvx` and the runtime is missing from PATH (opt-in convenience; commands still run verbatim by default).
* [ ] Dynamic tool-list refresh: register a per-upstream `ToolListChangedHandler`, re-list on change, and mutate the live aggregate with `AddTool`/`RemoveTools`. Requires leaving `DisableStandaloneSSE` unset for HTTP upstreams. Until then the tool list is a startup snapshot.
* [ ] User-facing per-upstream timeout config (currently a fixed 60s `upstream.DefaultTimeout`).
* [ ] Optional `Stateless`/`JSONResponse` flags for the streamable HTTP handler (some remote clients prefer stateless mode).
* [ ] Proxy upstream resources and prompts in addition to tools.
* [ ] Per-tool aliasing / richer policy (the routing map in `internal/gateway/aggregate.go` is built to keep the naming policy swappable).

## Ideas

* [ ] `plug ls --json` for machine-readable smoke tests.
* [ ] Health endpoint reporting per-upstream supervisor state.
* [ ] Reconnect-on-demand for optional upstreams that failed at startup (would lift the restart-required limitation).
