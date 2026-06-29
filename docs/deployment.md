# Deployment

Rubrical runs as a **Go server** (Postgres, file storage, HTMX app, extension API). The marketing landing page can also be exported as static HTML for [Vercel](https://vercel.com).

## Full stack (recommended for the app)

Run the Go server on your host (Fly.io, Railway, a VPS, WSL dev, etc.):

```bash
make db-up && make migrate-up
make css templ
make server
```

Set env vars from [configuration.md](./configuration.md). Production example:

```bash
RUBRICAL_PUBLIC_URL=https://rubrical.spencerls.dev
DATABASE_URL=postgres://...
```

Routes:

| Path | Purpose |
|------|---------|
| `/` | Public marketing landing |
| `/login` | Sign in / sign up |
| `/dashboard` | Your imported assignments (auth required) |
| `/assignments/*` | Assignment detail & analysis |

## Vercel (marketing site + proxy to app)

Use Vercel for the **landing page** at your domain root. App routes are rewritten to your Go backend (`rubrical.spencerls.dev` in `vercel.json` — change if your API lives elsewhere).

### Build locally

```bash
make vercel-build
# Output: public/index.html + public/static/css/app.css
```

### Deploy

1. Connect the repo to Vercel.
2. Vercel uses `vercel.json` (`buildCommand`: `make vercel-build`, `outputDirectory`: `public`).
3. Point `rubrical.spencerls.dev` (or your marketing domain) at the Vercel project.
4. Keep the Go API running at the URL used in `vercel.json` rewrites.

Signed-in users on the static landing still link to `/login` and `/dashboard`; those paths proxy to the backend.

### Same origin (no Vercel split)

If the Go server serves everything (landing + app), skip Vercel and deploy only the Go binary. The landing at `/` is included automatically.

## Extension

Build for production API base:

```bash
make extension-build-prod
```

Load unpacked or publish to the Chrome Web Store. Set `RUBRICAL_EXTENSION_ORIGINS` on the server to your extension origin.
