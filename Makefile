COMPOSE = docker compose -f docker-compose.dev.yml
PROD_COMPOSE = docker compose --env-file backend/.env -f backend/docker/docker-compose.prod.yml

.PHONY: up up-build up-d down clean clean-dev clean-prod clean-volumes-dev clean-volumes-prod \
	logs logs-api logs-worker logs-frontend \
	migrate-up migrate-down migrate-drop migrate-reset \
	prod-up prod-down prod-logs prod-ps prod-migrate

## ─── Full Stack ──────────────────────────────────────────

up:
	$(COMPOSE) up

up-build:
	$(COMPOSE) up --build

up-d:
	$(COMPOSE) up -d

down:
	$(COMPOSE) down

## ─── Cleanup ─────────────────────────────────────────────

# Safe cleanup: removes containers and networks, keeps all named volumes.
clean: clean-dev

clean-dev:
	$(COMPOSE) down --remove-orphans

clean-prod:
	$(PROD_COMPOSE) down --remove-orphans

# Explicitly destructive. Only use when you intend to delete persisted data.
clean-volumes-dev:
	@echo "WARNING: removing dev volumes (Postgres, Redis, MinIO, Kafka)"
	$(COMPOSE) down -v --remove-orphans

clean-volumes-prod:
	@echo "WARNING: removing the production MinIO volume"
	$(PROD_COMPOSE) down -v --remove-orphans

## ─── Logs ────────────────────────────────────────────────

logs:
	$(COMPOSE) logs -f

logs-api:
	$(COMPOSE) logs -f api

logs-worker:
	$(COMPOSE) logs -f worker

logs-frontend:
	$(COMPOSE) logs -f frontend

## ─── Migrations ──────────────────────────────────────────

migrate-up:
	$(COMPOSE) run --rm migrate

migrate-down:
	$(COMPOSE) run --rm migrate down

migrate-drop:
	$(COMPOSE) run --rm migrate drop -f

migrate-reset:
	$(COMPOSE) run --rm migrate drop -f
	$(COMPOSE) run --rm migrate

## ─── Production ──────────────────────────────────────────

build-backend-binaries:
	cd backend && CGO_ENABLED=0 GOOS=linux go build -o ./bin/server ./cmd/api/main.go && CGO_ENABLED=0 GOOS=linux go build -o ./bin/worker ./cmd/worker/main.go

# Builds the production images, runs pending migrations, then starts the
# production stack in the background. Requires backend/.env (including
# NEXT_PUBLIC_API_URL, DATABASE_URL, and REDIS_URL).
prod-up: build-backend-binaries
	$(PROD_COMPOSE) up -d --build

prod-down:
	$(PROD_COMPOSE) down

prod-logs:
	$(PROD_COMPOSE) logs -f

prod-ps:
	$(PROD_COMPOSE) ps

prod-migrate:
	$(PROD_COMPOSE) run --rm migrate
