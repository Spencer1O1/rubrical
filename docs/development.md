# Development guide

A rubric-aware preflight checker for student assignments. See [specification.md](./specification.md) for the full specification and [spec-checklist.md](./spec-checklist.md) for MVP progress.

## Quick start (WSL)

### Prerequisites

| Tool | Purpose |
|------|---------|
| Go 1.26+ | Backend server |
| pnpm | Tailwind + extension build |
| Docker | Postgres (recommended) |

```bash
# From repo root
pnpm install
make setup

# Start Postgres and run migrations
make db-up
make migrate-up

# Run the app (separate terminals for watch mode)
make css-watch    # terminal 1
make templ-watch  # terminal 2
make server       # terminal 3
```

Open http://localhost:8787

On startup the server creates a **local dev user** (`local@rubrical.dev`) and assigns all imports to that user. The dashboard shows only that user's assignments. Existing rows with no `user_id` are backfilled on startup.

Build the browser extension:

```bash
make extension-build
# Load extension/ unpacked in Chrome → Extensions → Developer mode → Load unpacked
```

After changing extension code, rebuild and click **Reload** on the Rubrical card in `edge://extensions` (or `chrome://extensions`).

### WSL + browser on Windows

The Go server runs in WSL, but Edge/Chrome usually runs on Windows. The extension POSTs to `http://127.0.0.1:8787` and `http://localhost:8787`.

From **Windows** PowerShell, verify the server is reachable:

```powershell
curl http://localhost:8787/health -UseBasicParsing
```

You should get `{"status":"ok"}`. On WSL2, **`localhost` usually works from Windows but `127.0.0.1` often does not** — the extension tries `localhost` first.

If both fail while `make server` is running in WSL, restart WSL (`wsl --shutdown` in PowerShell, then reopen) or check Docker Desktop WSL integration.

After changing `extension/manifest.json` host permissions, **reload the extension** — permission changes are not picked up by rebuild alone.

## Docker on WSL

Rubrical uses Postgres via `docker compose`. You need a Docker daemon reachable from WSL.

### Option A: Docker Desktop (recommended on WSL2)

1. Install [Docker Desktop for Windows](https://docs.docker.com/desktop/setup/install/windows-install/).
2. Open Docker Desktop → **Settings** → **Resources** → **WSL integration**.
3. Enable integration for your WSL distro (e.g. Ubuntu).
4. Back in WSL, verify:

```bash
docker --version
docker compose version
```

Then:

```bash
make db-up
make migrate-up
```

### Option B: Docker Engine inside WSL

If you prefer not to use Docker Desktop:

```bash
sudo apt update
sudo apt install -y docker.io docker-compose-v2
sudo usermod -aG docker "$USER"
# Log out of WSL and back in, then:
sudo service docker start
docker --version
```

### Option C: Postgres without Docker

If you don't want Docker at all, install Postgres directly in WSL:

```bash
sudo apt update
sudo apt install -y postgresql postgresql-contrib
sudo service postgresql start

sudo -u postgres psql -c "CREATE USER rubrical WITH PASSWORD 'rubrical' CREATEDB;"
sudo -u postgres psql -c "CREATE DATABASE rubrical OWNER rubrical;"
```

Run migrations with the same connection string:

```bash
export DATABASE_URL='postgres://rubrical:rubrical@localhost:5432/rubrical?sslmode=disable'
make migrate-up
make server
```

## Stack

- Go + chi
- templ + HTMX + Tailwind
- PostgreSQL + goose + sqlc
- TypeScript browser extension

## Environment

| Variable | Default |
|----------|---------|
| `RUBRICAL_ADDR` | `:8787` |
| `DATABASE_URL` | `postgres://rubrical:rubrical@localhost:5432/rubrical?sslmode=disable` |
| `RUBRICAL_STRICT_EXTRACTION` | unset (fallbacks enabled); set `1` to disable extraction/display fallbacks |
| `RUBRICAL_DATA_DIR` | `./data` — draft file bytes on disk |
| `POST_DUE_DATE_RETENTION_TIME` | `168h` (1 week) — purge uploaded draft files this long after assignment `due_at`; set `0` to skip due-date rule |
| `POST_UPLOAD_RETENTION_TIME` | `720h` (30 days) — purge uploaded draft files this long after upload when assignment has no `due_at`; set `0` to skip |
| `OPENAI_API_KEY` | — (Phase 7) |
| `ANTHROPIC_API_KEY` | — (Phase 7) |

`.env.local` and `.env` at repo root are loaded automatically (`.env.local` first).

## Makefile commands

| Command | Description |
|---------|-------------|
| `make setup` | Install JS deps, generate templ/sqlc/css |
| `make db-up` | Start Postgres container |
| `make db-reset` | Drop public schema and re-run migrations (after squashing) |
| `make migrate-up` | Apply database migrations |
| `make server` | Run Go server on :8787 |
| `make purge` | One-shot purge of draft files past retention |
| `make css-watch` | Watch Tailwind CSS |
| `make templ-watch` | Watch templ templates |
| `make extension-build` | Build Chrome extension bundle |
| `make test` | Run Go tests |

## templ IDE errors

If Cursor/VS Code shows `undefined: templ.ResolveAttributeValue` on `.templ` files, the **templ LSP and go.mod version are out of sync**. Keep them aligned:

```bash
go install github.com/a-h/templ/cmd/templ@v0.3.1020   # match go.mod
make templ
```

Then **Developer: Reload Window** in Cursor. The code was always valid if `go build ./...` passed — this was an LSP false positive from a version mismatch.

See [spec-checklist.md](./spec-checklist.md) and specification §14 for the MVP build order.
