.PHONY: dev server purge css css-watch templ templ-watch db-up db-down db-reset migrate-up migrate-down sqlc extension-build extension-package test build tidy setup-secrets-key

# Goose needs a URL. No Makefile defaults — load .env.local/.env, honor make/env overrides, then require every piece.
define goose_postgres
	@bash -euo pipefail -c '\
	  set -a; \
	  [ -f .env.local ] && source ./.env.local; \
	  [ -f .env ] && source ./.env; \
	  set +a; \
	  [ -n "$(POSTGRES_HOST)" ] && export POSTGRES_HOST="$(POSTGRES_HOST)"; \
	  [ -n "$(POSTGRES_PORT)" ] && export POSTGRES_PORT="$(POSTGRES_PORT)"; \
	  [ -n "$(POSTGRES_USER)" ] && export POSTGRES_USER="$(POSTGRES_USER)"; \
	  [ -n "$(POSTGRES_PASSWORD)" ] && export POSTGRES_PASSWORD="$(POSTGRES_PASSWORD)"; \
	  [ -n "$(POSTGRES_DB)" ] && export POSTGRES_DB="$(POSTGRES_DB)"; \
	  [ -n "$(POSTGRES_SSLMODE)" ] && export POSTGRES_SSLMODE="$(POSTGRES_SSLMODE)"; \
	  missing=""; \
	  for v in POSTGRES_HOST POSTGRES_PORT POSTGRES_USER POSTGRES_PASSWORD POSTGRES_DB POSTGRES_SSLMODE; do \
	    if [ -z "$${!v-}" ]; then missing="$$missing $$v"; fi; \
	  done; \
	  if [ -n "$$missing" ]; then echo "error: missing required env:$$missing" >&2; exit 1; fi; \
	  url="postgres://$${POSTGRES_USER}:$${POSTGRES_PASSWORD}@$${POSTGRES_HOST}:$${POSTGRES_PORT}/$${POSTGRES_DB}?sslmode=$${POSTGRES_SSLMODE}"; \
	  go run github.com/pressly/goose/v3/cmd/goose@v3.24.1 -dir migrations postgres "$$url" $(1)'
endef

dev:
	@echo "Run in separate terminals:"
	@echo "  make db-up && make migrate-up"
	@echo "  make css-watch"
	@echo "  make templ-watch"
	@echo "  make server"

server:
	go run ./cmd/server

purge:
	go run ./cmd/purge

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
	$(call goose_postgres,up)

migrate-down:
	$(call goose_postgres,down)

# Wipe public schema and re-apply 00001. Use after squashing migrations (goose version
# can point at a deleted 00002, which breaks migrate-down).
db-reset:
	docker compose exec -T postgres psql -U rubrical -d rubrical -v ON_ERROR_STOP=1 -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public; GRANT ALL ON SCHEMA public TO rubrical; GRANT ALL ON SCHEMA public TO public;"
	$(MAKE) migrate-up

sqlc:
	go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.29.0 generate

# PUBLIC_URL from .env.local/.env (default http://localhost:8787). Homeserver sets
# PUBLIC_URL=https://rubrical.spencerls.dev on deploy.
define extension_with_public_url
	@bash -euo pipefail -c '\
	  set -a; \
	  [ -f .env.local ] && source ./.env.local; \
	  [ -f .env ] && source ./.env; \
	  set +a; \
	  export PUBLIC_URL="$${PUBLIC_URL:-http://localhost:8787}"; \
	  echo "$(1) PUBLIC_URL=$$PUBLIC_URL"; \
	  pnpm --filter rubrical-extension typecheck; \
	  PUBLIC_URL="$$PUBLIC_URL" pnpm --filter rubrical-extension build; \
	  $(2)'
endef

extension-build:
	$(call extension_with_public_url,extension-build,true)

# Zip for /install at static/downloads/rubrical-extension.zip (same PUBLIC_URL rules).
extension-package:
	$(call extension_with_public_url,extension-package,python3 scripts/package-extension.py)

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
