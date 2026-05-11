# Contributing

Thank you for considering a contribution to MCPFS.

MCPFS is security-sensitive software. Changes that affect filesystem access, writes, HTTP transport, authentication, command execution, path resolution, symlink handling, or Git operations need careful review and tests where practical.

## Before you begin

Install Go and clone the repository:

```bash
git clone https://github.com/tedla-brandsema/mcpfs.git
cd mcpfs
```

Build and test before making changes:

```bash
go test ./...
go build ./cmd/mcpfs
```

## Development checks

Run these checks before opening a pull request:

```bash
gofmt -w .
go test ./...
go build ./cmd/mcpfs
```

CI checks formatting, tests, and build.

## Documentation style

Documentation should be Markdown-first, direct, and practical.

Use:

- short paragraphs;
- active voice;
- second person for user instructions;
- copy-paste commands;
- expected output where useful;
- relative links;
- Mermaid diagrams when they clarify security or architecture.

Avoid:

- hype;
- broad safety claims;
- unexplained acronyms;
- copying old example prose without rewriting it;
- putting long reference material in the README.

## Security-sensitive changes

Add or update tests when changing:

- root resolution;
- include or exclude matching;
- `.gitignore` handling;
- symlink handling;
- writable root behavior;
- command execution;
- HTTP auth;
- OIDC/JWT validation;
- output or file size limits.

Update `docs/security.md` when behavior changes the trust boundary or safe-use guidance.

## Examples

Canonical examples live in `examples/`.

Each example should include:

- what it demonstrates;
- prerequisites;
- commands to run;
- expected output;
- cleanup steps when needed;
- security notes;
- links to related docs.

Do not commit real tokens, credentials, private paths, or secrets.

## Pull request checklist

Before opening a pull request, check that:

- the change has a clear user or maintainer benefit;
- tests pass with `go test ./...`;
- Go files are formatted with `gofmt`;
- documentation is updated when behavior changes;
- examples are updated when recommended usage changes;
- security-sensitive behavior is documented;
- no secrets or local-only private paths are committed.

## Reporting security issues

Do not open public issues for vulnerabilities that could put users at risk.

Use GitHub Security Advisories if available, or contact the maintainer directly. See [SECURITY.md](SECURITY.md) for reporting guidance.
