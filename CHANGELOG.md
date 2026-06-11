# Changelog

## v0.5.0

### Breaking

* The project is renamed **MCPFS → MCPlug**: binary `mcpfs` → `plug`, module path `github.com/tedla-brandsema/mcpfs` → `github.com/tedla-brandsema/mcplug`, global config `~/.config/mcpfs/mcpfs.cfg.json` → `~/.config/mcplug/mcplug.cfg.json`.
* MCPFS is now a pure **MCP aggregating gateway**. All native tools are removed: `fs_*`, `git_*`, `project_overview`, `cmd_list`, `cmd_run`, `cmd_exec`.
* The `roots` and `commands` config sections are removed, along with the `mcpfs project add|rm|ls` CLI and project-local `.mcpfs/project.cfg.json` configs.
* Pre-0.5 MCPFS behavior is preserved on the `legacy/v1` branch. See the README migration section for mapping MCPFS roots onto reference servers.

### Added

* `mcpServers` config map (Claude/Cursor-compatible shape) describing upstream MCP servers: stdio entries (`command`/`args`/`env`) and streamable-HTTP entries (`url`), with MCPFS extensions `headers`, `cwd`, `disabled`, `optional`, `includeTools`, and `excludeTools`.
* Tool aggregation: every upstream tool is exposed as `<server>_<tool>` (server names sanitized to `[A-Za-z0-9_-]`, collision-checked at config validation).
* Supervised stdio children: explicit lifecycle state machine, restart with exponential backoff on unexpected exit, healthy-period backoff reset, graceful shutdown on SIGINT/SIGTERM.
* Required-by-default upstreams: a failing enabled upstream aborts startup unless marked `optional`.
* 60-second timeout on upstream connect/list/call; upstream failures (restarting, unreachable, timed out) surface as MCP tool errors, not protocol failures.
* `mcpfs ls`: probes all configured servers and lists exposed/original/filtered tool names without starting any transport; exits non-zero on required upstream failure.
* `mcpfs init` now writes a 0600 starter config with disabled example entries.
* Secret hygiene: header/env values are never logged, secret-like keys are redacted in config output, and world-readable configs containing such values produce a startup warning.

### Changed

* `github.com/modelcontextprotocol/go-sdk` upgraded v0.8.0 → v1.4.1.
* Transports (`stdio`, `http`, `http_ngrok`) and auth modes (`none`, `bearer`, `oidc`) are unchanged and now serve the aggregated endpoint.

## v0.4.0

### Added

* Opt-in filesystem write support through `fs_write` for roots configured with `mode: "read_write"`.
* Writable path resolution that checks existing parent symlinks before creating new files.
* Command execution framework with explicit modes:
  * `disabled`
  * `predefined`
  * `unguarded`
* `cmd_list` for listing configured command IDs.
* `cmd_run` for running predefined commands by ID.
* `cmd_exec` for arbitrary argv command execution in `unguarded` mode.
* Command execution timeouts, output limits, stdout/stderr capture, exit code metadata, timeout metadata, truncation metadata, and structured logs.
* Project-local config support through `.mcpfs/project.cfg.json`.
* Global MCPFS config bootstrap from an embedded default config.
* `mcpfs init` for writing project-local config.
* `mcpfs project add`, `mcpfs project rm`, and `mcpfs project ls` for managing configured roots.

### Changed

* MCPFS is now positioned as a power-user local MCP workbench rather than only a read-only filesystem bridge.
* README now documents the power-tool warning, write access, command execution modes, `cmd_exec`, and release-era usage examples.
* Default/global config version is now `0.4.0`.

### Security

* Added explicit warning language for write and command execution capabilities.
* Command execution is disabled by default.
* `cmd_exec` is registered only when `commands.mode` is `unguarded`.
* `cmd_run` and `cmd_exec` execute argv arrays directly and do not perform shell interpolation unless the configured/client-provided argv explicitly invokes a shell.

## v0.3.0

### Added

* `fs_tree` for bounded tree output.
* `fs_read_lines` for line-range reads.
* `fs_search_regex` for regex-based search.
* `git_show` for commit inspection.
* `git_blame` for blame inspection.
* `project_overview` for compact project summaries.
* Tool result metadata consistency across services.
* Global config bootstrap and project overview registry support.
