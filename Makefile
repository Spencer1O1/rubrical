.PHONY: dev server purge css css-watch templ templ-watch db-up db-down db-reset migrate-up migrate-down sqlc extension-build extension-build-prod vercel-build test build tidy setup-secrets-key

DATABASE_URL ?= postgres://rubrical:rubrical@localhost:5432/rubrical?sslmode=disable

dev:
	@echo "Run in separate terminals:"
	@echo "  make db-up && make migrate-up"
	@echo "  make css-watch"
	@echo "  make templ-watch"
	@echo "  make server"

server:
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/server

purge:
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/purge

build:
	go build -o bin/rubrical ./cmd/server
	go build -o bin/rubrical-purge ./cmd/purge

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

# Wipe public schema and re-apply 00001. Use after squashing migrations (goose version
# can point at a deleted 00002, which breaks migrate-down).
db-reset:
	docker compose exec -T postgres psql -U rubrical -d rubrical -v ON_ERROR_STOP=1 -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public; GRANT ALL ON SCHEMA public TO rubrical; GRANT ALL ON SCHEMA public TO public;"
	$(MAKE) migrate-up

sqlc:
	go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.29.0 generate

extension-build:
	pnpm --filter rubrical-extension typecheck
	pnpm --filter rubrical-extension build:dev

extension-build-prod:
	pnpm --filter rubrical-extension typecheck
	pnpm --filter rubrical-extension build

vercel-build: css templ
	go run ./cmd/export-landing

install-js:
	pnpm install

setup: install-js tidy templ sqlc css setup-secrets-key
	@echo "Skeleton ready. Start Postgres with: make db-up && make migrate-up"

setup-secrets-key:
	pnpm setup:secrets-key

test:
	go test ./...

install-tools:
	go install github.com/a-h/templ/cmd/templ@v0.3.1020
	go install github.com/pressly/goose/v3/cmd/goose@v3.24.1
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.29.0
