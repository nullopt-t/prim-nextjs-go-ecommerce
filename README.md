# PRIM — Full-Stack E-Commerce Platform

A production-grade, bilingual (EN/AR) e-commerce monorepo built with **Next.js 15 (App Router)** and **Go (Gin & Clean Architecture)**, backed by **PostgreSQL**, **Redis**, and **MinIO**.

---

## Architecture Overview

```
                                      ┌─────────────────────────────────┐
                                      │       Client / Browser          │
                                      └────────────────┬────────────────┘
                                                       │
                                        Port 3000      │
                                                       ▼
                      ┌─────────────────────────────────────────────────────────────────┐
                      │              Frontend (Next.js 15 App Router)                   │
                      │   - React 19, TypeScript, Tailwind CSS, Lucide Icons            │
                      │   - next-intl (Bilingual EN / AR with automatic RTL)            │
                      │   - Turbopack Hot Reloading & Responsive Layouts                │
                      └────────────────────────────────┬────────────────┘
                                                       │  REST API Calls
                                        Port 8081      │  (JWT Authentication)
                                                       ▼
                      ┌─────────────────────────────────────────────────────────────────┐
                      │                 Backend Service (Go / Gin)                      │
                      │   - Clean Architecture (Handler -> Service -> Repository)      │
                      │   - Asynchronous Background Workers                             │
                      │   - Air Live Reloading, Swagger / OpenAPI Documentation         │
                      └────────────┬───────────────────┬───────────────────┬────────────┘
                                   │                   │                   │
                  SQL Queries      │      Cache / Queues│      Media Uploads │ S3 Protocol
                  Port 5432        │      Port 6379     │      Port 9000     │
                                   ▼                   ▼                   ▼
                      ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
                      │   PostgreSQL    │ │      Redis      │ │   MinIO (S3)    │
                      │   (Catalog, DB) │ │  (Cache/Sessions│ │ (Asset Storage) │
                      └─────────────────┘ └─────────────────┘ └─────────────────┘
```

---

## Tech Stack

### Frontend
- **Framework:** Next.js 15 with React 19 & TypeScript
- **Styling:** Tailwind CSS with dynamic theme variables & responsive design tokens
- **Internationalization:** `next-intl` with full LTR (English) & RTL (Arabic) support
- **State & Context:** React Context (`CartContext`, `CatalogContext`, `ThemeContext`)
- **Notifications:** Sonner toasts

### Backend
- **Language & Framework:** Go 1.23+, Gin Web Framework
- **Architecture:** Clean Architecture (Handlers, Services, Repositories)
- **Database & Cache:** PostgreSQL (pgx/sql), Redis
- **Storage:** MinIO (S3-compatible Object Storage)
- **Hot-Reload:** Air
- **Database Migrations:** `golang-migrate`
- **Documentation:** Swagger / OpenAPI via `swag`

---

## Repository Structure

```
prim-nextjs-go-ecommerce/
├── backend/                       # Go REST API & background workers
│   ├── cmd/                       # Application entrypoints (api, worker)
│   ├── configs/                   # Configuration files & Air templates
│   ├── internal/                  # Domain packages (auth, catalog, cart, orders)
│   ├── migrations/                # Database schema migration scripts
│   ├── pkg/                       # Shared packages (database, storage, logger)
│   └── docs/                      # Swagger OpenAPI specification
├── frontend/                      # Next.js 15 bilingual storefront
│   ├── messages/                  # Localization dictionaries (en.json, ar.json)
│   ├── public/                    # Static assets & placeholders
│   └── src/
│       ├── api/                   # API client configuration
│       ├── app/                   # Next.js App Router (shop, auth, admin)
│       ├── components/            # Reusable UI components (cards, navbar, layout)
│       ├── context/               # Global state (Cart, Catalog, Theme)
│       └── features/              # Feature modules (home, products, cart, etc.)
├── Makefile                       # Development workflow shortcuts
├── docker-compose.dev.yml         # Full local development container stack
└── README.md                      # Project documentation
```

---

## Quick Start (Docker Development)

Ensure you have [Docker](https://docs.docker.com/) and [Docker Compose](https://docs.docker.com/compose/) installed.

### 1. Launch the Stack
```bash
make up
# Or run in the background:
make up-d
```

### 2. Available Services & Ports

| Service | URL | Description |
|---|---|---|
| **Storefront (Frontend)** | [http://localhost:3000](http://localhost:3000) | Next.js app with EN/AR localization |
| **Backend API** | [http://localhost:8081](http://localhost:8081) | Go REST API Server |
| **API Health Check** | [http://localhost:8081/health](http://localhost:8081/health) | API Liveness endpoint |
| **Swagger Docs (Public)** | [http://localhost:8081/swagger/public/index.html](http://localhost:8081/swagger/public/index.html) | Storefront API Specs |
| **Swagger Docs (Admin)** | [http://localhost:8081/swagger/admin/index.html](http://localhost:8081/swagger/admin/index.html) | Admin Management API Specs |
| **MinIO Console** | [http://localhost:9001](http://localhost:9001) | S3 Object Browser (`admin` / `supersecret`) |
| **PostgreSQL** | `localhost:5432` | DB: `prim`, User: `prim`, Pass: `prim` |
| **Redis** | `localhost:6379` | In-memory cache & sessions |

---

## Development Commands

All standard development tasks are exposed via the top-level `Makefile`:

```bash
# Manage containers
make up           # Start all services with live logs
make up-d         # Start all services in the background
make down         # Stop all containers
make clean        # Stop and wipe persistent volumes

# Stream logs
make logs         # Stream all container logs
make logs-api     # Stream Go backend logs
make logs-frontend# Stream Next.js frontend logs

# Database migrations
make migrate-up   # Apply pending migrations
make migrate-down # Rollback last migration
make migrate-reset# Wipe and re-apply all migrations
```

---

## Features & Highlights

- **Bilingual & Bidirectional:** Fully functional English and Arabic interface with dynamic RTL direction flipping.
- **Card Click Isolation:** Decoupled product card image/title navigation from action buttons (Add to Cart / Wishlist) to prevent anchor hijacking.
- **Side Drawer Cart:** Instant cart slide-over with real-time calculations, coupon discounts, and line-item management.
- **Product Filters:** Multi-criteria filtering by category, brand, stock availability, ratings, and discounts.
- **Hot-Reloading Everywhere:** Next.js Turbopack on the frontend and Air on the Go backend for near-instant iteration inside Docker.

---

## License

This project is licensed under the MIT License.
