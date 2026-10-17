# PRIM Backend

The Go API behind PRIM: REST endpoints for the storefront and admin back office, plus a worker that consumes Kafka events to send email and in-app notifications.

## Tech stack

- **Language & framework:** Go 1.26, Gin
- **Database:** PostgreSQL 18 (via `pgx`)
- **Cache & sessions:** Redis 8
- **Message broker:** Apache Kafka (KRaft)
- **Object storage:** MinIO (S3-compatible)
- **Auth:** JWT sessions stored in Redis, passwordless email OTP
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

- `internal/middleware` — auth, RBAC, CORS, locale, rate limiting, error handling, caching
- `pkg/config` — YAML + environment configuration
- `pkg/database` — connection pool and the `TxRunner` used for transactions
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

## Quick start

The easiest path is the full stack from the repository root:

```bash
make up        # or: make up-d
```

That starts the API on <http://localhost:8081> together with PostgreSQL, Redis, Kafka, MinIO, and the worker, and runs the migrations.

To run just the API against your own Postgres/Redis, from `backend/`:

```bash
go run ./cmd/api      # the API
go run ./cmd/worker   # the notification worker
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

Configuration comes from `configs/config.yaml`, overridden by environment variables (via `.env` and the compose file). The main variables:

| Variable | Default | Notes |
| --- | --- | --- |
| `HTTP_PORT` | `8081` | API port |
| `APP_ENV` | `development` | Enables dev conveniences; use `production` in prod |
| `DB_HOST` / `DB_PORT` / `DB_NAME` / `DB_USER` / `DB_PASSWORD` | `localhost` / `5432` / `prim` / `prim` / `prim` | PostgreSQL |
| `REDIS_HOST` / `REDIS_PORT` / `REDIS_PASSWORD` / `REDIS_DB` | `localhost` / `6379` / empty / `0` | Redis |
| `KAFKA_BROKERS` / `KAFKA_GROUP_ID` | `localhost:9092` / `prim-notification-workers` | Notification events |
| `MINIO_ENDPOINT` / `MINIO_PUBLIC_URL` / `MINIO_ACCESS_KEY` / `MINIO_SECRET_KEY` | `minio:9000` / `http://localhost:9000` / `admin` / `supersecret` | Object storage |
| `JWT_ACCESS_SECRET` / `JWT_REFRESH_SECRET` / `JWT_RESET_PASS_SECRET` | dev secrets | Replace in production |
| `JWT_ISSUER` / `JWT_AUDIENCE` | `prim` / `prim-api` | Token claims |
| `OTP_HASH_PEPPER` | dev value | Peppers OTP hashes |
| `CLIENT_BASE_URL` | `http://localhost:3000` | Used to build links in emails |
| `CORS_ALLOWED_ORIGINS` | — | Comma-separated allow list |
| `SMTP_HOST` / `SMTP_PORT` / `SMTP_USERNAME` / `SMTP_PASSWORD` / `SMTP_FROM_EMAIL` / `SMTP_FROM_NAME` | empty / `0` / … | Outbound email |
| `AUTH_CHALLENGE_TTL_MINUTES` / `AUTH_SESSION_TTL_DAYS` | `5` / `30` | Auth lifetimes |
| `RL_AUTH_START_REQ` / `RL_AUTH_RESEND_REQ` / `RL_AUTH_VERIFY_REQ` | `5` / `3` / `10` | Rate limits per window |

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
│   ├── middleware/           # auth, RBAC, CORS, locale, rate limit, errors
│   └── model/                # database models
├── pkg/                      # config, database, log, storage, mail, broker, api helpers
├── configs/                  # config.yaml and Air configs
├── migrations/               # schema and seed SQL
├── docs/                     # generated Swagger specs (public + admin)
├── seed/                     # seed media for the catalog
└── docker/                   # Dockerfiles and compose manifests
```
