.PHONY: up db-setup clear-db test-db-setup run stop logs build test

up:
	docker compose up -d --wait db

db-setup: up
	docker compose exec -T db psql -v ON_ERROR_STOP=1 -U postgres -d postgres -c "ALTER ROLE postgres WITH PASSWORD 'secret'"
	docker compose exec -T db psql -v ON_ERROR_STOP=1 -U postgres -d postgres -f /schema.sql

clear-db: up
	docker compose exec -T db psql -v ON_ERROR_STOP=1 -U postgres -d postgres -c "TRUNCATE TABLE tasks, users RESTART IDENTITY CASCADE"

test-db-setup: up
	docker compose exec -T db psql -v ON_ERROR_STOP=1 -U postgres -d postgres -c "ALTER ROLE postgres WITH PASSWORD 'secret'"

run: db-setup
	go run ./cmd

stop:
	docker compose down

logs:
	docker compose logs -f

build:
	go build ./...

test: test-db-setup
	go test ./...
