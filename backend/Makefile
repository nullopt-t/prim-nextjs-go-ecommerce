COMPOSE=docker compose -f docker-compose.dev.yml

.PHONY: up up-d down clean logs logs-api logs-worker migrate-up migrate-down migrate-drop migrate-reset migrate-force

up:
	$(COMPOSE) up --build

up-d:
	$(COMPOSE) up --build -d

down:
	$(COMPOSE) down

clean:
	$(COMPOSE) down -v

logs:
	$(COMPOSE) logs -f

logs-api:
	$(COMPOSE) logs -f api

logs-worker:
	$(COMPOSE) logs -f worker

migrate-up:
	$(COMPOSE) run --rm migrate

migrate-down:
	$(COMPOSE) run --rm migrate down

migrate-drop:
	$(COMPOSE) run --rm migrate drop -f

migrate-reset:
	$(COMPOSE) run --rm migrate drop -f
	$(COMPOSE) run --rm migrate

migrate-force:
	$(COMPOSE) run --rm migrate force $(version)
