# Command execution

Command execution is one of the highest-risk MCPFS features. Enable it only for trusted clients and reviewed roots.

## Modes

`disabled` registers no command tools. This is the safest default.

`predefined` registers `cmd_list` and `cmd_run`. Clients can run only configured command IDs.

`unguarded` registers `cmd_exec`, which allows arbitrary argv execution from a root-scoped working directory.

## Why predefined commands are preferred

Predefined commands make review possible. The operator can inspect:

- command id;
- root id;
- working directory;
- argv array;
- timeout;
- output limit;
- expected side effects.

## Workdir scoping

Commands run from root-scoped working directories. MCPFS resolves the working directory inside the configured root before execution.

## Output and timeout controls

Use timeouts and output limits to reduce runaway command risk and oversized responses. These controls do not make unsafe commands safe, but they help bound behavior.

## Shell invocation

MCPFS executes argv arrays directly. It does not perform shell interpolation unless the argv explicitly invokes a shell.

Avoid shell commands that combine unreviewed input, command substitution, glob expansion, or destructive operations.

## Remote clients

Do not expose unguarded command execution to remote clients unless the environment is isolated and the operator accepts terminal-level risk.

For remote use, prefer:

- no command execution; or
- predefined commands only; and
- HTTP auth; and
- narrow roots.
