# Release policy

MCPFS is currently **Beta**.

It is useful and tested, but pre-v1 behavior can still change. This page defines the current compatibility boundary and the criteria for a stable v1 release.

## Current maturity

MCPFS is Beta because it has working filesystem, Git, project overview, HTTP auth, write, and command execution features, but the public adoption surface is still being finalized.

Beta means:

- MCPFS can be used for real workflows by users who understand the trust boundary;
- documented security-sensitive behavior should be treated seriously;
- configuration and tool details may still change before v1;
- breaking changes should be documented in `CHANGELOG.md` and release notes.

## What is intended to be stable now

The following areas are intended to remain conceptually stable:

- explicit configured roots;
- read-only roots by default;
- per-root `read_write` opt-in;
- root-scoped filesystem access;
- `.gitignore`, include, and exclude filtering;
- read-only Git inspection;
- project overview support;
- command execution disabled by default;
- predefined command execution as the preferred command mode;
- HTTP auth modes for bearer and OIDC.

Exact field names, tool details, response shapes, and defaults may still change before v1.

## What may change before v1

Before v1, MCPFS may change:

- configuration schema details;
- transport settings;
- auth settings;
- command settings;
- MCP tool inputs and outputs;
- default limits;
- example layout;
- documentation organization;
- release packaging and installation instructions.

Breaking changes should include migration notes when practical.

## Changelog expectations

`CHANGELOG.md` should describe user-visible changes by release.

Call out:

- added features;
- changed behavior;
- deprecated behavior;
- removed behavior;
- security-sensitive changes;
- migration notes for config, auth, command, or tool behavior changes.

## Compatibility policy

Until v1, MCPFS uses a pre-v1 compatibility policy:

- breaking changes are allowed;
- breaking changes should be documented;
- security fixes may change behavior;
- examples and docs should track the current recommended API and config shape.

After v1, the compatibility policy should define which surfaces are stable, including CLI flags, configuration fields, tool names, and documented response behavior.

## v1 readiness criteria

MCPFS is ready for v1 when:

- README is concise and adopter-focused;
- quick start works from a fresh clone;
- security docs clearly define trust boundaries and unsafe modes;
- configuration, transport, command, and tool reference docs are complete;
- examples are repo-local, runnable, and documented;
- CI passes;
- tests pass;
- release notes and changelog are current;
- the maintainer is comfortable supporting documented behavior as stable.
