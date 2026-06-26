
# PRIM Backend

Backend service for the PRIM platform, built with Go. The project provides a REST API, background workers, authentication, email notifications, and a complete Docker-based development environment.

## Tech Stack

* Go
* Gin
* PostgreSQL
* Redis
* Docker & Docker Compose
* Air (hot reload)
* golang-migrate
* JWT Authentication

---

## Prerequisites

* Docker
* Docker Compose
* Make

No local installation of Go, PostgreSQL, Redis, or golang-migrate is required for development.

---

## Getting Started

Clone the repository:

```bash
git clone <repository-url>
cd backend
```

Start the development environment:

```bash
make up
```

This command starts:

* API Server
* Worker
* PostgreSQL
* Redis

---

## Available Commands

### Start the application

```bash
make up
```

Run in detached mode:

```bash
make up-d
```

Stop all services:

```bash
make down
```

Remove all containers and volumes:

```bash
make clean
```

---

## Database Migrations

Apply all migrations:

```bash
make migrate-up
```

Rollback the last migration:

```bash
make migrate-down
```

Drop the database:

```bash
make migrate-drop
```

Reset the database:

```bash
make migrate-reset
```

Force the migration version:

```bash
make migrate-force version=<version>
```

---

## Viewing Logs

View logs from all services:

```bash
make logs
```

View API logs:

```bash
make logs-api
```

View Worker logs:

```bash
make logs-worker
```

---

## Development

The project uses **Air** for hot reloading.

Any changes made to the source code are automatically rebuilt and restarted inside the development containers.

---

## Project Structure

```
cmd/
├── api/
└── worker/

internal/
pkg/
migrations/

Dockerfile
Dockerfile.dev
docker-compose.dev.yml
```

---

## Production

A production Dockerfile is included for building optimized deployment images.

---

## Environment Variables

Development uses Docker Compose for most configuration.

Sensitive values such as SMTP credentials should be provided locally and must not be committed to the repository.

Example:

```env
SMTP_USERNAME=your-email@example.com
SMTP_PASSWORD=your-app-password
SMTP_FROM_EMAIL=your-email@example.com
```

---

## License

This project is intended for the PRIM platform.
