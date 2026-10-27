COMPOSE = docker compose --env-file backend/.env -f backend/docker/docker-compose.prod.yml

.PHONY: up up-build up-d down clean clean-volumes logs logs-api logs-worker logs-frontend \
	migrate-up migrate-down migrate-drop migrate-reset \
	prod-up prod-down prod-ps prod-logs prod-migrate

## ─── Stack ───────────────────────────────────────────────

# Foreground / detached startup of the production stack. Images are rebuilt
# when they do not exist yet; use `up-build` to force a rebuild.
up:
	$(COMPOSE) up

up-build:
	$(COMPOSE) up --build

up-d:
	$(COMPOSE) up -d --build

down:
	$(COMPOSE) down

## ─── Cleanup ─────────────────────────────────────────────

# Safe cleanup: removes containers and networks, keeps all named volumes.
clean:
	$(COMPOSE) down --remove-orphans

# Explicitly destructive. Only use when you intend to delete persisted data.
clean-volumes:
	@echo "WARNING: removing stack volumes (Postgres, Redis, MinIO, Kafka)"
	$(COMPOSE) down -v --remove-orphans

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

MIGRATE = $(COMPOSE) run --rm --entrypoint migrate migrate -path /migrations -database "$${DATABASE_URL}"

migrate-up:
	$(MIGRATE) up

migrate-down:
	$(MIGRATE) down 1

migrate-drop:
	$(MIGRATE) drop -f

migrate-reset:
	$(MIGRATE) drop -f
	$(MIGRATE) up

## ─── Production aliases ──────────────────────────────────

prod-up:
	$(COMPOSE) up -d --build

prod-down:
	$(COMPOSE) down

prod-ps:
	$(COMPOSE) ps

prod-logs:
	$(COMPOSE) logs -f

prod-migrate:
	$(MIGRATE) up
