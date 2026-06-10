# ngrok development example

Exposes the gateway through an embedded ngrok tunnel so a remote MCP client (for example a ChatGPT connector) can reach it. Bearer auth is required — the public URL is reachable by anyone who finds it.

## Run

1. Edit `mcpfs.cfg.json`: replace `/absolute/path/to/project` in both entries. Optionally set `server.ngrok_url` to a reserved ngrok domain.
2. Export credentials:

   ```bash
   export NGROK_AUTHTOKEN=...               # from your ngrok dashboard
   export MCPFS_TOKEN=$(openssl rand -hex 32)
   ```

3. Smoke-test and start:

   ```bash
   mcpfs ls -config mcpfs.cfg.json
   mcpfs -config mcpfs.cfg.json
   ```

MCPFS logs the public MCP URL (`https://<subdomain>.ngrok.../mcp`) at startup.

## Connect a remote client

Add the printed URL as an MCP connector and configure it to send `Authorization: Bearer $MCPFS_TOKEN`. Treat the URL itself as sensitive and stop the process when you are done — the tunnel lives only as long as MCPFS runs.
