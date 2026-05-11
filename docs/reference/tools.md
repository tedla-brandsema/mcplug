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
| `fs_write` | Create or replace a file under a configured `read_write` root. Honors excludes, `.gitignore`, symlink checks, and limits. |
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

## Related docs

- [Security](../security.md)
- [Configuration](../configuration.md)
- [Commands](../commands.md)
- [Errors](errors.md)
