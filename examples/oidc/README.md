# OIDC example

Runs the gateway over HTTP with JWT validation against an external identity provider: issuer, audience, expiry, and a subject/email allowlist, with keys fetched from the provider's JWKS endpoint.

## Run

1. Edit `mcpfs.cfg.json`:
   - replace `/absolute/path/to/project` in both `mcpServers` entries;
   - set `auth.issuer`, `auth.audience`, and `auth.jwks_url` for your provider;
   - set `auth.allowed_subjects` (or use `allowed_emails`) to the identities allowed in.
2. Smoke-test and start:

   ```bash
   mcpfs ls -config mcpfs.cfg.json
   mcpfs -config mcpfs.cfg.json
   ```

## Connect

Clients send `Authorization: Bearer <JWT>` where the JWT is issued by the configured provider for the configured audience. Invalid or unlisted identities receive HTTP 401; JWKS fetch failures surface as HTTP 500.
