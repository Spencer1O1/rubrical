# Deployment

Production hosting for Rubrical is the **home server** pattern in [HOMESERVER.md](./HOMESERVER.md): one public hostname, Caddy for HTTPS, systemd for the Go binary, and **apt Postgres** (not Docker).


| Host                     | Role                                                  |
| ------------------------ | ----------------------------------------------------- |
| `rubrical.spencerls.dev` | `/` dashboard (or → `/onboarding`), auth, extension API (same origin) |


Templates live in `[deploy/homeserver/](../deploy/homeserver/)`. Config outside the git checkout so auto-deploy (`git reset --hard`) never wipes it:


| File                                | Owns                                                                                                                   |
| ----------------------------------- | ---------------------------------------------------------------------------------------------------------------------- |
| `/etc/homeserver/server.env`        | **Only** place for `RUBRICAL_HOST` / `RUBRICAL_PORT` (and hook host/port). Loaded by Caddy **and** `rubrical.service`. |
| `/etc/homeserver/apps/rubrical.env` | Secrets and app settings (`DATABASE_URL`, encryption key, OAuth, …). No listen address.                                |


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

Append `[deploy/homeserver/server.env.snippet](../deploy/homeserver/server.env.snippet)` to `/etc/homeserver/server.env`:

```env
RUBRICAL_HOST=127.0.0.1
RUBRICAL_PORT=8787
RUBRICAL_HOOK_HOST=127.0.0.1
RUBRICAL_HOOK_PORT=9011
```

Caddy reverse-proxies using these vars. The Go process listens using the same vars (`RUBRICAL_HOST` + `RUBRICAL_PORT`). Do **not** put a listen address in `apps/rubrical.env`.

Caddy must load that file (`systemctl edit caddy` → `EnvironmentFile=/etc/homeserver/server.env`). Then merge `[deploy/homeserver/Caddyfile.snippet](../deploy/homeserver/Caddyfile.snippet)` into `/etc/caddy/Caddyfile`:

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
# Your login user must be able to read this (deploy.sh / first build source it).
sudo chown "root:$USER" /etc/homeserver/apps/rubrical.env
sudo chmod 640 /etc/homeserver/apps/rubrical.env
# server.env is host/port only — world-readable is fine
sudo chmod 644 /etc/homeserver/server.env
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

First build + migrate as your normal user (**not** `sudo source` — `source` is a shell builtin):

```bash
# If you still get "Permission denied", re-run the chown/chmod from §5.
set -a
source /etc/homeserver/server.env
source /etc/homeserver/apps/rubrical.env
set +a

cd /srv/repos/rubrical
# If frozen-lockfile fails, the lockfile on main is stale — fix it in the
# laptop checkout with `pnpm install`, commit pnpm-lock.yaml, push, then pull here.
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

Same layout as spencerls: runtime script under `/srv/deploy/`, not inside the git checkout (so `git reset --hard` cannot rewrite the script mid-deploy). The file in the repo is only a template.

```bash
sudo mkdir -p /srv/deploy/rubrical
sudo cp /srv/repos/rubrical/deploy/homeserver/deploy.sh /srv/deploy/rubrical/deploy.sh
sudo chown "$USER:$USER" /srv/deploy/rubrical/deploy.sh
chmod +x /srv/deploy/rubrical/deploy.sh

sudo visudo
# add (exact service name only):
# <LINUX_USER> ALL=(root) NOPASSWD: /usr/bin/systemctl restart rubrical.service
```

If you change the template later, re-copy it to `/srv/deploy/rubrical/deploy.sh`.

Manual deploy:

```bash
/srv/deploy/rubrical/deploy.sh
```

---

## 8. Deploy-hook + GitHub webhook

### Webhook secret

GitHub does not issue this — you generate it, then paste the **same** value in two places (env file + GitHub UI):

```bash
openssl rand -hex 32
```

### Hook env + systemd

Same as spencerls: `DEPLOY_HOOK_*` in the hook env file; `RUBRICAL_HOOK_*` in `server.env` for Caddy. Keep the port numbers equal.

```bash
sudo mkdir -p /etc/homeserver/deploy-hooks
sudo cp /srv/repos/rubrical/deploy/homeserver/deploy-hook.env.example \
  /etc/homeserver/deploy-hooks/rubrical.env
sudo nano /etc/homeserver/deploy-hooks/rubrical.env   # paste GITHUB_WEBHOOK_SECRET
sudo chmod 600 /etc/homeserver/deploy-hooks/rubrical.env

sudo cp /srv/repos/rubrical/deploy/homeserver/deploy-hook-rubrical.service \
  /etc/systemd/system/deploy-hook-rubrical.service
sudo nano /etc/systemd/system/deploy-hook-rubrical.service   # set User=
sudo systemctl daemon-reload
sudo systemctl enable --now deploy-hook-rubrical

curl -i http://127.0.0.1:9011/_github/rubrical
# expect 405
curl -ki --resolve rubrical.spencerls.dev:443:127.0.0.1 \
  https://rubrical.spencerls.dev/_github/rubrical
# expect 405
```

GitHub → repo → Settings → Webhooks → Add webhook:

```text
Payload URL:   https://rubrical.spencerls.dev/_github/rubrical
Content type:  application/json
Secret:        the openssl value (same as GITHUB_WEBHOOK_SECRET)
Events:        Just the push event
Active:        checked
```

Push to `main` and confirm deploy logs:

```bash
journalctl -u deploy-hook-rubrical -f
```

---

## 9. Auth / email / extension

These are optional for a first bring-up (email/password signup works without Google). Do them when you want Google sign-in, real password-reset mail, and the hosted extension zip.

### Extension (hosted at `/install`)

Deploy already runs `make extension-package`, which builds the prod extension and writes `static/downloads/rubrical-extension.zip`. Users open [https://rubrical.spencerls.dev/install](https://rubrical.spencerls.dev/install).

In `/etc/homeserver/apps/rubrical.env` set the stable origin (from the public `key` in `extension/manifest.json`):

```env
RUBRICAL_EXTENSION_ORIGINS=chrome-extension://mdjogmaimihfjhgobpajfpkbibgmecce
```

```bash
sudo systemctl restart rubrical
```

Confirm:

```bash
curl -I https://rubrical.spencerls.dev/install
curl -I https://rubrical.spencerls.dev/install/rubrical-extension.zip
# expect: cache-control: no-store… and cf-cache-status: DYNAMIC or BYPASS (not HIT on a stale zip)
```

If an old zip was already cached at the Cloudflare edge, purge that URL once in the Cloudflare dashboard (Caching → Configuration → Purge Cache), or rely on the new `/install/rubrical-extension.zip?v=…` link after this deploy.

### Google OAuth (optional)

1. [Google Cloud Console](https://console.cloud.google.com/) → APIs & Services → Credentials → Create OAuth client ID (Web application).
2. Authorized JavaScript origins: `https://rubrical.spencerls.dev`
3. Authorized redirect URI: `https://rubrical.spencerls.dev/auth/google/callback`
4. Put client id/secret in `/etc/homeserver/apps/rubrical.env`:

```env
GOOGLE_OAUTH_CLIENT_ID=....apps.googleusercontent.com
GOOGLE_OAUTH_CLIENT_SECRET=...
```

5. `sudo systemctl restart rubrical` — “Continue with Google” appears on `/login`.

### Password-reset email (optional)

Without this, forgot-password won’t send real mail. Prefer [Resend](https://resend.com/):

```env
EMAIL_FROM="Rubrical <noreply@rubrical.spencerls.dev>"
RESEND_API_KEY=re_...
# do not set EMAIL_DEV_LOG in production
```

Or SMTP (`SMTP_HOST`, `SMTP_PORT`, `SMTP_USERNAME`, `SMTP_PASSWORD`). Restart `rubrical` after editing.

---

## Day-to-day

```text
git push origin main
  → GitHub webhook
  → deploy-hook
  → /srv/deploy/rubrical/deploy.sh
  → fetch/reset, pnpm, css/templ, migrate, build, restart rubrical
```


| Check           | Command                                                                                   |
| --------------- | ----------------------------------------------------------------------------------------- |
| App local       | `curl -I http://127.0.0.1:8787`                                                           |
| HTTPS via Caddy | `curl -kI --resolve rubrical.spencerls.dev:443:127.0.0.1 https://rubrical.spencerls.dev/` |
| Public DNS      | `dig +short rubrical.spencerls.dev @1.1.1.1`                                              |
| App logs        | `journalctl -u rubrical -f`                                                               |
| Deploy logs     | `journalctl -u deploy-hook-rubrical -f`                                                   |
