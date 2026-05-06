# MCPFS TODO

## DONE

* [x] Decide how OIDC JWKS/backend failures should map to HTTP responses.
  Currently `Authenticate` wraps JWKS loading failures as a plain error:

  ```go
  return nil, fmt.Errorf("load jwks: %w", err)
  ```
  The HTTP layer currently maps non-`UnauthorizedError` auth failures to `500`. That may be correct for an unavailable auth backend, while malformed/invalid tokens remain `401`.
  Covered by `internal/mcpfs/http_test.go`.

* [x] Add `fs_tree`.
  Return a bounded tree view for a configured root/path.
  Should honor the same include/exclude, `.gitignore`, symlink, and root escape rules as `fs_list`.

  Suggested input:

  ```json
  {
    "root_id": "project",
    "path": ".",
    "max_depth": 3,
    "max_entries": 300,
    "include_files": true
  }
  ```

  Suggested output should include both structured tree data and a compact text rendering for LLM-friendly context.

* [x] Add `git_show`.
  Return metadata and diff/body for a specific commit, optionally scoped to a path.

  Suggested input:

  ```json
  {
    "root_id": "project",
    "rev": "HEAD~1",
    "path": "optional/path.go",
    "max_bytes": 65536
  }
  ```

  Validate revision arguments carefully.
  Avoid shell interpolation.
  Deny ambiguous or option-like revisions where appropriate.
  Use explicit git argument separation.

* [x] Add tests for `git_show`.
  Cover:

  * normal commit lookup
  * path-scoped commit lookup
  * max byte truncation
  * invalid revision
  * revision/path argument safety
  * behavior outside a git repository

* [x] Add tests for `fs_tree`.
  Cover:

  * max depth
  * max entries
  * file inclusion/exclusion
  * `.gitignore` handling
  * explicit exclude handling
  * symlink escape protection
  * root escape rejection
  * deterministic ordering

* [x] Add README documentation for `fs_tree` and `git_show`.
  Include the tools in the MCP tools table and add example workflow snippets.

* [x] Add a shared truncation/result-limit helper.
  Several tools already need bounded output.
  Centralize max bytes / max entries behavior so future tools behave consistently.

## IN PROGRESS

* [ ] Add a shared path-scoped git helper.
  `git_diff`, `git_log`, and `git_show` should use a common helper for:

  * validating optional paths
  * resolving paths inside roots
  * appending `-- path`
  * handling git command limits/errors

## BACKLOG

* [ ] Add `git_blame`.
  Return blame information for a file, optionally scoped to a line range.

  Suggested input:

  ```json
  {
    "root_id": "project",
    "path": "internal/auth/oidc.go",
    "start_line": 40,
    "end_line": 90,
    "max_bytes": 65536
  }
  ```

  This is still generic and read-only, but lower priority than `git_show`.

* [ ] Add `fs_search_regex`.
  Add regex search as a separate tool rather than overloading substring search.

  Suggested input:

  ```json
  {
    "root_id": "project",
    "query": "func .*Authenticate",
    "glob": "**/*.go",
    "case_sensitive": false,
    "max_results": 100
  }
  ```

  Include safeguards for invalid regexes and expensive searches.

* [ ] Add `fs_read_lines`.
  Read a specific line range from a file.

  Suggested input:

  ```json
  {
    "root_id": "project",
    "path": "internal/auth/oidc.go",
    "start_line": 40,
    "end_line": 90
  }
  ```

  Useful when diagnostics, search results, or blame results point to specific ranges.

* [ ] Add stable location/range types.
  Introduce shared internal/output types for:

  * root id
  * relative path
  * line/character positions
  * ranges
  * diagnostics
  * symbols
  * references

  These will be useful before adding LSP support.

* [ ] Add `project_overview`.
  Return a compact generic summary of a root:

  * root id
  * top-level tree
  * git status summary
  * recent commits
  * detected important files
  * detected package/module files
  * test/config/documentation file counts

  This should be heuristic and language-agnostic.

* [ ] Add tool result metadata consistency.
  Standardize fields such as:

  * `root_id`
  * `path`
  * `truncated`
  * `max_bytes`
  * `max_entries`
  * `duration_ms`
  * `warnings`

* [ ] Add optional command execution framework for explicitly allowed read-only/dev commands.
  Keep disabled by default.

  This is preparation for tools such as test summaries or external analyzers.
  Config should require explicit allowlisting.

  Example:

  ```json
  {
    "commands": {
      "enabled": true,
      "allow": [
        "go test ./...",
        "npm test",
        "cargo test"
      ]
    }
  }
  ```

* [ ] Add an LSP host configuration model.
  LSP support should be generic and not tied to language-specific plugins.

  Example:

  ```json
  {
    "lsp": {
      "enabled": true,
      "servers": [
        {
          "id": "gopls",
          "languages": ["go"],
          "command": ["gopls"],
          "roots": ["project"]
        }
      ]
    }
  }
  ```

* [ ] Add generic LSP process lifecycle management.
  Handle:

  * server startup
  * initialization
  * shutdown
  * timeouts
  * crashes
  * logging
  * per-root workspace folders

* [ ] Add generic LSP diagnostics tool.
  Expose diagnostics without making the caller care which language server produced them.

  Suggested tool: `lsp_diagnostics`

  Suggested input:

  ```json
  {
    "root_id": "project",
    "path": "optional/file.go",
    "max_results": 100
  }
  ```

  Output should normalize LSP diagnostics into MCPFS location/range types.

* [ ] Add generic LSP document symbols tool.
  Suggested tool: `lsp_document_symbols`

  Suggested input:

  ```json
  {
    "root_id": "project",
    "path": "internal/auth/oidc.go"
  }
  ```

* [ ] Add generic LSP definition/references/hover tools.
  Suggested tools:

  * `lsp_definition`
  * `lsp_references`
  * `lsp_hover`

  These should use the same normalized location/range model as diagnostics.

* [ ] Decide how MCPFS should handle LSP file synchronization.
  Options:

  * initialize with real workspace paths and let the LSP read from disk
  * send opened document contents explicitly
  * hybrid approach

  For local developer use, real workspace paths are probably acceptable when LSP is explicitly enabled.

* [ ] Add plugin protocol design document.
  Define the long-term plugin system separately from LSP.

  Cover:

  * external plugin processes
  * protobuf/gRPC boundary
  * lifecycle
  * capabilities
  * trust model
  * root/path handling
  * output limits
  * version negotiation

* [ ] Add protobuf definitions for plugin API.
  Start with lifecycle/capability discovery only.

  Suggested services:

  * `PluginLifecycle`
  * `PluginCapabilities`

* [ ] Add plugin host skeleton.
  Support explicit plugin config, process startup, handshake, health check, and shutdown.
  Do not expose plugin tools yet.

* [ ] Add first plugin-backed tool as a proof of concept.
  Keep it small and non-critical.
  Prefer something generic or diagnostic-only before adding language-specific behavior.

* [ ] Add language-specific plugin SDKs.
  Start with Go SDK.
  Keep the wire protocol language-neutral through protobuf/gRPC.
