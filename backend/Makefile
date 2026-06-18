DB_URL=postgres://prim:prim@localhost:5432/prim?sslmode=disable
MIGRATIONS=./migrations

hot:
	air -c .air.toml

run:
	@go run cmd/api/main.go

build:
	@go build -o tmp/bin cmd/api/main.go

run-worker:
	@go run cmd/worker/main.go

all: hot-reload

migrate-new:
	migrate create -ext sql -dir $(MIGRATIONS) -seq $(name)

migrate-up:
	migrate -path $(MIGRATIONS) -database "$(DB_URL)" up

migrate-down:
	migrate -path $(MIGRATIONS) -database "$(DB_URL)" down

migrate-drop:
	migrate -path $(MIGRATIONS) -database "$(DB_URL)" drop -f

migrate-reset:
	migrate -path $(MIGRATIONS) -database "$(DB_URL)" drop -f
	migrate -path $(MIGRATIONS) -database "$(DB_URL)" up

.PHONY: migrate-new migrate-up migrate-down migrate-drop migrate-reset hot run-worker all build
