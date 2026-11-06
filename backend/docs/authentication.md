# Authentication

PRIM uses **passwordless email OTP auth** with server-side sessions stored in Redis. There are no passwords to reset or rotate.

## Endpoints

| Method | Endpoint | Session required | Description |
| --- | --- | --- | --- |
| `POST` | `/auth/challenge/start` | No | Generate and email an OTP to the given address |
| `POST` | `/auth/challenge/verify` | No | Verify an OTP and issue a session cookie |
| `GET` | `/auth/me` | Yes (`ps`) | Return the current authenticated user |
| `GET` | `/auth/sessions` | Yes (`ps`) | List the current user's active sessions |
| `POST` | `/auth/logout` | Yes (`ps`) | Revoke the current session and clear the cookie |

## Flow

1. **Start challenge** — The client posts an email to `/auth/challenge/start`. The API:
   - generates a short-lived OTP,
   - hashes it with a peppered HMAC-SHA-256 (`OTP_HASH_PEPPER`),
   - stores the hash in Redis keyed by email with a TTL (`AUTH_CHALLENGE_TTL`, default `5m`),
   - sends the code by email via SMTP, with a Redis-stream fallback if Kafka is unavailable.

2. **Verify challenge** — The client posts the email and code to `/auth/challenge/verify`. On success the challenge is consumed, a session cookie named `ps` is issued, and the session is recorded in Redis with TTL `AUTH_SESSION_TTL` (default `720h` = 30 days).

3. **Authenticated requests** — Subsequent requests include the `ps` cookie. Middleware validates it against Redis and loads the user into the request context (`GET /auth/me`).

4. **Session management** — `GET /auth/sessions` lists active sessions; `POST /auth/logout` revokes the current session and clears the cookie.

## Rate limits (per sliding window)

| Endpoint | Limit | Window | Env overrides |
| --- | --- | --- | --- |
| Challenge start | 5 | 1 minute | `RL_AUTH_START_REQUESTS` / `RL_AUTH_START_WINDOW` |
| Challenge resend | 3 | 1 minute | `RL_AUTH_RESEND_REQUESTS` / `RL_AUTH_RESEND_WINDOW` |
| Challenge verify | 10 | 1 minute | `RL_AUTH_VERIFY_REQUESTS` / `RL_AUTH_VERIFY_WINDOW` |

Windows are Go durations (e.g. `1m`, `1h`).

## Production requirements

- **`OTP_HASH_PEPPER`** must be a fresh 32+ random value (not the dev default). It prevents precomputation attacks over the small OTP space if Redis contents are exposed. Generate with `openssl rand -base64 48`.
- **`APP_ENV=production`** enables fails-fast startup validation that rejects the dev pepper, `prim` DB passwords, and `supersecret` MinIO keys.
- **`CLIENT_BASE_URL`** must point at the public HTTPS origin used in email links.
- Never log OTP values, challenge hashes, or full request bodies containing tokens.

## Session cookie

- Name: `ps`
- Stored server-side in Redis; the cookie holds only the session identifier.
- Invalidated on logout, on TTL expiry, and on explicit revocation via admin user management.
