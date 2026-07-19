# Configuration

Rubrical settings fall into three buckets.

**Where defaults live:** [`internal/config/defaults.go`](../internal/config/defaults.go) is the single source of truth for every default value. [`internal/config/config.go`](../internal/config/config.go) reads **only operator-tunable** settings from ENV; everything else uses constants from `defaults.go` directly in code.

## Per-user (database: `user_ai_settings`)

Configured in **Settings** (dashboard, Canvas embed `/settings?embed=1`, or extension toolbar popup).

| Setting | Description |
|---------|-------------|
| `provider` | `openai` or `anthropic` |
| `model` | Model id for the active provider |
| `openai_api_key` | BYOK OpenAI key (encrypted at rest) |
| `anthropic_api_key` | BYOK Anthropic key (encrypted at rest) |

Keys are encrypted with **AES-256-GCM** before being written to PostgreSQL. Generate a stable operator key during setup:

```bash
pnpm setup:secrets-key   # writes RUBRICAL_SECRETS_ENCRYPTION_KEY to .env.local
```

The server refuses to start without this key. The settings API never returns raw keys — only `openaiApiKeyConfigured` / `anthropicApiKeyConfigured` booleans.

Default provider/model when empty: `openai` / `gpt-4o-mini` (Anthropic: `claude-sonnet-4-20250514`).

## Auth (ENV)

| Variable | Description |
|----------|-------------|
| `RUBRICAL_PUBLIC_URL` | Canonical site origin for OAuth redirects and password-reset links (default `http://localhost:8787`) |
| `GOOGLE_OAUTH_CLIENT_ID` / `GOOGLE_OAUTH_CLIENT_SECRET` | Google OAuth (optional; omit to hide “Continue with Google”) |
| `RUBRICAL_EXTENSION_ORIGINS` | Comma-separated `chrome-extension://…` origins allowed by CORS with credentials |
| `SESSION_TTL` | Session cookie lifetime (default 720h) |
| `EMAIL_DEV_LOG` | Log outbound email to server stdout instead of sending |
| `EMAIL_FROM` | From address for password reset |
| `RESEND_API_KEY` or `SMTP_*` | Outbound email for password reset |

Sign up at `/login?mode=signup` before importing assignments. The extension sends the session cookie (`credentials: include`) to the API.

## The limit model (read this first)

Everything falls into **three lifecycle stages**. Only the middle column is “analysis tuning.”

```
┌─────────────┐     ┌──────────────┐     ┌─────────────────────┐
│   INGRESS   │ ──▶ │   STORAGE    │ ──▶ │      ANALYSIS       │
│ import JSON │     │ draft files  │     │  send to AI model   │
└─────────────┘     └──────────────┘     └─────────────────────┘
  hardcoded           DRAFT_* ENV           ANALYSIS_* ENV
  field caps          upload bytes/slots    file bytes + text chars
```

| Stage | What | Knobs | Purpose |
|-------|------|-------|---------|
| **Ingress** | Extension `POST /imports` JSON | Hardcoded in `defaults.go` | Stop garbage/huge payloads at the door (512KB instructions, rubric row counts, etc.) |
| **Storage** | Files on disk + draft rows | `DRAFT_MAX_UPLOAD_BYTES`, `DRAFT_MAX_UPLOAD_SLOTS` | Accept what students submit; don’t fill your disk |
| **Analysis** | Provider request | `ANALYSIS_MAX_TOTAL_BYTES`, `ANALYSIS_MAX_SUBMISSION_TEXT_CHARS` | Cap AI cost |

**Analysis has exactly two tunable limits:**

1. **`ANALYSIS_MAX_TOTAL_BYTES`** — sum of **binary file payloads** sent to the model (PDF/image attachments, etc.). Inline text sources also pass through the file pipeline first; this caps their raw bytes before extraction.
2. **`ANALYSIS_MAX_SUBMISSION_TEXT_CHARS`** — one **shared pool** for **student submission text** in the prompt: typed draft + fetched URL text + inline extracted file text. First content wins; later content truncates when the pool runs out.

**Not in the submission text pool:** instructions, rubric, assignment title/context, file tree manifests, attachment index lines, skipped-file notes. Instructions/rubric are already bounded at ingress (import field caps). Manifest trees and skipped-file notes share a separate hardcoded cap (`DefaultAnalysisMaxManifestChars`, 32 000 runes).

### Provider file routing asymmetry

OpenAI and Anthropic accept different native attachment types. Logic lives in `internal/analysis/files/`.

| File kind | OpenAI | Anthropic |
|-----------|--------|-----------|
| PDF | Native PDF attachment | Native PDF attachment |
| Images (png, jpg, …) | Native image attachment | Native image attachment |
| docx, plain text, code, markdown, html, json, xml | Native provider file upload | Inline text in prompt (docx extracted) |
| xlsx, pptx, other Office | Native provider file upload | Skipped — use docx/pdf or switch to OpenAI |
| Legacy `.doc` | Skipped | Skipped |
| zip | Expanded in pipeline (depth/size limits) | Expanded in pipeline (depth/size limits) |
| exe, media, rar/7z | Skipped with note | Skipped with note |

When a file is skipped, a note is included in the prompt so the model knows it was not analyzed.

## Server ENV (operator / deployment)

Override defaults at deploy time. See [`.env.example`](../.env.example).

Naming (`RUBRICAL_*` vs `POSTGRES_*` / `MINIO_*` / `STORAGE_*`) is a homeserver rule — see WorkForce `docs/HOMESERVER.md` §8 (shared across all apps on the server).

| Variable | Default (in code) | Purpose |
|----------|-------------------|---------|
| `RUBRICAL_HOST` | empty (all interfaces) | HTTP listen host; production: `127.0.0.1` in `/etc/homeserver/server.env` |
| `RUBRICAL_PORT` | `8787` | HTTP listen port; same `server.env` as Caddy reverse_proxy |
| `POSTGRES_HOST` / `POSTGRES_PORT` | required | Shared Postgres listen — production: `/etc/homeserver/server.env` |
| `POSTGRES_USER` / `PASSWORD` / `DB` | required | Per-app role; API builds URL with host/port from above |
| `POSTGRES_SSLMODE` | required | Homeserver loopback: `disable` |
| `RUBRICAL_DATA_DIR` | `./data` | Draft file storage on disk |
| `RUBRICAL_STRICT_EXTRACTION` | off | Dev: disable Canvas fallbacks |
| `POST_DUE_DATE_RETENTION_TIME` | `168h` | Purge draft files after due date |
| `POST_UPLOAD_RETENTION_TIME` | `720h` | Purge draft files after upload (no due date) |
| `DRAFT_MAX_UPLOAD_BYTES` | 500 MiB | Max bytes per uploaded blob (storage) |
| `DRAFT_MAX_UPLOAD_SLOTS` | 20 | Max top-level attachments per draft (a zip = one slot) |
| `ANALYSIS_MAX_SUBMISSION_TEXT_CHARS` | 120000 | Shared char pool for student submission text in prompt |
| `ANALYSIS_MAX_TOTAL_BYTES` | 64 MiB | Max total file bytes sent to the model per analyze |
| `AI_ENFORCE_RATE_LIMITS` | off | Enable per-user analyze rate limits |
| `AI_MAX_RUNS_PER_HOUR` | 0 (unlimited) | Rate limit |
| `AI_MAX_RUNS_PER_DAY` | 0 (unlimited) | Rate limit |
| `AI_MIN_SECONDS_BETWEEN_RUNS` | 0 | Min gap between analyzes |

## Hardcoded (not ENV, not per-user)

Change in `defaults.go` or product code — not runtime ENV.

| Area | Examples |
|------|----------|
| Import field caps | 512KB instructions/draft text, rubric row limits, 8MB JSON body |
| Analysis prompt caps | 32 000 runes for manifest trees + skipped-file notes |
| Zip extraction safety | Max nesting depth (2), max total uncompressed bytes per archive (128 MiB) |
| Provider API wiring | Base URLs, 120s provider timeout, OpenAI temperature 0.2 |

**Canvas context:** Canvas file-upload assignments allow up to **5 GiB per file**; media recordings up to **500 MiB**. Rubrical defaults to **500 MiB** per upload blob — enough for typical assignments without storing multi-gigabyte files locally.

## What is not configurable

- Canvas capture behavior (extension code)
- Output JSON schema for analysis (structured outputs)
- File type routing rules (OpenAI native vs Anthropic inline) — logic in `internal/analysis/files/`
