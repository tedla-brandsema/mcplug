# Error reference

This page summarizes common MCPFS error categories. Exact error messages may vary by tool and version.

## Root errors

Root errors occur when a request references an unknown root id or when the configured root cannot be used.

Check that:

- the root id exists in the active config;
- the client is using the expected config;
- the root path exists on the server host.

## Path errors

Path errors occur when a requested path is not allowed.

Common causes:

- absolute paths;
- `..` root escapes;
- symlink escapes;
- paths outside the configured root;
- paths excluded by config;
- paths ignored by `.gitignore`;
- files larger than `max_file_bytes`.

## Permission errors

Permission errors occur when a tool requires a capability that is not enabled.

Common causes:

- calling `fs_write` on a `read` root;
- calling command tools when `commands.mode` is `disabled`;
- calling `cmd_exec` when `commands.mode` is not `unguarded`.

## Git errors

Git errors occur when Git metadata cannot be read.

Common causes:

- root is not a Git repository;
- requested revision does not exist;
- requested file is not tracked;
- Git command output exceeds configured limits.

## Auth errors

Auth errors occur on HTTP transports when authentication fails.

Common causes:

- missing bearer token;
- bearer token mismatch;
- missing JWT;
- invalid JWT signature;
- wrong issuer;
- wrong audience;
- expired JWT;
- identity not present in `allowed_emails` or `allowed_subjects`.

## Command errors

Command errors occur when command execution cannot start or returns a non-zero exit code.

Common causes:

- unknown predefined command id;
- invalid root id;
- workdir rejected by root boundary checks;
- executable not found;
- timeout exceeded;
- output truncated by `max_output_bytes`;
- command exits with a non-zero status.

Command results can include stdout, stderr, exit code, duration, timeout metadata, and truncation metadata.
