# PRIM

<p align="center">
  <a href="https://github.com/nullopt-t/prim-nextjs-go-ecommerce/actions/workflows/go.yml"><img src="https://img.shields.io/github/actions/workflow/status/nullopt-t/prim-nextjs-go-ecommerce/go.yml?label=Go%20Backend%20CI&style=flat-square" alt="Go Backend CI"></a>
  <a href="https://github.com/nullopt-t/prim-nextjs-go-ecommerce/actions/workflows/node.js.yml"><img src="https://img.shields.io/github/actions/workflow/status/nullopt-t/prim-nextjs-go-ecommerce/node.js.yml?label=Frontend%20CI&style=flat-square" alt="Frontend CI"></a>
  <a href="https://github.com/nullopt-t/prim-nextjs-go-ecommerce/blob/main/LICENSE"><img src="https://img.shields.io/github/license/nullopt-t/prim-nextjs-go-ecommerce?style=flat-square" alt="License"></a>
  <a href="https://github.com/nullopt-t/prim-nextjs-go-ecommerce/commits/main"><img src="https://img.shields.io/github/last-commit/nullopt-t/prim-nextjs-go-ecommerce/main?style=flat-square" alt="Last commit"></a>
</p>

PRIM is a bilingual (English / Arabic) e-commerce platform. A Next.js storefront, account area, admin back office, and vendor portal sit on top of a Go REST API, backed by PostgreSQL for data, Redis for cache and sessions, Kafka for notifications, and MinIO for media.

It's meant to be the whole shop, not just a demo: a catalog with variants and inventory, a cart and checkout with coupons and Cash on Delivery, customer accounts with orders and addresses, an admin back office, and a media library. Some parts (the vendor portal, saved payment cards) are intentionally still in progress.

## What's in the box

| Surface | Description |
| --- | --- |
| **Storefront** | Bilingual EN/AR with automatic RTL, catalog browsing and filters, product pages with variants and reviews, a slide-over cart, coupons, wishlist, and checkout with saved addresses. Shipping, tax, the COD fee, available payment methods, and the maintenance/announcement banners all come from the admin settings. |
| **Account area** (`/user`) | Order history and details, addresses, reviews, wishlist, plus profile, preferences, session management, and a danger zone. |
| **Admin back office** (`/admin`) | Dashboard, products and variants, categories, brands, tags, attributes, inventory with an adjustment ledger, orders and tracking, payments and refunds, promotions and redemptions, customers, media library, store settings, and system settings. |
| **Vendor portal** (`/vendor`) | Dashboard, products, inventory, orders, and payouts. Frontend prototype backed by mock data; the backend vendor domain hasn't landed yet. |
| **Backend** | Gin API organized into per-domain packages (handler → service → repository on a shared transaction runner), passwordless email-OTP auth with Redis sessions, Kafka-driven notifications handled by a separate worker, and MinIO uploads. Swagger docs are generated for both the public and admin surfaces. |

## Architecture

```mermaid
graph LR
    Client["Client / Browser"]

    subgraph Frontend["Frontend (Next.js 16 App Router)"]
        FE_Details["- React 19, TypeScript, Tailwind CSS 4<br/>- next-intl (Bilingual EN/AR + RTL)<br/>- Turbopack"]
    end

    subgraph Backend["Backend Service (Go 1.26 / Gin Engine)"]
        direction TB
        Handlers["Handlers / Controllers<br/>(HTTP, Sessions, Validation)"]
        Services["Services / Use Cases<br/>(Business Logic)"]
        Repos["Repositories / DAL<br/>(Data Abstraction & Drivers)"]

        Handlers --> Services
        Services --> Repos
    end

    subgraph Storage["Infrastructure & Storage"]
        PG[("PostgreSQL<br/>:5432 (Primary DB)")]
        Redis["Redis<br/>:6379 (Cache / Sessions / Rate Limits)"]
        MinIO["MinIO (S3)<br/>:9000 (Media Storage)"]
        Kafka["Kafka<br/>:9092 (Notification Events)"]
    end

    subgraph Worker["Notification Worker (Go 1.26)"]
        W["Kafka Consumer<br/>Email & in-app notifications"]
    end

    Client -->|":3000 (HTTP / SSR)"| Frontend
    Frontend -->|":8081 (REST API + Session Cookie)"| Handlers

    Repos -->|"SQL Queries (:5432)"| PG
    Repos -->|"Cache / Sessions (:6379)"| Redis
    Repos -->|"S3 Protocol (:9000)"| MinIO
    Services -->|"Publish Events"| Kafka
    Kafka -->|"Consume"| W
```

## Tech stack

| Layer | Stack |
| --- | --- |
| Frontend | Next.js 16 (App Router), React 19, TypeScript 7, Tailwind CSS 4, next-intl, Sonner, Recharts |
| Backend | Go 1.26, Gin, pgx, Swag |
| Database | PostgreSQL 18 |
| Cache & sessions | Redis 8 |
| Messaging | Apache Kafka (KRaft) |
| Object storage | MinIO (S3-compatible) |
| Tooling | Docker Compose, Air (hot reload), golang-migrate, GitHub Actions, CodeQL |

## Getting started

You need Docker and Docker Compose. From the repository root:

```bash
make up       # start everything and follow the logs
make up-d     # or start it in the background
```

This brings up the storefront, API, worker, PostgreSQL, Redis, Kafka, and MinIO, seeds the media bucket, and runs the migrations. The first build takes a few minutes while Go modules and npm packages download.

Once it's up, open the storefront at [http://localhost:3000](http://localhost:3000). The API is at [http://localhost:8081](http://localhost:8081), and a quick health check lives at [http://localhost:8081/health](http://localhost:8081/health).

> **URL configuration.** `backend/.env` holds the browser-facing `NEXT_PUBLIC_API_URL` plus `MINIO_PUBLIC_URL` and `CLIENT_BASE_URL`. Server-side rendering uses `API_URL` to reach the API over the Compose network (`http://api:8081` in dev, `http://api:8080` in prod), so the public URL is never used from inside a container. Change the public values in `backend/.env` and rebuild the frontend after changing them.

### Backend documentation

The backend is documented in [`backend/README.md`](./backend/README.md), with deeper operational and architectural notes under [`backend/docs/`](./backend/docs/):

- [`backend/docs/deployment/environment.md`](./backend/docs/deployment/environment.md) — environment loading, production traps, secrets checklist
- [`backend/docs/adr-vendor-domain.md`](./backend/docs/adr-vendor-domain.md) — multi-vendor domain design decision
- [`backend/docs/transaction-runner-refactor.md`](./backend/docs/transaction-runner-refactor.md) — context-based executor & transaction architecture
- [`backend/docs/connection-pool-starvation.md`](./backend/docs/connection-pool-starvation.md) — post-mortem and fix for connection-pool starvation

> **Two stacks.** `make up*` starts the dev stack (`docker-compose.dev.yml`: Next dev server + Air hot reload). `make prod-*` starts the production stack (`backend/docker/docker-compose.prod.yml`: built standalone images). They use separate containers and volumes.

### Services and ports

| Service | Address | Notes |
| --- | --- | --- |
| Storefront | <http://localhost:3000> | Next.js app |
| API | <http://localhost:8081> | Go REST API |
| Health check | <http://localhost:8081/health> | Liveness |
| Public API docs | <http://localhost:8081/api/v1/docs/public/index.html> | Swagger UI |
| Admin API docs | <http://localhost:8081/api/v1/docs/admin/index.html> | Requires an admin session |
| MinIO API | <http://localhost:9000> | S3 endpoint |
| MinIO console | <http://localhost:9001> | `admin` / `supersecret` |
| PostgreSQL | `localhost:5432` | db `prim`, user `prim`, password `prim` |
| Redis | `localhost:6379` | |
| Kafka | `localhost:9092` | |

## Common commands

Everything day-to-day is wrapped by the root `Makefile`:

```bash
make up            # start the full stack in the foreground
make up-d          # start it detached
make up-build      # rebuild images, then start
make down          # stop the containers
make clean         # stop and remove volumes (fresh database)

make logs          # tail everything
make logs-api      # only the Go API
make logs-worker   # only the notification worker
make logs-frontend # only the Next.js app

make migrate-up    # apply pending migrations
make migrate-down  # roll back the last migration
make migrate-reset # drop and re-apply everything

make prod-up       # build and start the production stack
make prod-down     # stop the production stack
make prod-logs     # tail the production stack
make prod-ps       # production stack status
```

More backend-specific targets (Swagger generation, binary builds, linting) live in `backend/Makefile`.

## Repository layout

```
prim/
├── backend/                     # Go API + notification worker
│   ├── cmd/                     # entrypoints: api, worker
│   ├── internal/                # domain packages (auth, catalog, cart, checkout, order, user, setting, ...)
│   ├── migrations/              # SQL schema + seed data
│   ├── pkg/                     # shared config, database, storage, logging, mail, broker
│   ├── configs/                 # Air hot-reload configs
│   ├── docs/                    # generated Swagger specs (public + admin) + architecture notes
│   ├── seed/                    # seed media for the catalog
│   └── README.md                # backend-specific documentation
├── frontend/                    # Next.js 16 storefront, account, admin, and vendor apps
│   ├── messages/                # en.json / ar.json
│   ├── public/                  # static assets
│   └── src/                     # app routes, features, context, services, hooks
├── docker-compose.dev.yml       # the local development stack
├── Makefile                     # development shortcuts
└── README.md
```

## Where things stand

- The **vendor portal** is frontend-only for now: its pages read from mock data until the vendor backend and API land.
- **Saved payment methods** are a UI shell; cards aren't persisted yet, and checkout currently runs Cash on Delivery (plus any card option enabled in settings).
- Admin settings are wired through to real behavior — checkout pricing, the maintenance gate, payment availability, and the storefront's identity and announcement banner all read from the backend settings store.

## Contributing

- Work from `main`; create feature branches and open pull requests.
- Backend changes must pass `go build ./...`, `go test ./...`, `go vet ./...`, and `go test -race ./...`.
- Regenerate Swagger specs (`make swagger` from `backend/`) whenever route or doc annotations change.
- Add migrations under `backend/migrations/` with both `.up.sql` and `.down.sql`; never edit an applied migration.
- Keep `.env` files out of version control and never commit secrets.

## Security

Never commit `.env` files, credentials, OTP peppers, or SMTP secrets. If you discover a security issue, please disclose it responsibly by opening a private advisory or contacting the maintainers rather than filing a public issue.

## License

MIT
