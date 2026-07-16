# Deployment

Production hosting for Rubrical is the **home server** pattern in [HOMESERVER.md](./HOMESERVER.md): one public hostname, Caddy for HTTPS, systemd for the Go binary, and **apt Postgres** (not Docker).


| Host                     | Role                                                  |
| ------------------------ | ----------------------------------------------------- |
| `rubrical.spencerls.dev` | Landing, auth, dashboard, extension API (same origin) |


Templates live in [`deploy/homeserver/`](../deploy/homeserver/). Config outside the git checkout so auto-deploy (`git reset --hard`) never wipes it:

| File | Owns |
|------|------|
| `/etc/homeserver/server.env` | **Only** place for `RUBRICAL_HOST` / `RUBRICAL_PORT` (and hook host/port). Loaded by Caddy **and** `rubrical.service`. |
| `/etc/homeserver/apps/rubrical.env` | Secrets and app settings (`DATABASE_URL`, encryption key, OAuth, …). No listen address. |

---

## Values

```text
APP_NAME              rubrical
DOMAIN                rubrical.spencerls.dev
ZONE_NAME             spencerls.dev
REPO_PATH             /srv/repos/rubrical
APP_PORT              8787
HOOK_PORT             9011   # pick a free port if taken
SERVICE_NAME          rubrical
HOOK_SERVICE_NAME     deploy-hook-rubrical
BRANCH                main
```

Prerequisites on the server (once): Go 1.26+, pnpm (via corepack), Caddy, deploy-hook binary, Cloudflare DDNS, router 80/443 only — see HOMESERVER.md.

---

## 1. DNS

Add to `/etc/homeserver/cloudflare-ddns.records`:

```text
spencerls.dev      rubrical.spencerls.dev           true
```

```bash
sudo /usr/local/bin/cloudflare-ddns
dig +short rubrical.spencerls.dev @1.1.1.1
```

---

## 2. Clone the repo

```bash
cd /srv/repos
git clone <YOUR_GIT_URL> rubrical
# private repo: use a deploy key (HOMESERVER.md §16)
```

Install Go 1.26+ and enable pnpm (`corepack enable`) for the deploy user.

---

## 3. Postgres (apt — not Docker)

```bash
sudo apt update
sudo apt install -y postgresql

sudo -u postgres createuser -P rubrical    # set a strong password
sudo -u postgres createdb -O rubrical rubrical
```

Confirm it listens locally:

```bash
sudo ss -ltnp | grep 5432
# expect 127.0.0.1:5432 or [::1]:5432 — not 0.0.0.0
```

If needed, in `postgresql.conf` use `listen_addresses = 'localhost'` and reload Postgres.

Docker Compose Postgres stays for **local WSL only** (`make db-up`). Production uses apt.

---

## 4. Host/port (`server.env` — single source of truth)

Append [`deploy/homeserver/server.env.snippet`](../deploy/homeserver/server.env.snippet) to `/etc/homeserver/server.env`:

```env
RUBRICAL_HOST=127.0.0.1
RUBRICAL_PORT=8787
RUBRICAL_HOOK_HOST=127.0.0.1
RUBRICAL_HOOK_PORT=9011
```

Caddy reverse-proxies using these vars. The Go process listens using the same vars (`RUBRICAL_HOST` + `RUBRICAL_PORT`). Do **not** put a listen address in `apps/rubrical.env`.

Caddy must load that file (`systemctl edit caddy` → `EnvironmentFile=/etc/homeserver/server.env`). Then merge [`deploy/homeserver/Caddyfile.snippet`](../deploy/homeserver/Caddyfile.snippet) into `/etc/caddy/Caddyfile`:

```bash
sudo caddy fmt --overwrite /etc/caddy/Caddyfile
sudo caddy validate --config /etc/caddy/Caddyfile
sudo systemctl restart caddy   # restart if server.env changed; else reload
```

---

## 5. App secrets (`apps/rubrical.env`)

```bash
sudo mkdir -p /etc/homeserver/apps
sudo cp /srv/repos/rubrical/deploy/homeserver/rubrical.env.example \
  /etc/homeserver/apps/rubrical.env
sudo nano /etc/homeserver/apps/rubrical.env
# Deploy user must read this (deploy.sh sources it for migrate/build).
sudo chown root:<LINUX_USER> /etc/homeserver/apps/rubrical.env
sudo chmod 640 /etc/homeserver/apps/rubrical.env
```

Set at least:

```env
RUBRICAL_PUBLIC_URL=https://rubrical.spencerls.dev
DATABASE_URL=postgres://rubrical:<password>@127.0.0.1:5432/rubrical?sslmode=disable
RUBRICAL_DATA_DIR=/srv/repos/rubrical/data
RUBRICAL_SECRETS_ENCRYPTION_KEY=<stable 32-byte key>
```

Generate the secrets key once (`pnpm setup:secrets-key` or copy from `.env.local`) and paste it here. **Do not rotate** casually — it encrypts user AI API keys at rest.

Optional: Google OAuth, Resend/SMTP, `RUBRICAL_EXTENSION_ORIGINS`. See [configuration.md](./configuration.md).

---

## 6. systemd app service

```bash
sudo cp /srv/repos/rubrical/deploy/homeserver/rubrical.service \
  /etc/systemd/system/rubrical.service
sudo nano /etc/systemd/system/rubrical.service   # set User= to your deploy user
```

First build + migrate (as the deploy user, with env loaded):

```bash
set -a && source /etc/homeserver/apps/rubrical.env && set +a
cd /srv/repos/rubrical
pnpm install --frozen-lockfile
make css templ migrate-up build
mkdir -p data
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now rubrical
curl -I http://127.0.0.1:8787
curl -kI --resolve rubrical.spencerls.dev:443:127.0.0.1 https://rubrical.spencerls.dev/
```

Optional daily purge:

```bash
sudo cp /srv/repos/rubrical/deploy/homeserver/rubrical-purge.service \
  /etc/systemd/system/rubrical-purge.service
sudo cp /srv/repos/rubrical/deploy/homeserver/rubrical-purge.timer \
  /etc/systemd/system/rubrical-purge.timer
sudo nano /etc/systemd/system/rubrical-purge.service   # same User=
sudo systemctl daemon-reload
sudo systemctl enable --now rubrical-purge.timer
```

---

## 7. Deploy script + sudoers

The versioned script is `[deploy/homeserver/deploy.sh](../deploy/homeserver/deploy.sh)`. Point the webhook at that path (no copy under `/srv/deploy` required).

```bash
sudo visudo
# add (exact service name only):
# <LINUX_USER> ALL=(root) NOPASSWD: /usr/bin/systemctl restart rubrical.service
```

Manual deploy:

```bash
/srv/repos/rubrical/deploy/homeserver/deploy.sh
```

---

## 8. Deploy-hook + GitHub webhook

```bash
sudo cp /srv/repos/rubrical/deploy/homeserver/deploy-hook.env.example \
  /etc/homeserver/deploy-hooks/rubrical.env
sudo nano /etc/homeserver/deploy-hooks/rubrical.env   # set GITHUB_WEBHOOK_SECRET
sudo chmod 600 /etc/homeserver/deploy-hooks/rubrical.env
```

Create `/etc/systemd/system/deploy-hook-rubrical.service` from HOMESERVER.md §14 (`EnvironmentFile=/etc/homeserver/deploy-hooks/rubrical.env`).

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now deploy-hook-rubrical
curl -i http://127.0.0.1:9011/_github/rubrical
# expect 405
curl -ki --resolve rubrical.spencerls.dev:443:127.0.0.1 \
  https://rubrical.spencerls.dev/_github/rubrical
# expect 405
```

GitHub → repo → Settings → Webhooks:

```text
Payload URL:   https://rubrical.spencerls.dev/_github/rubrical
Content type:  application/json
Secret:        same as GITHUB_WEBHOOK_SECRET
Events:        Just the push event
```

Push to `main` and confirm deploy logs:

```bash
journalctl -u deploy-hook-rubrical -f
```

---

## 9. Auth / email / extension


| Item           | Action                                                                                                                       |
| -------------- | ---------------------------------------------------------------------------------------------------------------------------- |
| Google OAuth   | Authorized redirect: `https://rubrical.spencerls.dev/auth/google/callback`                                                   |
| Password reset | Set `RESEND_API_KEY` or SMTP; leave `EMAIL_DEV_LOG` unset/off                                                                |
| Extension      | `make extension-build-prod` locally; set `RUBRICAL_EXTENSION_ORIGINS=chrome-extension://…` on the server; restart `rubrical` |


---

## Day-to-day

```text
git push origin main
  → GitHub webhook
  → deploy-hook
  → deploy/homeserver/deploy.sh
  → fetch/reset, pnpm, css/templ, migrate, build, restart rubrical
```


| Check           | Command                                                                                   |
| --------------- | ----------------------------------------------------------------------------------------- |
| App local       | `curl -I http://127.0.0.1:8787`                                                           |
| HTTPS via Caddy | `curl -kI --resolve rubrical.spencerls.dev:443:127.0.0.1 https://rubrical.spencerls.dev/` |
| Public DNS      | `dig +short rubrical.spencerls.dev @1.1.1.1`                                              |
| App logs        | `journalctl -u rubrical -f`                                                               |
| Deploy logs     | `journalctl -u deploy-hook-rubrical -f`                                                   |


---

## Optional: Vercel marketing split

Not used for `rubrical.spencerls.dev`. If you ever host a **different** marketing hostname on Vercel, `make vercel-build` + `vercel.json` can proxy app paths to this Go host. Do not point the same hostname at both Vercel and Caddy.