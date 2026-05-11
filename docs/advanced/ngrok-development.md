# ngrok development

Use ngrok only when you need a short-lived public URL for development testing.

A tunnel increases exposure because clients outside your local machine can reach MCPFS while the tunnel is active.

## Safer development posture

Prefer:

- read-only roots;
- bearer or OIDC auth;
- command execution disabled;
- short-lived tokens;
- short-lived tunnels;
- stopping the tunnel immediately after testing.

Avoid:

- `auth.mode: "none"` on public tunnels;
- `read_write` roots on public tunnels;
- `commands.mode: "unguarded"` on public tunnels;
- exposing directories that contain secrets.

## Basic flow

1. Configure MCPFS with `transport: "http_ngrok"`.
2. Use a narrow read-only root.
3. Enable bearer or OIDC auth.
4. Start MCPFS with the ngrok auth token available to the process.
5. Copy the public MCP URL from the logs.
6. Configure the remote development client.
7. Stop MCPFS and the tunnel when testing is complete.

## Related example

See [ngrok development example](../../examples/ngrok-dev/).
