# Contributing

Thank you for considering a contribution to MCPlug.

MCPlug is security-sensitive software. Changes that affect upstream process spawning, tool aggregation and filtering, HTTP transport, authentication, secret handling, or supervisor restart behavior need careful review and tests where practical.

## Before you begin

Install Go and clone the repository:

```bash
git clone https://github.com/tedla-brandsema/mcplug.git
cd mcplug
```

Build and test before making changes:

```bash
go test ./...
go build ./cmd/plug
```

## Development checks

Run these checks before opening a pull request:

```bash
gofmt -w .
go test ./...
go build ./cmd/plug
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

- config validation or the server-name sanitizer;
- upstream spawning, env merging, or header injection;
- tool filtering (`includeTools`/`excludeTools`) or tool-name prefixing;
- supervisor lifecycle, restart, or shutdown behavior;
- the tool-error vs protocol-error mapping;
- secret redaction or logging;
- HTTP auth;
- OIDC/JWT validation;
- upstream timeouts.

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
