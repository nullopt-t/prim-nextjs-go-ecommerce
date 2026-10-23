# Environment & Secrets Setup

## How the backend loads config

The backend is **env-only**. `pkg/config.Load()` reads environment variables via `cleanenv.ReadEnv`. There is **no implicit `configs/config.yaml` discovery** — that file has been removed from the repository.

An explicit YAML path is still honored for tooling/tests (a path argument or `CONFIG_PATH`), but normal development and production must not rely on it. `app.Run()` also calls `godotenv.Load()` so a local `.env` is picked up automatically; production injects variables through the platform.

> Build once, inject config at runtime. Never bake `.env` or secrets into the image.

All required variables are listed in `backend/.env.example`.

---

## The two traps in the current repo

### 1. `APP_ENV`, not `ENV`

`pkg/config` reads `APP_ENV` (`ServerConfig.Env`). The dev compose now sets `APP_ENV: development`; older revisions used `ENV: development`, which the config loader ignored. `pkg/log` falls back to `ENV`, which used to mask the problem in dev.

**Consequence if you use `ENV`:** the app runs as `development` even in production:
- dev-only OTP behavior remains enabled,
- the production OTP-pepper validation is skipped,
- logs may still be JSON (because `pkg/log` falls back to `ENV`), hiding the mistake.

**Fix:** set `APP_ENV` appropriately for the target environment.

### 2. Compose interpolation vs. container env

`env_file:` injects variables **into the container**. It does not automatically feed `${VAR}` interpolation in the compose file. Compose loads interpolation values from the shell and from `.env` in the **project directory** (the directory of the compose file, i.e. `docker/`), not from `env_file:`.

`docker/docker-compose.prod.yml` uses both:
- `env_file: ../.env` for the API/worker containers,
- `${DB_NAME}`, `${DB_PASSWORD}`, etc. for the Postgres/MinIO containers.

If you run `docker compose -f docker/docker-compose.prod.yml up` without `--env-file`, the API may get one password from `../.env` while Postgres falls back to `prim`. Always pass the same file:

```bash
docker compose --env-file .env -f docker/docker-compose.prod.yml up -d
```

(Best run from the repo root so `.env` is unambiguous.)

Also note the port mapping: `ports: "${HTTP_PORT:-8080}:8080"`. The app listens on `HTTP_PORT` from `env_file`. Keep `HTTP_PORT=8080` in production, or the host mapping and the app listener will disagree.

---

## Recommended deployment patterns

### A. Docker Compose on a single VM (matches this repo)

Keep a `.env` **on the server only**, next to the repo (it is already in `.gitignore`):

```bash
# On the server, as the deploy user
install -m 600 /dev/null /srv/prim/.env
$EDITOR /srv/prim/.env          # fill from .env.example
chown deploy:deploy /srv/prim/.env

cd /srv/prim/backend
docker compose --env-file /srv/prim/.env -f docker/docker-compose.prod.yml up -d
```

Harden the compose file:

```yaml
services:
  api:
    env_file: ${PRIM_ENV_FILE:-../.env}
    environment:
      APP_ENV: ${APP_ENV:-production}
  worker:
    env_file: ${PRIM_ENV_FILE:-../.env}
    environment:
      APP_ENV: ${APP_ENV:-production}
  postgres:
    environment:
      POSTGRES_PASSWORD: ${DB_PASSWORD:?DB_PASSWORD is required}
```

Then run with `--env-file /srv/prim/.env -e PRIM_ENV_FILE=/srv/prim/.env` or a symlink.

### B. systemd + Docker

```ini
[Service]
EnvironmentFile=/etc/prim/api.env
ExecStart=/usr/bin/docker compose --env-file /etc/prim/api.env -f /srv/prim/docker/docker-compose.prod.yml up
```

```bash
install -m 600 /dev/null /etc/prim/api.env
```

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

Current code validates only `OTP_HASH_PEPPER` when `APP_ENV=production`. Other secrets silently fall back to dev defaults (`prim`, `supersecret`). Add fail-fast checks in `pkg/config` when `IsProduction`:

- `DB_PASSWORD` non-empty and not `prim`;
- `MINIO_SECRET_KEY` non-empty and not `supersecret`;
- `CORS_ALLOWED_ORIGINS` non-empty and does not contain `*`;
- `DB_SSL_MODE != "disable"` (unless explicitly overridden);
- `CLIENT_BASE_URL` not `localhost`;
- SMTP credentials present if email is required;
- `APP_ENV` explicitly set (reject the implicit default in prod).

This turns a silent misconfiguration into a failed deploy.

---

## Quick checklist

- [ ] `.env` is gitignored and `chmod 600` on servers.
- [ ] `.env.example` is committed and contains no real secrets.
- [ ] `APP_ENV` is set appropriately.
- [ ] `HTTP_PORT` matches the compose port mapping.
- [ ] `--env-file` is passed to Compose so interpolation and containers agree.
- [ ] Secrets are generated fresh, not the dev defaults.
- [ ] `CORS_ALLOWED_ORIGINS` has no `*`.
- [ ] `DB_SSL_MODE=require`/`verify-full` for remote DBs.
- [ ] Managed DB/Redis use provider-issued credentials.
- [ ] Startup fails if a required prod secret is missing.
- [ ] Secrets are rotated after any exposure and on staff turnover.
