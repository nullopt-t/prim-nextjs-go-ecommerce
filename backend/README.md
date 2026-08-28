
# PRIM Backend

Go backend service providing REST APIs, background workers, authentication, catalog management, and media storage.

## Tech Stack

- **Language & Framework:** Go, Gin
- **Data & Cache:** PostgreSQL, Redis
- **Object Storage:** MinIO (S3-compatible)
- **Hot Reload:** Air
- **Migrations:** golang-migrate
- **Auth & Docs:** JWT, Swag (Swagger/OpenAPI)

---

## Quick Start

```bash
# Start all development services (API, DB, Redis, MinIO, Migrations)
make up

# Or run detached
make up-d
```

Services will be available at:
- **API Server:** `http://localhost:8080`
- **Swagger Docs (Public):** `http://localhost:8080/swagger/public/index.html`
- **Swagger Docs (Admin):** `http://localhost:8080/swagger/admin/index.html`
- **MinIO Console:** `http://localhost:9001` (User: `admin`, Pass: `supersecret`)

---

## Commands Reference

| Task | Command |
|---|---|
| **Start Dev** | `make up` (or `make up-d` / `make up-build`) |
| **Stop Dev** | `make down` |
| **Clean Volumes** | `make clean` |
| **View Logs** | `make logs` / `make logs-api` / `make logs-worker` |
| **Run Migrations** | `make migrate-up` / `make migrate-down` / `make migrate-reset` |
| **Generate Swagger** | `make swagger` |
| **Build Binaries** | `make build` |
| **Run Tests** | `go test ./...` |

---

## Project Structure

```
backend/
├── cmd/                # Entrypoints (api, worker)
├── internal/           # Core domain logic (auth, catalog, order, etc.)
├── pkg/                # Reusable packages (config, database, log, storage)
├── configs/            # Configs (config.yaml, air.api.toml, air.worker.toml)
├── docker/             # Dockerfiles & compose manifests (dev & prod)
├── migrations/         # SQL migration scripts
└── docs/               # Auto-generated Swagger documentation
```

---

## Environment Variables

Sensitive or deployment-specific variables can be configured via `.env`:

```env
# Optional SMTP email settings
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@example.com
SMTP_PASSWORD=your-app-password
SMTP_FROM_EMAIL=your-email@example.com
```
