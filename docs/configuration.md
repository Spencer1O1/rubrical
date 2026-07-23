# Configuration

## Where config lives

| Environment | Files | Notes |
|-------------|-------|-------|
| **Laptop (dev)** | `.env.local` (from [`.env.example`](../.env.example)) | Gitignored. Never used on the homeserver. |
| **Homeserver (prod)** | [`homeserver.yaml`](../homeserver.yaml) → `/etc/homeserver/apps/rubrical.env` | Secrets prompted / keys generated on `install-pack`. |
| **Homeserver (shared)** | `/etc/homeserver/server.env` | Listen host/port (`RUBRICAL_*`, `POSTGRES_HOST`/`PORT`). |
| **Homeserver (DB role)** | `/etc/homeserver/databases/rubrical.env` | `POSTGRES_USER` / `PASSWORD` / `DB` (provisioned). |

Defaults for optional knobs live in [`internal/config/defaults.go`](../internal/config/defaults.go). `config.Load` reads ENV only for operator-tunable settings.

## Per-user AI (database: `user_ai_settings`)

Not ENV. Each user sets provider/model/API keys in **Settings** (dashboard, embed, or extension). Keys are AES-GCM encrypted with `SECRETS_ENCRYPTION_KEY`.

```bash
# local only:
pnpm setup:secrets-key   # writes SECRETS_ENCRYPTION_KEY to .env.local
# prod: generated_keys in homeserver.yaml
```

Default provider/model when empty: `openai` / `gpt-4o-mini` (Anthropic: `claude-sonnet-5`).

## Auth + email

| Variable | Dev (`.env.local`) | Prod (`homeserver.yaml`) |
|----------|--------------------|--------------------------|
| `PUBLIC_URL` | `http://localhost:8787` | `env:` |
| `GOOGLE_OAUTH_CLIENT_ID` | optional | `env:` with empty value → prompt on install (public; blank = hide Google) |
| `GOOGLE_OAUTH_CLIENT_SECRET` | optional | `secrets:` (prompt; leave empty if unused) |
| `EXTENSION_ORIGINS` | optional | `env:` |
| `EMAIL_FROM` | optional | `env:` |
| `EMAIL_DEV_LOG` | `1` (stdout) | unset (real send) |
| `RESEND_API_KEY` | omit when logging | `secrets:` |
| `SMTP_*` | alternative to Resend | usually omit if using Resend |

## Operator limits

Production values live in [`homeserver.yaml`](../homeserver.yaml) `env:`. Local overrides go in `.env.local`. Unset → code defaults.

| Variable | Default | Purpose |
|----------|---------|---------|
| `DRAFT_MAX_UPLOAD_BYTES` | 500 MiB | Max bytes per uploaded blob |
| `DRAFT_MAX_UPLOAD_SLOTS` | 20 | Max top-level attachments per draft |
| `ANALYSIS_MAX_SUBMISSION_TEXT_CHARS` | 120000 | Shared char pool for student text in prompt |
| `ANALYSIS_MAX_TOTAL_BYTES` | 64 MiB | Max file bytes sent to the model |
| `AI_ENFORCE_RATE_LIMITS` | off | Enable per-user analyze rate limits |
| `AI_MAX_RUNS_PER_HOUR` / `_DAY` | 0 (unlimited) | Rate limits when enforced |
| `AI_MIN_SECONDS_BETWEEN_RUNS` | 0 | Min gap between analyzes |
| `POST_DUE_DATE_RETENTION_TIME` | `168h` | Purge after due date |
| `POST_UPLOAD_RETENTION_TIME` | `720h` | Purge after upload (no due date) |
| `STRICT_EXTRACTION` | off | Dev: disable Canvas fallbacks |
| `ALLOW_LOCAL_URL_FETCH` | off | Dev: allow fetching localhost URLs |

## Listen + database

| Variable | Where |
|----------|-------|
| `RUBRICAL_HOST` / `RUBRICAL_PORT` | Dev: `.env.local`. Prod: `server.env` (from app key `rubrical`). |
| `POSTGRES_HOST` / `PORT` | Dev: `.env.local`. Prod: `server.env`. |
| `POSTGRES_USER` / `PASSWORD` / `DB` / `SSLMODE` | Dev: `.env.local`. Prod: `databases/rubrical.env`. |

## Limit model

```
┌─────────────┐     ┌──────────────┐     ┌─────────────────────┐
│   INGRESS   │ ──▶ │   STORAGE    │ ──▶ │      ANALYSIS       │
│ import JSON │     │ draft files  │     │  send to AI model   │
└─────────────┘     └──────────────┘     └─────────────────────┘
  hardcoded           DRAFT_* ENV           ANALYSIS_* ENV
```

**Analysis has two tunable limits:** submission text chars, and total file bytes to the model. Per-user API keys are not ENV.

## Hardcoded (not ENV)

Import field caps, zip extraction safety, provider base URLs/timeouts — change in `defaults.go` / product code.
