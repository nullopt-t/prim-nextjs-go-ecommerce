COMPOSE=docker compose -f docker/docker-compose.dev.yml

# Swagger tags used exclusively by /admin routes, derived from the source
# annotations so the public and admin specs stay disjoint. Every admin handler
# must keep a single "Admin <Resource>" tag; a tag shared with a storefront route
# would leak either way, so give the admin routes their own tag instead.
ADMIN_TAGS := $(shell grep -rhoE '@Tags[[:space:]]+Admin [A-Za-z ]+' --include='*_handler.go' . | sed -E 's/@Tags[[:space:]]+//; s/[[:space:]]+$$//' | sort -u | paste -sd, -)
# The same tags negated: swag excludes every tag prefixed with '!'.
PUBLIC_EXCLUDE_TAGS := $(shell printf '%s' '$(ADMIN_TAGS)' | sed 's/^/!/; s/,/,!/g')

.PHONY: up up-d down clean logs logs-api logs-worker migrate-up migrate-down migrate-drop migrate-reset migrate-force swagger

swagger:
	@test -n "$(ADMIN_TAGS)" || { echo "swagger: no 'Admin *' tags found, refusing to generate specs" >&2; exit 1; }
	# --parseDependency is deliberately omitted: swag v1.16.4 cannot parse the
	# Go 1.27 standard library, so dependency parsing fails on /usr/lib/go/src
	# sources (internal/poll, math/rand/v2, net/http/internal/http2). Types used
	# in annotations must therefore resolve without it: reference in-repo DTOs,
	# and give stdlib-typed fields (e.g. json.RawMessage) a swaggertype:"..." tag.
	swag init -g cmd/api/main.go -o docs/public --instanceName public -t "$(PUBLIC_EXCLUDE_TAGS)" --useStructName
	swag init -g cmd/api/main.go -o docs/admin --instanceName admin -t "$(ADMIN_TAGS)" --useStructName

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
