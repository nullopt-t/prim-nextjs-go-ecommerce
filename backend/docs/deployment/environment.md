# Environment & Secrets Setup

## How the backend loads config

The backend is **env-only**. `pkg/config.Load()` reads environment variables via `cleanenv.ReadEnv`. There is **no implicit `configs/config.yaml` discovery** — that file has been removed from the repository.

An explicit YAML path is still honored for tooling/tests (a path argument or `CONFIG_PATH`), but normal development and production must not rely on it. `app.Run()` also calls `godotenv.Load()` so a local `.env` is picked up automatically; production injects variables through the platform.

> Build once, inject config at runtime. Never bake `.env` or secrets into the image.

Production variables are listed in `backend/.env.example`; local Docker Compose values are in `backend/dev.env.example`.

---

## The two traps in the current repo

### 1. `APP_ENV`, not `ENV`

`pkg/config` reads `APP_ENV` (`ServerConfig.Env`). The runtime `.env` sets `APP_ENV=development`; older revisions used `ENV=development`, which the config loader ignored. `pkg/log` falls back to `ENV`, which used to mask the problem in dev.

**Consequence if you use `ENV`:** the app runs as `development` even in production:
- dev-only OTP behavior remains enabled,
- the production OTP-pepper validation is skipped,
- logs may still be JSON (because `pkg/log` falls back to `ENV`), hiding the mistake.

**Fix:** set `APP_ENV` appropriately for the target environment.

### 2. Compose files read values from `backend/.env`

The Compose files only reference `env_file` for container runtime configuration. All runtime values live in `backend/.env`.

The production Compose file also uses `${NEXT_PUBLIC_API_URL}` for the frontend **build argument**. Docker Compose interpolation is only involved in that build step; pass `--env-file backend/.env` (the `make prod-up` target does this) so the value is available at build time.

The Compose files hardcode service ports. `HTTP_PORT` in `backend/.env` must match the app's host/container port:

- dev: `HTTP_PORT=8081`
- prod: `HTTP_PORT=8080`

If you change the port, change it in both `.env` and the Compose port mapping.

---

## Notification delivery fallback

Email notifications are published to Kafka first. If Kafka is unavailable or rejects a write within the publish timeout, the API writes the same `NotificationEvent` to the Redis stream `notifications.email.fallback`. The worker consumes that stream and delivers through SMTP, so OTP, welcome, and password-reset emails continue during a Kafka outage.

Operational notes:

- The worker always runs the Redis fallback consumer alongside the Kafka consumers.
- Failed SMTP deliveries are retried and then written to `notifications.email.fallback.dlq`.
- The fallback uses a Redis consumer group; unacknowledged events can be claimed after a worker restart.
- In-app notifications remain Kafka-only. This fallback covers SMTP email delivery.
- Redis is already required for auth and rate limiting, so no additional service is needed.

---

## Frontend production image

`frontend/Dockerfile` builds a Next.js standalone image. `NEXT_PUBLIC_API_URL` is compiled into the browser bundle, so it is supplied as a build argument from `backend/.env`:

```env
NEXT_PUBLIC_API_URL=https://api.example.com
```

`API_URL` is read at runtime for server-side calls. When using Compose, set it in `backend/.env`:

```env
API_URL=http://api:8080
```

Build and start the production stack from the repository root:

```bash
make prod-up
```

That runs `docker compose --env-file backend/.env -f backend/docker/docker-compose.prod.yml up -d --build`. The production Compose file runs pending migrations before starting the API and worker.

---

## Recommended deployment patterns

### A. Docker Compose on a single VM (matches this repo)

Keep `backend/.env` **on the server only** (it is already in `.gitignore`). The prod Compose file reads `../.env` relative to `backend/docker/`, i.e. `backend/.env`.

The production Compose file includes bundled PostgreSQL and Redis for a single-VM deployment, plus MinIO for object storage. Point `DATABASE_URL` and `REDIS_URL` at those services (or at managed services reachable from the Compose network) before starting it:

```env
DATABASE_URL=postgres://prim:prim@postgres:5432/prim?sslmode=disable
REDIS_URL=redis://redis:6379/0
```

```bash
# On the server, as the deploy user
install -m 600 /dev/null /srv/prim/backend/.env
$EDITOR /srv/prim/backend/.env   # fill from .env.example
chown deploy:deploy /srv/prim/backend/.env

cd /srv/prim
make prod-up
```

`make prod-up` builds the images, runs migrations, and starts the stack in the background. Use `make prod-ps`, `make prod-logs`, and `make prod-down` to manage it.

If your secrets live elsewhere, symlink them:

```bash
ln -s /etc/prim/api.env /srv/prim/backend/.env
```

### B. systemd + Docker

```ini
[Service]
WorkingDirectory=/srv/prim
ExecStart=/usr/bin/docker compose --env-file backend/.env -f backend/docker/docker-compose.prod.yml up
```

Put the values in `/srv/prim/backend/.env` (or symlink it to `/etc/prim/api.env`) and restrict permissions with `chmod 600`.

### C. Kubernetes

Split non-secret and secret values:

```yaml
apiVersion: v1
kind: ConfigMap
metadata: { name: prim-config }
data:
  APP_ENV: "production"
  HTTP_PORT: "8080"
  DB_HOST: "prim-postgres"
  DB_SSL_MODE: "require"
  CORS_ALLOWED_ORIGINS: "https://app.example.com"
---
apiVersion: v1
kind: Secret
metadata: { name: prim-secrets }
type: Opaque
stringData:
  DB_PASSWORD: "..."
  REDIS_PASSWORD: "..."
  MINIO_SECRET_KEY: "..."
  OTP_HASH_PEPPER: "..."
  SMTP_PASSWORD: "..."
```

Reference them from the Deployment:

```yaml
envFrom:
  - configMapRef: { name: prim-config }
  - secretRef:    { name: prim-secrets }
```

Use Sealed Secrets / External Secrets / SOPS rather than committing the Secret manifest.

### D. Managed secret stores

For AWS/GCP/Azure or a PaaS, store secrets in:

- AWS Secrets Manager / SSM Parameter Store,
- GCP Secret Manager,
- Azure Key Vault,
- HashiCorp Vault, Doppler, Infisical, or 1Password Secrets Automation.

Inject them at container start (entrypoint fetching secrets, CSI driver, External Secrets Operator, or the platform's native env/secret mechanism). Never `COPY .env` in a Dockerfile.

---

## What must be set in production

Copy `.env.example` and fill it. Minimum production set:

| Variable | Required | Notes |
|---|---|---|
| `APP_ENV` | yes | `production` (not `ENV`) |
| `HTTP_PORT` | yes | keep `8080` with the prod compose mapping |
| `CORS_ALLOWED_ORIGINS` | yes | explicit HTTPS origins, comma-separated |
| `DB_HOST` / `DB_PORT` / `DB_NAME` / `DB_USER` | yes | |
| `DB_PASSWORD` | yes, secret | never `prim` |
| `DB_SSL_MODE` | yes | `require` or `verify-full` for remote/managed |
| `REDIS_HOST` / `REDIS_PORT` / `REDIS_DB` | yes | |
| `REDIS_PASSWORD` | yes, secret | if Redis requires AUTH |
| `MINIO_ENDPOINT` / `MINIO_PUBLIC_URL` / `MINIO_ACCESS_KEY` | yes | |
| `MINIO_SECRET_KEY` | yes, secret | never `supersecret` |
| `OTP_HASH_PEPPER` | yes, secret | 32+ random bytes; not the dev default |
| `CLIENT_BASE_URL` | yes | public HTTPS origin |
| `SMTP_*` | yes if email is enabled | |
| `KAFKA_BROKERS` | yes if worker/events are used | |
| `AUTH_*`, `RL_*` | recommended | explicit values for prod |

Generate secrets with:

```bash
openssl rand -base64 48   # OTP_HASH_PEPPER
openssl rand -base64 32   # DB / Redis / MinIO passwords
```

---

## Startup validation

When `APP_ENV=production`, `pkg/config` fails fast if:

- `OTP_HASH_PEPPER` is empty or still the development default;
- `DB_PASSWORD` is empty or still `prim`;
- `MINIO_SECRET_KEY` is empty or still `supersecret`;
- `CORS_ALLOWED_ORIGINS` contains `*` or a localhost origin;
- `CLIENT_BASE_URL` is empty or points at localhost.

Not validated automatically: SMTP credentials, `DB_SSL_MODE`, and `REDIS_PASSWORD`. Set those explicitly for your environment.

This turns the most dangerous silent misconfigurations into a failed deploy.

---

## Quick checklist

- [ ] `.env` is gitignored and `chmod 600` on servers.
- [ ] `.env.example` is committed and contains no real secrets.
- [ ] `APP_ENV` is set appropriately.
- [ ] `HTTP_PORT` matches the hardcoded Compose port (`8080` in prod).
- [ ] `backend/.env` is the only runtime env source; Compose defines no values.
- [ ] Secrets are generated fresh, not the dev defaults.
- [ ] `CORS_ALLOWED_ORIGINS` has no `*`.
- [ ] `DB_SSL_MODE=require`/`verify-full` for remote DBs.
- [ ] Managed DB/Redis use provider-issued credentials.
- [ ] Startup fails if a required prod secret is missing.
- [ ] Secrets are rotated after any exposure and on staff turnover.
