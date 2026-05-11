# Authentication model

MCPFS authentication applies to HTTP transports. STDIO transport relies on local process boundaries and client trust.

## Auth modes

| Mode | Description |
| --- | --- |
| `none` | No HTTP authentication. Use only for local-only or tightly controlled development. |
| `bearer` | Static shared token loaded from an environment variable. |
| `oidc` | JWT validation using issuer, audience, JWKS, and identity allowlists. |

## No auth

`auth.mode: "none"` is acceptable for local-only development when the HTTP listener is not exposed to untrusted clients.

Do not use no-auth HTTP on untrusted networks.

## Bearer auth

Bearer auth is simple and works well for controlled deployments where clients can share one secret.

Trade-offs:

- simple to configure;
- easy to rotate manually;
- no per-user identity by default;
- token exposure grants access until rotation.

## OIDC auth

OIDC auth validates JWTs from an external identity provider.

Trade-offs:

- stronger identity integration;
- supports issuer, audience, and identity allowlists;
- requires identity-provider setup;
- requires careful validation and testing.

## Choosing a mode

Use STDIO for local clients when possible.

Use bearer auth for simple HTTP deployments.

Use OIDC when you need identity-provider-backed access control.

Avoid no-auth HTTP except for local or short-lived controlled development.
