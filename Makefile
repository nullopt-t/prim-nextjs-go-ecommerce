COMPOSE = docker compose -f docker-compose.dev.yml

.PHONY: up up-build up-d down clean logs logs-api logs-frontend migrate-up migrate-down migrate-drop migrate-reset

## ─── Full Stack ──────────────────────────────────────────

up:
	$(COMPOSE) up

up-build:
	$(COMPOSE) up --build

up-d:
	$(COMPOSE) up -d

down:
	$(COMPOSE) down

clean:
	$(COMPOSE) down -v

## ─── Logs ────────────────────────────────────────────────

logs:
	$(COMPOSE) logs -f

logs-api:
	$(COMPOSE) logs -f api

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
