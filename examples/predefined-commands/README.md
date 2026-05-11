# Predefined commands example

This example shows how to expose a small allowlist of reviewed development commands to a trusted MCP client.

## Scenario

You want an MCP client to run known project commands, such as tests, without allowing arbitrary command execution.

## User goal

Run MCPFS locally with `commands.mode: "predefined"`.

## Files

- `README.md` — this guide.
- `mcpfs.cfg.json` — example config with predefined commands.

## Command flow

Build MCPFS and start the example config:

```bash
go build -o ./bin/mcpfs ./cmd/mcpfs
./bin/mcpfs -config examples/predefined-commands/mcpfs.cfg.json
```

Connect a trusted local MCP client over STDIO.

From the client:

1. Call `cmd_list`.
2. Confirm the `test` command is listed.
3. Call `cmd_run` with `id: "test"`.
4. Confirm an unknown command id is rejected.

## Expected output

- `cmd_list` returns the configured command id.
- `cmd_run` runs the configured argv array.
- Undefined command ids are rejected.
- `cmd_exec` is not registered because this example does not use unguarded mode.

## Security notes

Predefined commands are safer than unguarded commands because the operator can review each command before exposing it. Avoid commands that publish, deploy, delete broad directories, print secrets, or modify global machine state.

## Related docs

- [Commands](../../docs/commands.md)
- [Add predefined commands](../../docs/how-to/add-predefined-commands.md)
- [Command execution](../../docs/advanced/command-execution.md)
