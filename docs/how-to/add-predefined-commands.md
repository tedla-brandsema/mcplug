# Add predefined commands

Use this guide when a trusted MCP client should run a small allowlist of known project commands.

## Before you begin

Decide which commands are safe for MCPFS to run. Prefer commands with predictable side effects, such as tests, format checks, lint checks, or local builds.

## Configure command mode

Set `commands.mode` to `predefined`:

```json
"commands": {
  "mode": "predefined",
  "defaults": {
    "timeout_seconds": 60,
    "max_output_bytes": 65536
  },
  "items": []
}
```

## Add a command

Add a command item:

```json
{
  "id": "test",
  "description": "Run all Go tests",
  "root_id": "project",
  "workdir": ".",
  "command": ["go", "test", "./..."],
  "timeout_seconds": 120
}
```

The `command` field is an argv array. MCPFS does not use shell interpolation unless you explicitly invoke a shell in the argv array.

## Run the server

```bash
mcpfs -config /path/to/mcpfs.cfg.json
```

## Verify the configuration

From your MCP client:

1. Call `cmd_list`.
2. Confirm the configured command id appears.
3. Call `cmd_run` with the command id.
4. Confirm an undefined command id is rejected.

## Troubleshoot

If the command does not appear, check `commands.mode` and `commands.items`.

If the command fails, check the `root_id`, `workdir`, executable name, timeout, output limit, and environment available to the MCPFS process.

## Next steps

- Read [Commands](../commands.md).
- Read [Command execution](../advanced/command-execution.md).
- Try the [predefined commands example](../../examples/predefined-commands/).
