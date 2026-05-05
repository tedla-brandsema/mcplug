# mcpfs

`mcpfs` is a small Model Context Protocol filesystem server.

It exposes explicitly configured filesystem roots to MCP clients using a JSON configuration file. Each root has an access mode, include/exclude rules, optional `.gitignore` support, and file size limits.

The initial version is intentionally read-only.

## Goals

* Expose selected folders to MCP clients.
* Keep filesystem authority explicit and scoped.
* Never allow ambient filesystem access.
* Support `.gitignore` so project defaults are respected.
* Keep extra excludes available for secrets and local-only files.
* Log every allowed and denied operation.

## Tools

* `fs_roots`
* `fs_list`
* `fs_read`
* `fs_search`
* `fs_stat`

## Usage

```bash
go mod tidy
go test ./...
go run ./cmd/mcpfs -config config.example.json
```

## Configuration

```json
{
  "server": {
    "name": "mcpfs",
    "version": "0.1.0",
    "transport": "stdio"
  },
  "roots": [
    {
      "id": "project",
      "path": "/path/to/project",
      "mode": "read",
      "include": ["**/*.go", "**/*.md", "**/*.json"],
      "exclude": ["**/.env", "**/*.pem", "**/*.key"],
      "use_gitignore": true,
      "max_file_bytes": 262144
    }
  ]
}
```

## Security model

Tool paths are always interpreted relative to a configured root.

`mcpfs` rejects:

* absolute paths
* `..` root escapes
* symlink escapes
* explicitly excluded paths
* `.gitignore` ignored paths
* files larger than the configured root limit

`.gitignore` support is an additional policy layer. It never weakens the root boundary.

## Notes

The `mode` field already accepts `read_write`, but v0.1 does not expose write tools.
