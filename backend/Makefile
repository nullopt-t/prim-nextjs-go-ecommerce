up:
	docker compose -f docker-compose.dev.yml up --build

down:
	docker compose -f docker-compose.dev.yml down

migrate-up:
	docker compose -f docker-compose.dev.yml run --rm migrate

migrate-down:
	docker compose -f docker-compose.dev.yml run --rm migrate down

migrate-drop:
	docker compose -f docker-compose.dev.yml run --rm migrate drop -f

migrate-reset:
	docker compose -f docker-compose.dev.yml run --rm migrate drop -f
	docker compose -f docker-compose.dev.yml run --rm migrate

migrate-force:
	docker compose -f docker-compose.dev.yml run --rm migrate force $(version)
