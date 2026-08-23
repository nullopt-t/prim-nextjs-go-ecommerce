COMPOSE=docker compose -f docker-compose.dev.yml

.PHONY: up up-d down clean logs logs-api logs-worker migrate-up migrate-down migrate-drop migrate-reset migrate-force swagger

swagger:
	swag init -g cmd/api/main.go -o docs/public --instanceName public -t "!Admin Orders,!Admin Categories,!Admin Product Variants,!Admin Variant Media,!Admin Variant Inventory,!Admin Tags,!Admin Reviews,!Admin Products,!Admin Product Media,!Admin Brands"
	swag init -g cmd/api/main.go -o docs/admin --instanceName admin -t "Admin Orders,Admin Categories,Admin Product Variants,Admin Variant Media,Admin Variant Inventory,Admin Tags,Admin Reviews,Admin Products,Admin Product Media,Admin Brands"

up:
	$(COMPOSE) up

up-build:
	$(COMPOSE) up --build

build:
	CGO_ENABLED=0 GOOS=linux go build -o ./bin/server ./cmd/api/main.go
	CGO_ENABLED=0 GOOS=linux go build -o ./bin/worker ./cmd/worker/main.go

up-d:
	$(COMPOSE) up -d

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
