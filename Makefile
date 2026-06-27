.PHONY: dev server css css-watch templ templ-watch db-up db-down migrate-up migrate-down sqlc extension-build test build tidy

DATABASE_URL ?= postgres://rubrical:rubrical@localhost:5432/rubrical?sslmode=disable

dev:
	@echo "Run in separate terminals:"
	@echo "  make db-up && make migrate-up"
	@echo "  make css-watch"
	@echo "  make templ-watch"
	@echo "  make server"

server:
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/server

build:
	go build -o bin/rubrical ./cmd/server

tidy:
	go mod tidy

templ:
	go run github.com/a-h/templ/cmd/templ@v0.3.1020 generate

templ-watch:
	go run github.com/a-h/templ/cmd/templ@v0.3.1020 generate --watch

css:
	pnpm run css:build

css-watch:
	pnpm run css:watch

db-up:
	docker compose up -d postgres

db-down:
	docker compose down

migrate-up:
	go run github.com/pressly/goose/v3/cmd/goose@v3.24.1 -dir migrations postgres "$(DATABASE_URL)" up

migrate-down:
	go run github.com/pressly/goose/v3/cmd/goose@v3.24.1 -dir migrations postgres "$(DATABASE_URL)" down

sqlc:
	go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.29.0 generate

extension-build:
	pnpm --filter rubrical-extension typecheck
	pnpm --filter rubrical-extension build

install-js:
	pnpm install

setup: install-js tidy templ sqlc css
	@echo "Skeleton ready. Start Postgres with: make db-up && make migrate-up"

test:
	go test ./...

install-tools:
	go install github.com/a-h/templ/cmd/templ@v0.3.1020
	go install github.com/pressly/goose/v3/cmd/goose@v3.24.1
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.29.0
