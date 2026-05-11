# Client recipes

Use these recipes to connect MCPFS to local and HTTP MCP clients.

## Local STDIO client

Use STDIO when your client can start MCPFS as a local process.

Command:

```text
/path/to/mcpfs/bin/mcpfs
```

Arguments are optional. To use an explicit config, pass `-config` followed by the config path.

Recommended setup:

- use a read-only root first;
- keep command execution disabled;
- verify with `fs_roots` before reading files.

## MCP Inspector

Start MCP Inspector:

```bash
bunx @modelcontextprotocol/inspector
```

Use these settings for a built binary:

```text
Transport Type: STDIO
Command: /path/to/mcpfs/bin/mcpfs
```

Use these settings for an explicit config:

```text
Transport Type: STDIO
Command: /path/to/mcpfs/bin/mcpfs
Arguments: -config /path/to/mcpfs.cfg.json
```

## HTTP client with bearer auth

Use bearer auth when a client connects over HTTP with a shared token.

Server config excerpt:

```json
"server": {
  "transport": "http",
  "addr": "127.0.0.1:8080",
  "path": "/mcp",
  "auth": {
    "mode": "bearer",
    "token_env": "MCPFS_TOKEN"
  }
}
```

Start MCPFS:

```bash
export MCPFS_TOKEN="$(openssl rand -hex 32)"
mcpfs -config /path/to/mcpfs.cfg.json
```

Client endpoint:

```text
http://127.0.0.1:8080/mcp
```

Configure the client to send the bearer token from your environment or secret store. Do not paste real tokens into committed config files.

## HTTP client with OIDC/JWT

Use OIDC when your client can obtain a JWT from an identity provider.

Server config excerpt:

```json
"auth": {
  "mode": "oidc",
  "issuer": "https://issuer.example.com",
  "audience": "mcpfs",
  "jwks_url": "https://issuer.example.com/.well-known/jwks.json",
  "allowed_subjects": ["user-or-client-subject-id"]
}
```

The client must send a bearer JWT with the expected issuer, audience, signature, expiry, and allowed identity.

## Troubleshooting

### The client cannot start MCPFS

Check that:

- the command path is absolute or valid from the client process;
- the binary exists and is executable;
- the config path exists if you passed `-config`;
- `go build -o ./bin/mcpfs ./cmd/mcpfs` succeeds.

### The client sees no roots

Check that:

- you added a root with `mcpfs project add`; or
- your explicit config contains at least one root; and
- the client is using the config you expect.

### HTTP requests fail with auth errors

Check that:

- the client sends the expected auth header;
- bearer token value matches the configured environment variable;
- OIDC issuer, audience, JWKS URL, and identity allowlists match the JWT;
- the token is not expired.

### File access is rejected

Check that:

- paths are relative to the configured root;
- the path is inside the root;
- include/exclude rules allow the path;
- `.gitignore` does not ignore the path;
- file size is below `max_file_bytes`.

## Related docs

- [Quick start](quick-start.md)
- [Transports](transports.md)
- [Security](security.md)
- [MCP tools reference](reference/tools.md)
