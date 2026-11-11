# PRIM Backend

<p align="center">
  <a href="https://github.com/nullopt-t/prim-nextjs-go-ecommerce/actions/workflows/go.yml"><img src="https://img.shields.io/github/actions/workflow/status/nullopt-t/prim-nextjs-go-ecommerce/go.yml?label=Go%20Backend%20CI&style=flat-square" alt="Go Backend CI"></a>
  <a href="https://github.com/nullopt-t/prim-nextjs-go-ecommerce/blob/main/LICENSE"><img src="https://img.shields.io/github/license/nullopt-t/prim-nextjs-go-ecommerce?style=flat-square" alt="License"></a>
  <a href="https://github.com/nullopt-t/prim-nextjs-go-ecommerce/commits/main"><img src="https://img.shields.io/github/last-commit/nullopt-t/prim-nextjs-go-ecommerce/main?style=flat-square" alt="Last commit"></a>
  <a href="https://pkg.go.dev/github.com/nullopt-t/prim-backend"><img src="https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square" alt="Go reference"></a>
</p>

The Go API behind PRIM: REST endpoints for the storefront and admin back office, plus a worker that consumes Kafka events to send email and in-app notifications.

Built on Gin 1.12 with Go 1.26. The API exposes two documented surfaces — **public** (storefront) and **admin** — each served as Swagger UI. A separate notification worker consumes Kafka and sends email and in-app notifications.

## Tech stack

| Layer | Choice |
| --- | --- |
| Language & framework | Go 1.26, Gin |
| Database | PostgreSQL 18 (via `pgx`) |
| Cache & sessions | Redis 8 |
| Message broker | Apache Kafka (KRaft) |
| Object storage | MinIO (S3-compatible) |
| Auth | Passwordless email OTP with server-side sessions stored in Redis |
| API docs | Swagger / OpenAPI, generated with `swag` |
| Migrations | `golang-migrate` |
| Hot reload | Air |
| Testing | `testify` + `testcontainers-go` |

## Quick start

You need Go 1.26 and Docker (for integration tests and the dev stack).

**Option A — run the API and worker directly:**

```bash
cd backend
cp .env.example .env        # fill in OTP_HASH_PEPPER and any secrets
go mod download
go run ./cmd/api            # the API on :8081
go run ./cmd/worker          # Kafka notification worker + reservation reaper
```

**Option B — backend dev stack (API + Postgres + Redis + MinIO) via Compose:**

```bash
make up          # start the backend dev stack
make up-build    # rebuild images first
make down        # stop it
make clean       # stop and remove volumes (fresh database)
```

Verify the API is healthy:

```bash
curl http://localhost:8081/health   # -> {"status":"ok"}
```

Public Swagger UI: <http://localhost:8081/api/v1/docs/public/index.html>  
Admin Swagger UI: <http://localhost:8081/api/v1/docs/admin/index.html> (requires an authenticated admin session)

Regenerate the specs after changing doc annotations:

```bash
make swagger
```

## Architecture

Each domain is its own package and follows the same shape:

```
HTTP request
  → gin handler     (decode, validate, shape the response)
  → service         (business rules, transactions)
  → repository      (SQL)
  → PostgreSQL
```

Shared concerns live outside the domains:

- `internal/shared/middleware` — auth, RBAC, CORS, locale, rate limiting, error handling, caching, gzip response compression, security headers (nosniff, CSP, HSTS)
- `pkg/config` — environment configuration (see [Configuration](#configuration))
- `pkg/database` — connection pool and the `TxRunner` used for transactions (see [Transaction Runner Architecture](docs/transaction-runner-refactor.md) and [Connection Pool Starvation Prevention](docs/connection-pool-starvation.md))
- `pkg/broker/kafka` — producer used by the API
- `pkg/storage` — MinIO provider
- `internal/notifier` — Kafka consumer + email delivery used by the worker

### Domains

| Package | Responsibility |
| --- | --- |
| `auth` | Email OTP challenges, sessions, logout |
| `user` | Profiles, addresses, notifications, admin user management |
| `catalog/brand`, `catalog/category`, `catalog/tag`, `catalog/category/attribute` | Product taxonomy |
| `catalog/product`, `catalog/product/variant` | Products, variants, media, publishing workflow |
| `catalog/inventory` | Stock levels and the adjustment ledger |
| `catalog/review` | Product reviews and moderation |
| `catalog/coupon` | Promotions, coupons, redemptions |
| `cart` | Shopping cart |
| `checkout` | Order placement, coupon application, settings-driven pricing |
| `order` | Orders, tracking, payments, dashboard stats |
| `wishlist` | Customer wishlists |
| `object` | Media uploads and the media library |
| `setting` | Store settings, public settings, maintenance gate |

### Stock reservation lifecycle

Checkout never deducts stock by guessing: it creates a reservation row linked to
the order, and that hold is resolved explicitly.

- **Cash on delivery** is confirmed when placed, so its reservation is committed
  immediately in the checkout transaction (on-hand is deducted). It is never
  cancelled by the expiry reaper.
- **Online payment** holds stock for the configured reservation window. The hold
  is committed into a sale when an admin marks the payment `captured`; if the
  window lapses first, the reaper cancels the still-pending order and releases
  the hold back to salable stock.
- **Cancellation** resolves the correct compensation: release a still-held
  reservation (unpaid order), or write a `return` ledger entry when the sale was
  already committed (paid order). This avoids the old double-count where
  cancelling an unpaid order inflated stock.
- The **worker** runs the reservation reaper once a minute. Both commit and
  release are transactional and idempotent, so overlapping runs and multiple
  worker/API instances are safe.

Reservations are keyed by `order_id` (`inventory_reservations.order_id`, part of
`000001_init`). Stock math is: `on_hand` = sum of ledger entries,
`reserved` = live holds, `available` = `max(on_hand - reserved, 0)`.

### Media uploads

`POST /admin/media` accepts raster images only (JPEG, PNG, GIF, WebP, AVIF) up
to 10 MiB. The handler sniffs the real bytes rather than trusting the client's
`Content-Type`, which rejects SVG and other script-bearing payloads that could
cause stored XSS when served from the public bucket.

## Migrations

The schema and seed data live in `migrations/` and are applied with `golang-migrate`.
There are exactly two migration versions:

| Version | Purpose |
| --- | --- |
| `000001_init` | The full schema: types, tables, indexes, constraints. |
| `000002_seed` | Demo/development data for the storefront. |

```bash
make migrate-up      # apply pending migrations
make migrate-down    # roll back the last one
make migrate-reset   # drop everything and re-apply
make migrate-force   # force a version (when a migration is stuck)
```

Add new schema changes to `000001_init.up.sql` (with the matching rollback in
`000001_init.down.sql`) rather than creating another numbered migration.

## Endpoint overview

All routes are under `/api/v1`.

| Area | Examples |
| --- | --- |
| Auth | `POST /auth/challenge/start`, `POST /auth/challenge/verify`, `GET /auth/me`, `GET /auth/sessions`, `POST /auth/logout` |
| Catalog | `GET /products`, `GET /products/:slug`, `GET /categories`, `GET /brands`, `GET /tags`, `GET /variants/:id` |
| Cart | `GET /cart`, `POST /cart/items`, `PATCH /cart/items/:id`, `DELETE /cart/items/:id` |
| Checkout | `POST /checkout`, `POST /coupons/apply` |
| Orders | `GET /orders`, `GET /orders/:id`, `POST /orders/:id/cancel` |
| Account | `/user/addresses`, `/user/notifications`, `/reviews` |
| Wishlist | `GET /wishlist`, `POST /wishlist/items`, `DELETE /wishlist/items/:id` |
| History | `GET /user/history/products`, `POST /user/history/products/:id`, `POST /user/history/products/sync`, `DELETE /user/history/products` |
| Settings | `GET /settings/public` |
| Admin | `/admin/products`, `/admin/categories`, `/admin/inventory`, `/admin/orders`, `/admin/payments`, `/admin/promotions`, `/admin/users`, `/admin/media`, `/admin/settings`, `/admin/dashboard` |

`/health` and `/api/v1/health` return `{"status":"ok"}`.

## Configuration

Configuration is **environment-only**; there is no committed `configs/config.yaml` and the Compose files contain no env values. `pkg/config.Load()` reads environment variables via `cleanenv.ReadEnv`. Copy `.env.example` to `.env` for local development, or `.env.example` to `.env` for production (see [`docs/deployment/environment.md`](docs/deployment/environment.md)).

> Never commit `.env` files, credentials, OTP peppers, or SMTP secrets. `.env` is gitignored.

### Core variables

| Variable | Default | Notes |
| --- | --- | --- |
| `APP_ENV` | `development` | `development` or `production` (prod enables fails-fast secret validation) |
| `PORT` / `HTTP_PORT` | `8081` | API port; prod Compose overrides to `8080` |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:3000,...` | Comma-separated allow list; never `"*"` with credentials |
| `CLIENT_BASE_URL` | `http://localhost:3000` | Used to build links in emails |
| `DATABASE_URL` | `postgresql://prim:prim@postgres:5432/prim?sslmode=disable` | PostgreSQL connection URL |
| `REDIS_URL` | `redis://redis:6379/0` | Redis connection URL |
| `KAFKA_BROKERS` / `KAFKA_GROUP_ID` | `kafka:9092` / `prim-notification-workers` | Notification events |
| `MINIO_ENDPOINT` / `MINIO_PUBLIC_URL` | `minio:9000` / `http://localhost:9000` | Object storage |
| `MINIO_ACCESS_KEY` / `MINIO_SECRET_KEY` | `admin` / `supersecret` | Dev defaults only; change for production |
| `MINIO_BUCKET` | `product-media` | Media bucket |
| `OTP_HASH_PEPPER` | dev value | Peppers OTP hashes; **generate a fresh value for production** |
| `AUTH_CHALLENGE_TTL` / `AUTH_SESSION_TTL` | `5m` / `720h` | Go durations; `720h` = 30 days |

### Rate limits (requests per sliding window)

| Variable | Default | Window |
| --- | --- | --- |
| `RL_AUTH_START_REQUESTS` / `RL_AUTH_START_WINDOW` | `5` | `1m` |
| `RL_AUTH_RESEND_REQUESTS` / `RL_AUTH_RESEND_WINDOW` | `3` | `1m` |
| `RL_AUTH_VERIFY_REQUESTS` / `RL_AUTH_VERIFY_WINDOW` | `10` | `1m` |

### SMTP

| Variable | Default | Notes |
| --- | --- | --- |
| `SMTP_HOST` / `SMTP_PORT` | empty / `0` | Outbound email host |
| `SMTP_USERNAME` / `SMTP_PASSWORD` | — | Dev: leave empty for Mailpit |
| `SMTP_FROM_EMAIL` / `SMTP_FROM_NAME` | `no-reply@localhost` / `Prim` | Sender identity |

### Production minimum set

| Variable | Required | Notes |
| --- | --- | --- |
| `APP_ENV` | yes | `production` |
| `CORS_ALLOWED_ORIGINS` | yes | Explicit HTTPS origins |
| `DB_PASSWORD` | yes, secret | Never `prim` |
| `DB_SSL_MODE` | yes | `require` or `verify-full` for remote DBs |
| `MINIO_SECRET_KEY` | yes, secret | Never `supersecret` |
| `OTP_HASH_PEPPER` | yes, secret | 32+ random bytes |
| `CLIENT_BASE_URL` | yes | Public HTTPS origin |
| `SMTP_*` | yes if email | |
| `KAFKA_BROKERS` | yes if events | |

See [`docs/deployment/environment.md`](docs/deployment/environment.md) for the production traps (`APP_ENV` vs `ENV`), secrets management, and the startup validation that fails fast on dangerous misconfigurations.

## Authentication

PRIM uses **passwordless email OTP auth** with server-side sessions stored in Redis. There are no passwords to reset.

### Flow

1. **Start challenge** — `POST /auth/challenge/start` with an email. The API generates an OTP, hashes it (peppered HMAC-SHA-256), stores it in Redis with a TTL, and sends the code by email (SMTP, with a Redis-stream fallback if Kafka is unavailable).
2. **Verify challenge** — `POST /auth/challenge/verify` with the email and code. On success, a session cookie (`ps`) is issued and the challenge is consumed.
3. **Session** — `GET /auth/me` returns the current user from the session. Sessions are tracked in Redis and can be listed/revoked via `GET /auth/sessions`.
4. **Logout** — `POST /auth/logout` revokes the current session and clears the cookie.

### Rate limits

| Endpoint | Limit | Window |
| --- | --- | --- |
| Challenge start | 5 | 1 minute |
| Challenge resend | 3 | 1 minute |
| Challenge verify | 10 | 1 minute |

### Production requirements

- `OTP_HASH_PEPPER` must be a fresh 32+ random value (not the dev default) — it prevents precomputation attacks over the small OTP space if Redis is exposed.
- `APP_ENV=production` enables fails-fast validation that rejects the dev pepper, `prim` DB passwords, and `supersecret` MinIO keys at startup.
- `CLIENT_BASE_URL` must point at the public HTTPS origin used in email links.

See [`docs/authentication.md`](docs/authentication.md) for the full auth contract.

## Testing

```bash
go test ./...       # all packages
go vet ./...        # static checks
```

Integration tests spin up real PostgreSQL and Redis containers via `testcontainers-go`, so Docker must be running.

## Backend commands

From `backend/`:

```bash
make up          # start the backend dev stack
make up-build    # rebuild images first
make down        # stop it
make clean       # stop and remove volumes
make logs        # tail the stack
make logs-api    # only the API
make logs-worker # only the worker
make swagger     # regenerate Swagger specs
make build       # cross-compile the api and worker binaries into ./bin
```

## Project structure

```
backend/
├── cmd/
│   ├── api/                  # API entrypoint
│   └── worker/               # notification worker entrypoint
├── internal/
│   ├── app/                  # wiring: config, clients, routes, middleware
│   ├── auth/                 # OTP challenges and sessions
│   ├── cart/                 # shopping cart
│   ├── catalog/              # brands, categories, products, variants, inventory, reviews, tags, coupons
│   ├── checkout/             # order placement
│   ├── order/                # orders, payments, dashboard
│   ├── user/                 # profiles, addresses, notifications, admin users
│   ├── wishlist/             # wishlists
│   ├── object/               # media uploads
│   ├── setting/              # store settings + maintenance middleware
│   ├── notifier/             # Kafka consumer and email sender
│   ├── shared/               # middleware, error codes, jobs, OTP, crypto, email templates
│   └── model/                # database models
├── pkg/                      # config, database, log, storage, mail, broker, api helpers
├── configs/                  # Air configs
├── migrations/               # schema and seed SQL
├── docs/                     # Swagger specs, architecture & refactor documentation
├── seed/                     # seed media for the catalog
└── docker/                   # Dockerfiles and compose manifests
```

## Design decisions & operational notes

| Document | What it covers |
| --- | --- |
| [`docs/transaction-runner-refactor.md`](docs/transaction-runner-refactor.md) | Context-based executor propagation and transaction-aware execution architecture |
| [`docs/connection-pool-starvation.md`](docs/connection-pool-starvation.md) | Post-mortem and fix for connection-pool starvation that stalled the API |
| [`docs/adr-vendor-domain.md`](docs/adr-vendor-domain.md) | Multi-vendor domain design decision (vendor-owned products, payout state machine) |
| [`docs/deployment/environment.md`](docs/deployment/environment.md) | Environment loading, production traps, secrets checklist |

## Contributing

- Work from `main`; create feature branches and open pull requests.
- Changes must pass `go build ./...`, `go test ./...`, `go vet ./...`, and `go test -race ./...`.
- Follow the handler → service → repository shape; do not introduce a new pattern for one endpoint.
- Regenerate Swagger specs (`make swagger`) whenever route or doc annotations change.
- Add migrations under `migrations/` with both `.up.sql` and `.down.sql`; never edit an applied migration.
- Keep `.env` files out of version control and never commit secrets.

## Security

Never commit `.env` files, credentials, OTP peppers, or SMTP secrets. If you discover a security issue, please disclose it responsibly by opening a private advisory or contacting the maintainers rather than filing a public issue.

