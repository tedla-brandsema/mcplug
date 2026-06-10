# MCPFS examples

Each example aggregates the reference filesystem server (restricted to read-style tools via `includeTools`) and the reference git server (marked `optional`), and differs only in how the endpoint is exposed and authenticated.

| Example | Use when | Risk level |
| --- | --- | --- |
| [HTTP bearer](http-bearer/) | A local/trusted-network client connects over HTTP with a shared token. | Medium |
| [OIDC](oidc/) | You need identity-provider-backed JWT validation on the HTTP endpoint. | Medium to high |
| [ngrok development](ngrok-dev/) | A remote client (e.g. a ChatGPT connector) needs a short-lived public URL. | High |

For plain local stdio use no example is needed — see the [quick start](../docs/quick-start.md).

## Safety notes

- Replace `/absolute/path/to/project` before running; smoke-test with `mcpfs ls -config <file>`.
- Keep `includeTools` narrow: every connected client gets every exposed tool.
- Never expose the HTTP endpoint publicly without `bearer` or `oidc` auth.
- Read [docs/security.md](../docs/security.md) before using `http` or `http_ngrok`.
