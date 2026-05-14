# MCP tools reference

MCPFS registers tools based on the configured capabilities.

## Filesystem tools

| Tool | Description |
| --- | --- |
| `fs_roots` | List configured filesystem roots and their modes. Does not expose absolute host paths. |
| `fs_list` | List files under a configured root. Honors explicit excludes and `.gitignore` rules. |
| `fs_tree` | Return a bounded tree view with structured entries and compact text output. |
| `fs_read` | Read a bounded file from a configured root. |
| `fs_read_lines` | Read a 1-based inclusive line range from a file. |
| `fs_hash` | Return SHA-256, size, mtime, and mode metadata for a file. Useful for guarded writes and patches. |
| `fs_write` | Create or replace a file under a configured `read_write` root. Honors excludes, `.gitignore`, symlink checks, and limits. Supports optional `expected_sha256`. |
| `fs_patch` | Apply exact old/new text replacements under a configured `read_write` root. Each old block must match exactly once. Supports dry-run previews and optional `expected_sha256`. |
| `fs_search` | Search text files using a case-sensitive substring query. |
| `fs_search_regex` | Search text files using a regular expression query. |
| `fs_stat` | Return metadata for a file or directory. |

## Git tools

| Tool | Description |
| --- | --- |
| `git_status` | Return `git status --porcelain=v1 -b` as structured JSON. |
| `git_diff` | Return a diff for the whole root or a specific path. Supports staged diffs and synthetic diffs for untracked files. |
| `git_blame` | Return read-only blame information for a file, optionally scoped to a 1-based inclusive line range. |
| `git_show` | Return metadata and bounded patch output for a single commit, optionally scoped to a path. |
| `git_log` | Return recent commit history, optionally scoped to a path. |

## Project tools

| Tool | Description |
| --- | --- |
| `project_overview` | Return a compact project summary: tree, important files, counts, Git status, and recent commits. |

## Command tools

Command tools are registered only when command execution is enabled.

| Tool | Registered when | Description |
| --- | --- | --- |
| `cmd_list` | `commands.mode` is `predefined` or `unguarded` | List configured command IDs. |
| `cmd_run` | `commands.mode` is `predefined` or `unguarded` | Run a predefined command by configured command ID. |
| `cmd_exec` | `commands.mode` is `unguarded` | Run an arbitrary argv command from a root-scoped working directory. |

## Common arguments

Most tools that access a root require `root_id`. File and directory tools also use a root-relative `path`.

Paths are interpreted relative to the configured root. Absolute paths and root escapes are rejected.

### Guarded filesystem edits

`fs_hash` can be used before editing to capture the current file identity:

```json
{
  "root_id": "project",
  "path": "README.md"
}
```

Use the returned `sha256` as `expected_sha256` on `fs_write` or `fs_patch` to prevent editing a file version the caller did not inspect:

```json
{
  "root_id": "project",
  "path": "README.md",
  "expected_sha256": "<sha256 from fs_hash>",
  "edits": [
    {
      "old": "old text",
      "new": "new text"
    }
  ]
}
```

If the file changed in between, MCPFS rejects the write or patch instead of applying it to a stale file.

## Related docs

- [Security](../security.md)
- [Configuration](../configuration.md)
- [Commands](../commands.md)
- [Errors](errors.md)
