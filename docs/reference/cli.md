# CLI reference

MCPFS provides a server command and project configuration helper commands.

## `mcpfs`

Run the MCP server.

```bash
mcpfs
mcpfs -config /path/to/mcpfs.cfg.json
```

If `-config` is omitted, MCPFS loads or creates the default global config at `os.UserConfigDir()/mcpfs/mcpfs.cfg.json`.

If `-config` is provided, only that explicit path is loaded.

## `mcpfs init`

Create `.mcpfs/project.cfg.json` for a project using the embedded default project registry config.

```bash
mcpfs init
mcpfs init -path /path/to/project
```

| Flag | Description |
| --- | --- |
| `-path` | Project directory. Defaults to the current directory. |

`mcpfs init` does not add the project to any MCPFS server config. It only writes the project-local config if it does not already exist.

## `mcpfs project add`

Add a project root to an MCPFS config.

```bash
mcpfs project add
mcpfs project add -path /path/to/project
mcpfs project add -id my-project
mcpfs project add -cfg /path/to/mcpfs.cfg.json
```

| Flag | Description |
| --- | --- |
| `-path` | Project directory. Defaults to the current directory. |
| `-id` | Root id to add. Defaults to the project directory name. |
| `-cfg` | MCPFS config path to update. Defaults to the global user config. |

The added root uses read mode, `**/*` includes, standard sensitive-file excludes, `.gitignore` support, and the default max file size.

To allow writes, change the root mode from `read` to `read_write` in the target config.

## `mcpfs project rm`

Remove a project root from an MCPFS config by root id.

```bash
mcpfs project rm
mcpfs project rm -path /path/to/project
mcpfs project rm -id my-project
mcpfs project rm -cfg /path/to/mcpfs.cfg.json
```

| Flag | Description |
| --- | --- |
| `-path` | Project directory used to derive the default root id. Defaults to current directory. |
| `-id` | Root id to remove. Defaults to the project directory name. |
| `-cfg` | MCPFS config path to update. Defaults to the global user config. |

`project rm` removes by id. If the project was added with a custom id, pass that id explicitly.

## `mcpfs project ls`

List project roots configured in an MCPFS config.

```bash
mcpfs project ls
mcpfs project ls -cfg /path/to/mcpfs.cfg.json
```

| Flag | Description |
| --- | --- |
| `-cfg` | MCPFS config path to read. Defaults to the global user config. |

Output is a table with root id, mode, and path.
