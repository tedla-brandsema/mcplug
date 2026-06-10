# HTTP bearer example

Runs the gateway over localhost HTTP with a shared bearer token, aggregating the reference filesystem server (read-style tools only) and the reference git server.

## Run

1. Edit `mcplug.cfg.json`: replace `/absolute/path/to/project` in both entries.
2. Export a strong token (never store it in the config):

   ```bash
   export MCPLUG_TOKEN=$(openssl rand -hex 32)
   ```

3. Smoke-test and start:

   ```bash
   plug ls -config mcplug.cfg.json
   plug -config mcplug.cfg.json
   ```

## Connect

Streamable-HTTP MCP clients connect to `http://127.0.0.1:8080/mcp` with header `Authorization: Bearer $MCPLUG_TOKEN`. Requests without the token receive HTTP 401. `GET /healthz` reports liveness without auth.

See `env.example` for the expected environment variable.
