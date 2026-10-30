# PRIM Backend

The Go API behind PRIM: REST endpoints for the storefront and admin back office, plus a worker that consumes Kafka events to send email and in-app notifications.

## Tech stack

- **Language & framework:** Go 1.26, Gin
- **Database:** PostgreSQL 18 (via `pgx`)
- **Cache & sessions:** Redis 8
- **Message broker:** Apache Kafka (KRaft)
- **Object storage:** MinIO (S3-compatible)
- **Auth:** Passwordless email OTP with server-side sessions stored in Redis
- **API docs:** Swagger / OpenAPI, generated with `swag`
- **Migrations:** `golang-migrate`
- **Hot reload:** Air
- **Testing:** `testify` + `testcontainers-go`

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
- `pkg/config` — YAML + environment configuration
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

Reservations are keyed by `order_id` (`inventory_reservations.order_id`,
migration `000004`). Stock math is: `on_hand` = sum of ledger entries,
`reserved` = live holds, `available` = `max(on_hand - reserved, 0)`.

### Media uploads

`POST /admin/media` accepts raster images only (JPEG, PNG, GIF, WebP, AVIF) up
to 10 MiB. The handler sniffs the real bytes rather than trusting the client's
`Content-Type`, which rejects SVG and other script-bearing payloads that could
cause stored XSS when served from the public bucket.

## Quick start

The easiest path is the full stack from the repository root:

```bash
make up        # or: make up-d
```

That starts the API on <http://localhost:8081> together with PostgreSQL, Redis, Kafka, MinIO, and the worker, and runs the migrations.

To run just the API against your own Postgres/Redis, from `backend/`:

```bash
go run ./cmd/api      # the API
go run ./cmd/worker   # Kafka notification worker + reservation reaper
```

Or use the backend's own compose stack (API + Postgres + Redis + MinIO) from `backend/`:

```bash
make up
```

### API documentation

Swagger UI is generated for two surfaces and served by the API:

- **Public (storefront)** — <http://localhost:8081/api/v1/docs/public/index.html>
- **Admin** — <http://localhost:8081/api/v1/docs/admin/index.html> (requires an authenticated admin session)

Regenerate the specs after changing annotations:

```bash
make swagger
```

### Endpoint overview

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
| Settings | `GET /settings/public` |
| Admin | `/admin/products`, `/admin/categories`, `/admin/inventory`, `/admin/orders`, `/admin/payments`, `/admin/promotions`, `/admin/users`, `/admin/media`, `/admin/settings`, `/admin/dashboard` |

`/health` and `/api/v1/health` return `{"status":"ok"}`.

## Migrations

The schema and seed data live in `migrations/` and are applied with `golang-migrate`:

```bash
make migrate-up      # apply pending migrations
make migrate-down    # roll back the last one
make migrate-reset   # drop everything and re-apply
make migrate-force   # force a version (when a migration is stuck)
```

## Configuration

Configuration is **environment-only**; there is no committed `configs/config.yaml` and the Compose files contain no env values. Copy `dev.env.example` to `.env` for local Docker Compose, or `.env.example` to `.env` for production (see `docs/deployment/environment.md`). The main variables:

| Variable | Default | Notes |
| --- | --- | --- |
| `HTTP_PORT` | `8081` | API port |
| `APP_ENV` | `development` | Enables dev conveniences; use `production` in prod |
| `DB_HOST` / `DB_PORT` / `DB_NAME` / `DB_USER` / `DB_PASSWORD` | `localhost` / `5432` / `prim` / `prim` / `prim` | PostgreSQL |
| `REDIS_HOST` / `REDIS_PORT` / `REDIS_PASSWORD` / `REDIS_DB` | `localhost` / `6379` / empty / `0` | Redis |
| `KAFKA_BROKERS` / `KAFKA_GROUP_ID` | `localhost:9092` / `prim-notification-workers` | Notification events |
| `MINIO_ENDPOINT` / `MINIO_PUBLIC_URL` / `MINIO_ACCESS_KEY` / `MINIO_SECRET_KEY` | `minio:9000` / `http://localhost:9000` / `admin` / `supersecret` | Object storage |
| `OTP_HASH_PEPPER` | dev value | Peppers OTP hashes |
| `CLIENT_BASE_URL` | `http://localhost:3000` | Used to build links in emails |
| `CORS_ALLOWED_ORIGINS` | — | Comma-separated allow list |
| `SMTP_HOST` / `SMTP_PORT` / `SMTP_USERNAME` / `SMTP_PASSWORD` / `SMTP_FROM_EMAIL` / `SMTP_FROM_NAME` | empty / `0` / … | Outbound email |
| `AUTH_CHALLENGE_TTL` / `AUTH_SESSION_TTL` | `5m` / `720h` | Go durations; `720h` = 30 days |
| `RL_AUTH_START_REQUESTS` / `RL_AUTH_RESEND_REQUESTS` / `RL_AUTH_VERIFY_REQUESTS` | `5` / `3` / `10` | Requests per sliding window |
| `RL_AUTH_START_WINDOW` / `RL_AUTH_RESEND_WINDOW` / `RL_AUTH_VERIFY_WINDOW` | `1m` / `1m` / `1m` | Sliding window as a Go duration |

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
