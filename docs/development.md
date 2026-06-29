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
make setup          # includes pnpm setup:secrets-key

# Start Postgres and run migrations
make db-up
make migrate-up

# Run the app (separate terminals for watch mode)
make css-watch    # terminal 1
make templ-watch  # terminal 2
make server       # terminal 3
```

`make setup` runs `pnpm setup:secrets-key`, which creates `.env.local` (from `.env.example` if needed) and writes `RUBRICAL_SECRETS_ENCRYPTION_KEY`. The server requires this key to encrypt BYOK API keys at rest. Re-run `pnpm setup:secrets-key` only on a fresh machine — changing the key invalidates saved API keys.

Open http://localhost:8787 and **sign up** for an account (email/password, or Google OAuth when configured). Each user sees only their own imported assignments.

For Google sign-in, set `GOOGLE_OAUTH_CLIENT_ID`, `GOOGLE_OAUTH_CLIENT_SECRET`, and `RUBRICAL_PUBLIC_URL` in `.env.local`. Password reset emails use `EMAIL_DEV_LOG=1` in dev (logged to the server console) or Resend/SMTP in production.

The browser extension sends session cookies with API requests — sign in via the web app (same origin as `RUBRICAL_PUBLIC_URL` or localhost) before importing from Canvas.

Build the browser extension:

```bash
make extension-build
# Load extension/ unpacked in Chrome → Extensions → Developer mode → Load unpacked
```

After changing extension code, rebuild and click **Reload** on the Rubrical card in `edge://extensions` (or `chrome://extensions`).

### WSL + browser on Windows

The Go server runs in WSL, but Edge/Chrome usually runs on Windows. Local dev builds of the extension talk to `http://localhost:8787` (see `make extension-build`).

From **Windows** PowerShell, verify the server is reachable:

```powershell
curl http://localhost:8787/health -UseBasicParsing
```

You should get `{"status":"ok"}`. On WSL2, **`localhost` usually works from Windows but `127.0.0.1` often does not** — the dev extension uses `localhost`.

For a production build (`make extension-build-prod`), the extension talks to `https://rubrical.spencerls.dev` only.

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
| `make extension-build` | Build extension for local dev (`http://localhost:8787`) |
| `make extension-build-prod` | Build extension for production (`https://rubrical.spencerls.dev`) |
| `make test` | Run Go tests |

## templ IDE errors

If Cursor/VS Code shows `undefined: templ.ResolveAttributeValue` on `.templ` files, the **templ LSP and go.mod version are out of sync**. Keep them aligned:

```bash
go install github.com/a-h/templ/cmd/templ@v0.3.1020   # match go.mod
make templ
```

Then **Developer: Reload Window** in Cursor. The code was always valid if `go build ./...` passed — this was an LSP false positive from a version mismatch.

See [spec-checklist.md](./spec-checklist.md) and specification §14 for the MVP build order.

## Browser extension

| Piece | Role |
|-------|------|
| **Content script** (`content.ts`) | Canvas DOM hooks, file staging (IndexedDB on the Canvas origin), import capture, row indicators |
| **Service worker** (`background.ts`) | Proxies Rubrical API `fetch` / multipart from Canvas pages (Private Network Access to localhost) |
| **Popup** | Extension settings |

**File staging:** when a student picks a file on Canvas, the content script stores the `Blob` in IndexedDB via `staged-files/store.ts` (Canvas site origin). After **Check with Rubrical**, assignment files upload via multipart (`POST …/draft/upload`), not inline JSON.

Rebuild after extension changes: `make extension-build`, then reload the extension in the browser.
