# Home Server Hosting

This document describes the generic pattern for hosting public apps, domains, subdomains, GitHub auto-deploy webhooks, dynamic DNS, and local services on a self-hosted Ubuntu server.

Recommended location:

```text
/srv/docs/HOMESERVER.md
```

---

# 1. Core mental model

Every public app follows this pattern:

```text
Internet
  ↓
Cloudflare DNS
  ↓
current home WAN/public IP
  ↓
router port forwarding: 80/443 only
  ↓
Ubuntu server LAN IP
  ↓
Caddy
  ↓
local app service on 127.0.0.1:<APP_PORT>
```

Every auto-deployed repo follows this pattern:

```text
git push to GitHub
  ↓
GitHub webhook
  ↓
https://<DOMAIN>/_github/<APP_NAME>
  ↓
Caddy
  ↓
local deploy-hook service on 127.0.0.1:<HOOK_PORT>
  ↓
deploy script
  ↓
git fetch/reset
  ↓
install/build
  ↓
restart app service
```

Dynamic DNS follows this pattern:

```text
systemd timer
  ↓
cloudflare-ddns script
  ↓
detect current WAN IP
  ↓
update Cloudflare A records
  ↓
domains keep pointing at the home server
```

Important rules:

```text
Expose only ports 80 and 443 to the internet.
Do not expose app ports like 3000, 8080, or 9001.
Do not expose deploy-hook ports.
Do not expose SSH publicly.
Use Tailscale or LAN SSH for administration.
Use Caddy as the public HTTPS entrypoint.
Use systemd for long-running services.
Use deploy scripts for repeatable releases.
Use Cloudflare DNS + DDNS for domains that point at the home WAN IP.
Do not hardcode the WAN IP in app configs, Caddy configs, systemd services, GitHub webhooks, laptop hosts files, or WSL hosts files.
```

---

# 2. Placeholder glossary

Use these placeholders throughout this runbook:

```text
<PUBLIC_IP>          The current router/WAN/public IP address
<SERVER_LAN_IP>      The server's LAN IP, e.g. 192.168.1.212
<SERVER_TS_IP>       The server's Tailscale IP, e.g. 100.x.y.z
<LINUX_USER>         The Ubuntu deploy/admin user
<DOMAIN>             A root domain, e.g. example.dev
<SUBDOMAIN>          A subdomain, e.g. app.example.dev
<ZONE_NAME>          Cloudflare DNS zone, usually the root domain, e.g. example.dev
<APP_NAME>           Short service name, e.g. portfolio, workforce, cheque
<REPO_NAME>          GitHub repo name, e.g. portfolio-site
<REPO_PATH>          Local checkout path, e.g. /srv/repos/portfolio-site
<APP_PORT>           Local app port, e.g. 3000
<HOOK_PORT>          Local webhook port, e.g. 9001
<PACKAGE_FILTER>     pnpm/turbo package filter, e.g. portfolio
<APP_ENV_PREFIX>     Uppercase env prefix, e.g. PORTFOLIO
<SERVICE_NAME>       systemd app service, e.g. portfolio-web
<HOOK_SERVICE_NAME>  systemd webhook service, e.g. deploy-hook-portfolio
<BRANCH>             Deploy branch, usually main
```

---

# 3. Standard server layout

Use this layout for predictable operations:

```text
/srv/docs/
  HOMESERVER.md

/srv/repos/
  <REPO_NAME>/

/srv/deploy/
  <APP_NAME>/
    deploy.sh

/opt/deploy-hook/
  main.go
  bin/
    deploy-hook

/usr/local/bin/
  cloudflare-ddns

/etc/homeserver/
  server.env
  cloudflare-ddns.env
  cloudflare-ddns.records
  minio.env
  apps/
    <APP_NAME>.env
  deploy-hooks/
    <APP_NAME>.env

/var/lib/minio/

/etc/caddy/
  Caddyfile

/etc/systemd/system/
  <SERVICE_NAME>.service
  <HOOK_SERVICE_NAME>.service
  cloudflare-ddns.service
  cloudflare-ddns.timer
```

Recommended ownership for app files:

```bash
sudo mkdir -p /srv/repos /srv/deploy /srv/docs
sudo chown -R <LINUX_USER>:<LINUX_USER> /srv/repos /srv/deploy /srv/docs
```

Recommended ownership for secrets:

```bash
sudo mkdir -p /etc/homeserver/apps /etc/homeserver/deploy-hooks
sudo chown -R root:root /etc/homeserver
sudo chmod 755 /etc/homeserver /etc/homeserver/apps /etc/homeserver/deploy-hooks
sudo chmod 600 /etc/homeserver/*.env
sudo chmod 600 /etc/homeserver/deploy-hooks/*.env
# After creating minio.env (section 8), keep it mode 640 or 600 like other secrets.
```

The records file does not contain secrets, but it is still server configuration:

```bash
sudo chown root:root /etc/homeserver/cloudflare-ddns.records
sudo chmod 644 /etc/homeserver/cloudflare-ddns.records
```

---

# 4. Network and router setup

## Router reserved IP

Reserve a stable LAN IP for the server in the router.

Example:

```text
Server -> <SERVER_LAN_IP>
```

Do this in the router UI using a setting like:

```text
Reserved IP
DHCP reservation
Static lease
Address reservation
```

Prefer router-side DHCP reservation over manually hardcoding static IPs in Ubuntu.

## Router port forwarding

Forward only web ports:

```text
TCP 80  -> <SERVER_LAN_IP>:80
TCP 443 -> <SERVER_LAN_IP>:443
```

Do not forward these:

```text
22
3000
3001
8080
9001
any app-specific port
any webhook-specific port
```

Do not enable DMZ.

---

# 5. DNS setup

DNS should be hosted through Cloudflare for domains that point to this home server.

The domain registration can stay wherever the domain was purchased.

```text
Registrar: domain purchase/renewal only
Cloudflare: authoritative DNS provider
DDNS updater: keeps Cloudflare A records pointed at the current WAN IP
```

Do not transfer domain registration just to use Cloudflare DNS.

## Cloudflare migration flow

For each root domain:

```text
1. Add/onboard the domain in Cloudflare.
2. Let Cloudflare import existing DNS records.
3. Review imported DNS records.
4. Preserve mail, verification, and required TXT/CNAME/MX records.
5. Keep self-hosted app A records as DNS only at first.
6. Replace nameservers at the current registrar with Cloudflare nameservers.
7. Wait for Cloudflare activation.
8. Add hosted records to /etc/homeserver/cloudflare-ddns.records.
9. Run the DDNS updater manually.
10. Let the systemd timer keep records updated automatically.
```

## DNS record pattern

For a root domain:

```text
A     @      <PUBLIC_IP>      DNS only
A     www    <PUBLIC_IP>      DNS only
```

For a subdomain:

```text
A     <SUBDOMAIN_PREFIX>     <PUBLIC_IP>      DNS only
```

Example:

```text
A     app       <PUBLIC_IP>      DNS only
A     api       <PUBLIC_IP>      DNS only
A     cheque    <PUBLIC_IP>      DNS only
```

The `<PUBLIC_IP>` value should be maintained by the DDNS updater, not manually hardcoded in server configs.

## DNS only vs proxied

Start with app records set to:

```text
DNS only
```

Reason:

```text
Caddy already handles HTTPS on the server.
DNS only changes DNS without changing the traffic/proxy model.
```

Cloudflare proxy mode can be enabled later after DNS, DDNS, Caddy, and webhooks are confirmed working.

Cloudflare proxied mode changes the path from:

```text
Browser -> home WAN IP -> Caddy -> app
```

to:

```text
Browser -> Cloudflare -> home WAN IP -> Caddy -> app
```

That can be useful, but it is a separate change. Do not debug DNS migration and proxy migration at the same time.

## Cloudflare proxy verification

When records are proxied, public DNS should return Cloudflare IPs instead of the home WAN IP.

Check:

```bash
dig +short <DOMAIN> @1.1.1.1
curl -I https://<DOMAIN>
```

Expected:

```text
DNS returns Cloudflare IPs, not the home WAN IP.
HTTP response includes server: cloudflare.
HTTP response may also include Via: 1.1 Caddy.
```

## Domain naming for proxied app services

When Cloudflare proxy is enabled, prefer one-level subdomains under the root zone.

Good:

```text
workforce.example.dev
workforce-api.example.dev
admin.example.dev
api.example.dev
```

Avoid deeper nested service domains unless the certificate setup explicitly supports them:

```text
api.workforce.example.dev
admin.internal.example.dev
```

For a product with both a web dashboard and an API, prefer:

```text
<Product>.example.dev        -> web dashboard
<Product>-api.example.dev    -> API server
```

## Do not use registrar forwarding

Do not use domain forwarding/redirects for normal app hosting.

Use DNS records only.

Redirects like `www -> apex` should be handled in Caddy, not DNS forwarding.

---

# 6. Cloudflare DDNS

The home WAN/public IP is not assumed to be permanent.

The DDNS updater keeps Cloudflare A records pointed at the current WAN IP.

The WAN IP should only live in:

```text
Cloudflare DNS records
```

It should not be hardcoded in:

```text
Caddy config
systemd services
app configs
GitHub webhook URLs
laptop hosts files
WSL hosts files
server runbook examples except as <PUBLIC_IP>
```

## DDNS files

Script:

```text
/usr/local/bin/cloudflare-ddns
```

Secret/env file:

```text
/etc/homeserver/cloudflare-ddns.env
```

Records file:

```text
/etc/homeserver/cloudflare-ddns.records
```

systemd service:

```text
/etc/systemd/system/cloudflare-ddns.service
```

systemd timer:

```text
/etc/systemd/system/cloudflare-ddns.timer
```

## DDNS env file

Create/edit:

```bash
sudo nano /etc/homeserver/cloudflare-ddns.env
```

Expected contents:

```env
CLOUDFLARE_API_TOKEN=REDACTED
CLOUDFLARE_DDNS_RECORDS_FILE=/etc/homeserver/cloudflare-ddns.records
```

Secure:

```bash
sudo chown root:root /etc/homeserver/cloudflare-ddns.env
sudo chmod 600 /etc/homeserver/cloudflare-ddns.env
```

## Cloudflare API token

The API token should have DNS edit access to every Cloudflare zone listed in the records file.

Required permissions:

```text
Zone -> Zone -> Read
Zone -> DNS  -> Edit
```

Zone scope options:

```text
Specific zone -> <ZONE_NAME>
```

or, if managing multiple domains:

```text
All zones in this account
```

Use the smallest scope that still covers all domains this server needs to update.

Do not use the global API key.

Do not commit the token.

Do not paste the token into chat, docs, repos, or logs.

## DDNS records file

Create/edit:

```bash
sudo nano /etc/homeserver/cloudflare-ddns.records
```

Format:

```text
# zone_name      record_name                  proxied
example.dev      example.dev                  false
example.dev      www.example.dev              false
example.dev      app.example.dev              false
another.com      another.com                  false
another.com      www.another.com              false
```

Fields:

```text
zone_name     Cloudflare zone/root domain
record_name   full DNS record name
proxied       true or false
```

Use `false` at first.

Use `true` only after intentionally enabling Cloudflare proxy mode for that record.

## Adding a new domain to DDNS

For a totally new root domain:

```text
1. Add/onboard the domain in Cloudflare.
2. Change the registrar nameservers to Cloudflare nameservers.
3. Wait for Cloudflare activation.
4. Add A records in Cloudflare or let the updater create them.
5. Add the records to /etc/homeserver/cloudflare-ddns.records.
6. Make sure the API token can access that zone.
7. Run sudo /usr/local/bin/cloudflare-ddns.
8. Verify with dig.
```

Example records:

```text
newdomain.dev      newdomain.dev              false
newdomain.dev      www.newdomain.dev          false
newdomain.dev      api.newdomain.dev          false
```

## Adding a new subdomain to DDNS

For a new subdomain under an existing Cloudflare zone, add one line:

```text
example.dev      newapp.example.dev           false
```

Then run:

```bash
sudo /usr/local/bin/cloudflare-ddns
```

## Manual DDNS run

Run:

```bash
sudo /usr/local/bin/cloudflare-ddns
```

Expected output shape:

```text
Detected WAN IP: <PUBLIC_IP>

Checking A record: example.dev in zone example.dev
Already current: example.dev -> <PUBLIC_IP> proxied=false

Checking A record: www.example.dev in zone example.dev
Already current: www.example.dev -> <PUBLIC_IP> proxied=false

Cloudflare DDNS update complete.
```

If a record does not exist, the updater should create it.

If the IP changed, the updater should update it.

## Verify DNS

Check Cloudflare DNS resolution:

```bash
dig +short <DOMAIN> @1.1.1.1
dig +short www.<DOMAIN> @1.1.1.1
dig +short <SUBDOMAIN> @1.1.1.1
```

Expected for DNS-only records:

```text
<PUBLIC_IP>
```

Expected for proxied records:

```text
Cloudflare IPs, not <PUBLIC_IP>
```

## DDNS systemd service

Service file:

```text
/etc/systemd/system/cloudflare-ddns.service
```

Expected contents:

```ini
[Unit]
Description=Update Cloudflare DNS records with current WAN IP
Wants=network-online.target
After=network-online.target

[Service]
Type=oneshot
ExecStart=/usr/local/bin/cloudflare-ddns
```

## DDNS systemd timer

Timer file:

```text
/etc/systemd/system/cloudflare-ddns.timer
```

Expected contents:

```ini
[Unit]
Description=Run Cloudflare DDNS updater periodically

[Timer]
OnBootSec=2min
OnUnitActiveSec=10min
RandomizedDelaySec=30s
Persistent=true
Unit=cloudflare-ddns.service

[Install]
WantedBy=timers.target
```

This means:

```text
Run 2 minutes after boot.
Then run about every 10 minutes.
If the server was off during a scheduled run, run after it comes back.
Add a small random delay.
```

A 10-minute interval means that if the WAN IP changes, DNS might point to the old IP until the next updater run, plus normal DNS/client caching. That is normal for a home DDNS setup.

If faster recovery is needed, change:

```ini
OnUnitActiveSec=10min
```

to:

```ini
OnUnitActiveSec=5min
```

or:

```ini
OnUnitActiveSec=2min
```

Do not run it every few seconds.

## Enable DDNS timer

Run:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now cloudflare-ddns.timer
```

Run once manually through systemd:

```bash
sudo systemctl start cloudflare-ddns.service
```

Check result:

```bash
systemctl status cloudflare-ddns.service --no-pager
journalctl -u cloudflare-ddns.service -n 100 --no-pager
```

Check timer:

```bash
systemctl list-timers cloudflare-ddns.timer
```

---

# 7. Caddy setup

Caddy should be installed directly on Ubuntu.

The main config is:

```text
/etc/caddy/Caddyfile
```

Caddy should reverse-proxy public domains to local services.

## Shared environment file

Create a shared env file:

```bash
sudo mkdir -p /etc/homeserver
sudo nano /etc/homeserver/server.env
```

Example:

```env
# App ports
SPENCERLS_WEB_HOST=127.0.0.1
SPENCERLS_WEB_PORT=3000

WORKFORCE_WEB_HOST=127.0.0.1
WORKFORCE_WEB_PORT=3001

WORKFORCE_API_HOST=127.0.0.1
WORKFORCE_API_PORT=8081

# Webhook routing ports
SPENCERLS_HOOK_HOST=127.0.0.1
SPENCERLS_HOOK_PORT=9001

WORKFORCE_HOOK_HOST=127.0.0.1
WORKFORCE_HOOK_PORT=9002

# Shared Postgres (host/port only; credentials stay in apps/*.env)
POSTGRES_HOST=127.0.0.1
POSTGRES_PORT=5432

# Shared MinIO (host/port only; root creds in minio.env, app keys in apps/*.env)
MINIO_HOST=127.0.0.1
MINIO_PORT=9100
```

## Make Caddy load the shared env file

Run:

```bash
sudo systemctl edit caddy
```

Paste:

```ini
[Service]
EnvironmentFile=/etc/homeserver/server.env
```

Then:

```bash
sudo systemctl daemon-reload
sudo systemctl restart caddy
```

Verify:

```bash
sudo systemctl cat caddy
sudo systemctl show caddy --property=EnvironmentFiles
```

Expected:

```text
EnvironmentFiles=/etc/homeserver/server.env
```

## Example Caddyfile for multiple apps

```caddyfile
www.<DOMAIN> {
	redir https://<DOMAIN>{uri} permanent
}

<DOMAIN> {
	@deployHook path /_github/<APP_NAME> /_github/<APP_NAME>/*
	handle @deployHook {
		reverse_proxy {$SPENCERLS_HOOK_HOST}:{$SPENCERLS_HOOK_PORT}
	}

	handle {
		reverse_proxy {$SPENCERLS_WEB_HOST}:{$SPENCERLS_WEB_PORT}
	}
}

workforce.<DOMAIN> {
	reverse_proxy {$WORKFORCE_WEB_HOST}:{$WORKFORCE_WEB_PORT}
}

workforce-api.<DOMAIN> {
	reverse_proxy {$WORKFORCE_API_HOST}:{$WORKFORCE_API_PORT}
}
```

For a simple app without a webhook route:

```caddyfile
app.<DOMAIN> {
	reverse_proxy 127.0.0.1:<APP_PORT>
}
```

## Caddy commands

Format:

```bash
sudo caddy fmt --overwrite /etc/caddy/Caddyfile
```

Validate:

```bash
sudo caddy validate --config /etc/caddy/Caddyfile
```

Reload:

```bash
sudo systemctl reload caddy
```

Restart:

```bash
sudo systemctl restart caddy
```

Logs:

```bash
journalctl -u caddy -f
```

Test a domain locally through Caddy:

```bash
curl -kI --resolve <DOMAIN>:443:127.0.0.1 https://<DOMAIN>/
```

Expected:

```text
HTTP/2 200
```

---

# 8. Shared data services (Postgres + MinIO)

## Env naming (all apps)

This is a **homeserver** contract — every product follows it (WorkForce, Rubrical, later apps).

| Kind               | Names                                                                       | Where                                                                            |
| ------------------ | --------------------------------------------------------------------------- | -------------------------------------------------------------------------------- |
| Product settings   | `<PRODUCT>_…` (e.g. `WORKFORCE_*`, `RUBRICAL_*`)                            | `server.env` for listen host/port; `apps/<app>.env` for product secrets/settings |
| Shared infra       | `POSTGRES_*`, `MINIO_*`, `STORAGE_*` — **no** product prefix                | Host/port in `server.env`; per-app role/keys in `apps/<app>.env`                 |
| Vendor / framework | Their required names (`STRIPE_*`, `NEXT_PUBLIC_*`, `GOOGLE_*`, `SMTP_*`, …) | Usually `apps/<app>.env`                                                         |

Examples that are correct: `POSTGRES_USER`, `STORAGE_ACCESS_KEY`, `WORKFORCE_PUBLIC_WEB_URL`, `RUBRICAL_DATA_DIR`.  
Wrong: `WORKFORCE_POSTGRES_USER`, `WORKFORCE_STORAGE_*`, `RUBRICAL_POSTGRES_*`, baking `POSTGRES_HOST` into app env.

`STORAGE_*` is the per-app MinIO user/bucket (same idea as `POSTGRES_USER` / `POSTGRES_DB`). Listen address is only `MINIO_HOST` / `MINIO_PORT` / `MINIO_CONSOLE_PORT`.

## Postgres host/port

One apt Postgres for the server. **Host and port belong in `server.env`**, not hardcoded inside each app’s secrets file:

```env
POSTGRES_HOST=127.0.0.1
POSTGRES_PORT=5432
```

Per-app `/etc/homeserver/apps/<APP_NAME>.env` holds only the role secrets and database name, for example:

```env
POSTGRES_USER=<app>
POSTGRES_PASSWORD=CHANGE_ME
POSTGRES_DB=<app>
POSTGRES_SSLMODE=disable
```

Apps assemble the connection string from `POSTGRES_HOST` / `POSTGRES_PORT` (server.env) plus user/password/db (app env). Never bake host/port into the app env.

Each app still gets its own Postgres **role + database** on that shared instance.

## Object storage (MinIO)

One MinIO instance for the whole server — same idea as one Postgres.

Apps do **not** each run their own MinIO. Each app gets:

- its own **MinIO user** (access key + secret) — same idea as a Postgres role
- its own **bucket**
- those credentials in `/etc/homeserver/apps/<APP_NAME>.env` as `STORAGE_ACCESS_KEY`, `STORAGE_SECRET_KEY`, `STORAGE_BUCKET`, `STORAGE_PUBLIC_ENDPOINT` (optional `STORAGE_ENDPOINT` override)

`MINIO_ROOT_*` in `minio.env` is admin-only (create users/buckets). Do **not** put root into app env.

One public hostname for browser/presigned access (all apps):

```text
Internet
  ↓
Caddy → storage.<ZONE_NAME>
  ↓
127.0.0.1:9100  →  minio.service  →  /var/lib/minio
```

Do not expose `9100` / `9101` on the WAN. Console stays on loopback (`127.0.0.1:9101`).

## Install binaries (once)

Server (`minio`) plus optional admin CLI (`mc`) — same pattern as installing `psql` tools for Postgres:

```bash
curl -fsSL https://dl.min.io/server/minio/release/linux-amd64/minio -o /tmp/minio
chmod +x /tmp/minio
sudo mv /tmp/minio /usr/local/bin/minio

curl -fsSL https://dl.min.io/client/mc/release/linux-amd64/mc -o /tmp/mc
chmod +x /tmp/mc
sudo mv /tmp/mc /usr/local/bin/mc

minio --version
mc --version
```

`mc` is only for operator tasks (create users/buckets). Runtime apps talk to MinIO over the S3 API and do not need `mc`.

## Data directory

```bash
sudo mkdir -p /var/lib/minio
sudo chown <LINUX_USER>:<LINUX_USER> /var/lib/minio
```

## Env file

```bash
sudo nano /etc/homeserver/minio.env
sudo chown root:<LINUX_USER> /etc/homeserver/minio.env
sudo chmod 640 /etc/homeserver/minio.env
```

```env
MINIO_ROOT_USER=minio
MINIO_ROOT_PASSWORD=CHANGE_ME
MINIO_VOLUMES=/var/lib/minio
```

`MINIO_ROOT_*` are server admin credentials only. Create a per-app MinIO user and put that access key/secret in the app env (never the root pair). Do **not** put listen addresses in `minio.env` — those live only in `server.env`.

## server.env host/port (single source of truth)

Same pattern as Postgres. Caddy, apps, and `minio.service` all read these — there is no second bind address in `minio.env`:

```env
MINIO_HOST=127.0.0.1
MINIO_PORT=9100
MINIO_CONSOLE_PORT=9101
```

## systemd unit

```bash
sudo nano /etc/systemd/system/minio.service
```

```ini
[Unit]
Description=MinIO object storage
Documentation=https://docs.min.io
Wants=network-online.target
After=network-online.target
AssertFileIsExecutable=/usr/local/bin/minio

[Service]
Type=simple
User=<LINUX_USER>
Group=<LINUX_USER>
ProtectProc=invisible
EnvironmentFile=/etc/homeserver/server.env
EnvironmentFile=/etc/homeserver/minio.env
ExecStartPre=/bin/bash -c "test -n \"${MINIO_HOST}\" && test -n \"${MINIO_PORT}\" && test -n \"${MINIO_CONSOLE_PORT}\" && test -n \"${MINIO_VOLUMES}\" || { echo 'MINIO_HOST/PORT/CONSOLE_PORT (server.env) and MINIO_VOLUMES (minio.env) required' >&2; exit 1; }"
ExecStart=/usr/local/bin/minio server --address ${MINIO_HOST}:${MINIO_PORT} --console-address ${MINIO_HOST}:${MINIO_CONSOLE_PORT} ${MINIO_VOLUMES}
Restart=always
RestartSec=5
LimitNOFILE=65536
TasksMax=infinity
TimeoutStopSec=infinity
SendSIGKILL=no

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now minio
curl -s -o /dev/null -w '%{http_code}\n' http://127.0.0.1:9100/minio/health/live
# expect 200
```

## Public Caddy hostname (once)

Add DNS for `storage.<ZONE_NAME>` and one Caddy site:

```caddy
storage.example.dev {
	reverse_proxy {$MINIO_HOST}:{$MINIO_PORT}
}
```

Apps set their storage **public** endpoint to `https://storage.<ZONE_NAME>` and use separate buckets. Do not invent `app-storage.` hostnames per product.

Store the app’s MinIO access key/secret/bucket in `/etc/homeserver/apps/<APP_NAME>.env` first (generate the secret with `openssl rand -base64 32`). That file is the single source of truth — `mc` and the app both load it.

```bash
# example for WorkForce — STORAGE_* already in apps/workforce.env
set -a
source /etc/homeserver/server.env
source /etc/homeserver/minio.env
source /etc/homeserver/apps/workforce.env
set +a

mc alias set local "http://${MINIO_HOST}:${MINIO_PORT}" "$MINIO_ROOT_USER" "$MINIO_ROOT_PASSWORD"
mc admin info local
mc admin user add local "$STORAGE_ACCESS_KEY" "$STORAGE_SECRET_KEY"
mc mb -p "local/${STORAGE_BUCKET}"

# Bucket-scoped policy only — policy name = bucket name. Never attach built-in `readwrite`.
cat >"/tmp/${STORAGE_BUCKET}.json" <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["s3:*"],
      "Resource": [
        "arn:aws:s3:::${STORAGE_BUCKET}",
        "arn:aws:s3:::${STORAGE_BUCKET}/*"
      ]
    }
  ]
}
EOF
mc admin policy create local "$STORAGE_BUCKET" "/tmp/${STORAGE_BUCKET}.json"
mc admin policy attach local "$STORAGE_BUCKET" --user "$STORAGE_ACCESS_KEY"
```

The app may also ensure its bucket exists on boot.

## Adding another app that needs storage

1. Put public URL / access key / secret / bucket in that app’s env (generate secret once into the file).
2. Source envs; create user + bucket + **bucket-scoped** policy (above). Never `readwrite`.
3. Do **not** install another MinIO or another storage hostname.

---

# 9. App service pattern

Every long-running app should have a systemd service.

The app should listen on:

```text
127.0.0.1:<APP_PORT>
```

not on a public interface.

Caddy is responsible for public HTTPS.

## Products with multiple hosted services

A product may have more than one hosted service.

Example:

```text
workforce.example.dev
  -> web dashboard
  -> Next.js
  -> workforce-web.service
  -> 127.0.0.1:3001

workforce-api.example.dev
  -> API server
  -> Go
  -> workforce-api.service
  -> 127.0.0.1:8081
```

The mobile app should call the API domain, not the dashboard domain.

Each hosted service gets its own local port, HOST/PORT env vars, Caddy route, systemd service, and test command.

A single repo deploy script may restart multiple services.

---

# 10. Next.js app service template

Use this for a normal self-hosted Next.js app.

Important: Caddy does not serve `.next` directly.

Correct model:

```text
pnpm build
pnpm start
Caddy reverse-proxies to the running Next server
```

Create:

```bash
sudo nano /etc/systemd/system/<SERVICE_NAME>.service
```

Template:

```ini
[Unit]
Description=<DOMAIN> Next.js app
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=<LINUX_USER>
WorkingDirectory=<REPO_PATH>
Environment=NODE_ENV=production
EnvironmentFile=/etc/homeserver/server.env
ExecStart=/usr/bin/bash -lc 'exec pnpm --filter=<PACKAGE_FILTER> exec next start -H "$<APP_ENV_PREFIX>_HOST" -p "$<APP_ENV_PREFIX>_PORT"'
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

Example env prefix:

```text
PORTFOLIO
```

Then the service command uses:

```text
$PORTFOLIO_HOST
$PORTFOLIO_PORT
```

Enable/start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now <SERVICE_NAME>
```

Restart:

```bash
sudo systemctl restart <SERVICE_NAME>
```

Logs:

```bash
journalctl -u <SERVICE_NAME> -f
```

Test:

```bash
curl -I http://127.0.0.1:<APP_PORT>
```

---

# 11. Go app service template

For a Go app, build a binary and run it through systemd.

Example repo layout:

```text
<REPO_PATH>/
  cmd/server/main.go
  bin/<APP_NAME>
```

Build:

```bash
cd <REPO_PATH>
go build -o bin/<APP_NAME> ./cmd/server
```

Create:

```bash
sudo nano /etc/systemd/system/<SERVICE_NAME>.service
```

Template:

```ini
[Unit]
Description=<DOMAIN> Go app
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=<LINUX_USER>
WorkingDirectory=<REPO_PATH>
EnvironmentFile=/etc/homeserver/server.env
ExecStart=<REPO_PATH>/bin/<APP_NAME>
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

The Go app should read host/port env vars, for example:

```text
<APP_ENV_PREFIX>_HOST
<APP_ENV_PREFIX>_PORT
```

and listen on:

```text
127.0.0.1:<APP_PORT>
```

Enable/start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now <SERVICE_NAME>
```

Test:

```bash
curl -I http://127.0.0.1:<APP_PORT>
```

---

# 12. Static site service pattern

For a truly static site, no app service is needed.

Use Caddy file server.

Example directory:

```text
/srv/sites/<APP_NAME>/
  index.html
  assets/
```

Caddy:

```caddyfile
<DOMAIN> {
	root * /srv/sites/<APP_NAME>
	try_files {path} /index.html
	file_server
}
```

For React/Vite SPAs, `try_files {path} /index.html` prevents refresh 404s.

This static pattern is not the same as a normal Next.js app.

Use it only for static builds.

---

# 13. GitHub deploy automation pattern

There are three pieces:

```text
1. deploy script
2. deploy-hook service
3. GitHub webhook
```

## Deploy script

Path:

```text
/srv/deploy/<APP_NAME>/deploy.sh
```

Create:

```bash
sudo mkdir -p /srv/deploy/<APP_NAME>
sudo chown -R <LINUX_USER>:<LINUX_USER> /srv/deploy/<APP_NAME>
nano /srv/deploy/<APP_NAME>/deploy.sh
```

Generic pnpm/Next.js deploy script:

```bash
#!/usr/bin/env bash
set -euo pipefail

REPO="<REPO_PATH>"
BRANCH="<BRANCH>"
LOCK_FILE="/tmp/<APP_NAME>-deploy.lock"

(
  flock -n 9 || {
    echo "Deploy already running; exiting."
    exit 0
  }

  echo "=== Deploy started at $(date) ==="

  cd "$REPO"

  git fetch origin "$BRANCH"
  git reset --hard "origin/$BRANCH"

  corepack enable || true
  pnpm install --frozen-lockfile

  pnpm build:<APP_NAME>

  sudo /usr/bin/systemctl restart <SERVICE_NAME>.service

  echo "=== Deploy finished at $(date) ==="
) 9>"$LOCK_FILE"
```

Alternative if there is no app-specific build script:

```bash
pnpm --filter=<PACKAGE_FILTER> build
```

Make executable:

```bash
chmod +x /srv/deploy/<APP_NAME>/deploy.sh
```

Manual test:

```bash
/srv/deploy/<APP_NAME>/deploy.sh
```

## sudoers permission

The deploy script should only be allowed to restart the exact service it owns.

Edit:

```bash
sudo visudo
```

Add:

```sudoers
<LINUX_USER> ALL=(root) NOPASSWD: /usr/bin/systemctl restart <SERVICE_NAME>.service
```

Do not give broad passwordless sudo.

---

# 14. Deploy-hook generic contract

The reusable deploy-hook binary should support these env vars:

```env
WEBHOOK_PATH=/_github/<APP_NAME>
GITHUB_WEBHOOK_SECRET=...
DEPLOY_SCRIPT=/srv/deploy/<APP_NAME>/deploy.sh
DEPLOY_HOOK_HOST=127.0.0.1
DEPLOY_HOOK_PORT=<HOOK_PORT>
DEPLOY_BRANCH=main
```

Recommended model:

```text
One reusable deploy-hook binary.
One systemd deploy-hook service per repo/app.
One env file per repo/app.
One Caddy webhook route per repo/app.
One GitHub webhook per repo/app.
```

Name deploy-hook services per app/repo:

```text
deploy-hook-<APP_NAME>.service
```

Examples:

```text
deploy-hook-spencerls.service
deploy-hook-workforce.service
```

Avoid generic names like `deploy-hook.service` once the server hosts more than one repo.

Example env file:

```bash
sudo nano /etc/homeserver/deploy-hooks/<APP_NAME>.env
```

Contents:

```env
WEBHOOK_PATH=/_github/<APP_NAME>
GITHUB_WEBHOOK_SECRET=REDACTED
DEPLOY_SCRIPT=/srv/deploy/<APP_NAME>/deploy.sh
DEPLOY_HOOK_HOST=127.0.0.1
DEPLOY_HOOK_PORT=<HOOK_PORT>
DEPLOY_BRANCH=main
```

Secure:

```bash
sudo chown root:root /etc/homeserver/deploy-hooks/<APP_NAME>.env
sudo chmod 600 /etc/homeserver/deploy-hooks/<APP_NAME>.env
```

---

# 15. Deploy-hook service template

Create:

```bash
sudo nano /etc/systemd/system/deploy-hook-<APP_NAME>.service
```

Template:

```ini
[Unit]
Description=GitHub deploy webhook for <APP_NAME>
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=<LINUX_USER>
WorkingDirectory=/opt/deploy-hook
EnvironmentFile=/etc/homeserver/deploy-hooks/<APP_NAME>.env
Environment=PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
ExecStart=/opt/deploy-hook/bin/deploy-hook
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
```

Enable/start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now deploy-hook-<APP_NAME>
```

Logs:

```bash
journalctl -u deploy-hook-<APP_NAME> -f
```

Test direct hook route:

```bash
curl -i http://127.0.0.1:<HOOK_PORT>/_github/<APP_NAME>
```

Expected:

```text
405 Method Not Allowed
```

Test through Caddy:

```bash
curl -ki --resolve <DOMAIN>:443:127.0.0.1 https://<DOMAIN>/_github/<APP_NAME>
```

Expected:

```text
405 Method Not Allowed
```

---

# 16. GitHub webhook setup

In GitHub:

```text
Repo -> Settings -> Webhooks -> Add webhook
```

Use:

```text
Payload URL: https://<DOMAIN>/_github/<APP_NAME>
Content type: application/json
Secret: same as GITHUB_WEBHOOK_SECRET
Events: Just the push event
Active: checked
```

Expected response for real push:

```text
deploy started
```

Watch logs:

```bash
journalctl -u deploy-hook-<APP_NAME> -f
```

Expected logs:

```text
request received method=POST path=/_github/<APP_NAME> event=push
signature valid event=push
push payload ref=refs/heads/main
accepted deploy request
deploy started
=== Deploy started ...
=== Deploy finished ...
deploy finished
```

---

# 17. Cloning repos

Repos should live under:

```text
/srv/repos/
```

Public repo:

```bash
cd /srv/repos
git clone https://github.com/<OWNER>/<REPO_NAME>.git <REPO_NAME>
```

Private repo with deploy key:

```bash
ssh-keygen -t ed25519 -C "<REPO_NAME>-server-deploy" -f ~/.ssh/github_<REPO_NAME>_deploy
cat ~/.ssh/github_<REPO_NAME>_deploy.pub
```

Add public key in GitHub:

```text
Repo -> Settings -> Deploy keys -> Add deploy key
```

Usually leave write access unchecked.

Add SSH alias:

```bash
nano ~/.ssh/config
```

```sshconfig
Host github.com-<REPO_NAME>
    HostName github.com
    User git
    IdentityFile ~/.ssh/github_<REPO_NAME>_deploy
    IdentitiesOnly yes
```

Clone:

```bash
cd /srv/repos
git clone git@github.com-<REPO_NAME>:<OWNER>/<REPO_NAME>.git <REPO_NAME>
```

Test:

```bash
cd /srv/repos/<REPO_NAME>
git fetch origin
```

---

# 18. Adding a new app from an existing monorepo

Use this when the repo is already cloned and the new app is another package/app inside it.

Example values:

```text
<APP_NAME> = cheque
<DOMAIN> = cheque.example.dev
<ZONE_NAME> = example.dev
<REPO_PATH> = /srv/repos/main-monorepo
<PACKAGE_FILTER> = cheque
<APP_ENV_PREFIX> = CHEQUE
<APP_PORT> = 3002
<HOOK_PORT> = 9003
<SERVICE_NAME> = cheque-web
<HOOK_SERVICE_NAME> = deploy-hook-cheque
```

Steps:

```text
1. Add the domain/subdomain to /etc/homeserver/cloudflare-ddns.records.
2. Run sudo /usr/local/bin/cloudflare-ddns.
3. Verify DNS with dig.
4. Add HOST/PORT values to /etc/homeserver/server.env.
5. Create app systemd service.
6. Add Caddy domain block.
7. Create deploy script or update existing monorepo deploy script.
8. Add sudoers permission for exact service restart.
9. Create deploy-hook env file.
10. Create deploy-hook systemd service.
11. Add GitHub webhook.
12. Test direct app, Caddy route, webhook route, and push deploy.
```

Example DDNS record:

```text
example.dev      cheque.example.dev           false
```

Example env vars:

```env
CHEQUE_HOST=127.0.0.1
CHEQUE_PORT=3002

CHEQUE_HOOK_HOST=127.0.0.1
CHEQUE_HOOK_PORT=9003
```

---

# 19. Adding a new app from a totally different repo/domain

Example values:

```text
<APP_NAME> = example-app
<DOMAIN> = example.dev
<ZONE_NAME> = example.dev
<REPO_NAME> = example-app
<REPO_PATH> = /srv/repos/example-app
<APP_ENV_PREFIX> = EXAMPLE_APP
<APP_PORT> = 3010
<HOOK_PORT> = 9010
<SERVICE_NAME> = example-web
<HOOK_SERVICE_NAME> = deploy-hook-example-app
```

Steps:

```text
1. Add/onboard <ZONE_NAME> in Cloudflare.
2. Change the domain's registrar nameservers to Cloudflare nameservers.
3. Make sure the Cloudflare API token can edit DNS for <ZONE_NAME>.
4. Clone repo into /srv/repos/<REPO_NAME>.
5. Add domain records to /etc/homeserver/cloudflare-ddns.records.
6. Run sudo /usr/local/bin/cloudflare-ddns.
7. Verify DNS with dig.
8. Add app HOST/PORT and hook HOST/PORT env vars.
9. Create app systemd service.
10. Add Caddy route for the domain.
11. Create deploy script in /srv/deploy/<APP_NAME>/deploy.sh.
12. Add exact sudoers permission for service restart.
13. Create deploy-hook env file.
14. Create deploy-hook systemd service.
15. Add GitHub webhook in that repo.
16. Test direct app, Caddy route, webhook route, manual deploy, and push deploy.
```

Example DDNS records:

```text
example.dev      example.dev                  false
example.dev      www.example.dev              false
```

Example env vars:

```env
EXAMPLE_APP_HOST=127.0.0.1
EXAMPLE_APP_PORT=3010

EXAMPLE_APP_HOOK_HOST=127.0.0.1
EXAMPLE_APP_HOOK_PORT=9010
```

---

# 20. Generic checklist for any new public app

## Inputs

```text
App name:
Domain:
Cloudflare zone:
Repo path:
Repo branch:
Package filter:
App env prefix:
App type: Next.js / Go / static / other
Local app port:
Webhook needed: yes/no
Webhook port:
Systemd service name:
Deploy script path:
```

## DNS / DDNS

```text
Domain zone exists in Cloudflare
Registrar nameservers point to Cloudflare
Cloudflare API token can edit the zone
A records added to /etc/homeserver/cloudflare-ddns.records
DDNS manual run works
DNS resolves publicly
```

## Router

```text
80/443 forwarded to server
No extra app ports forwarded
No webhook ports forwarded
No SSH port forwarded
No DMZ
```

## Server env

```text
HOST/PORT values added
Caddy restarted if env changed
```

## Shared data (section 8)

```text
POSTGRES_HOST / POSTGRES_PORT in server.env
MINIO_HOST / MINIO_PORT in server.env
App env has POSTGRES_USER / PASSWORD / DB / SSLMODE (host/port only in server.env)
Shared minio.service already running (bind from server.env MINIO_*; minio.env = root + volume only)
storage.<ZONE> Caddy site → {$MINIO_HOST}:{$MINIO_PORT}
Per-app MinIO user + bucket + public URL https://storage.<ZONE> (keys in apps/<app>.env, not root)
```

## App service

```text
systemd service created
systemd daemon-reload done
service enabled
service active
local curl works
```

## Caddy

```text
Caddy block added
Caddy formatted
Caddy validated
Caddy reloaded
local --resolve HTTPS curl works
```

## Deploy

```text
deploy.sh exists
deploy.sh executable
manual deploy works
sudoers exact restart permission added
```

## Webhook

```text
deploy-hook env file exists
deploy-hook service exists
deploy-hook service active
direct hook route returns 405
Caddy hook route returns 405
GitHub webhook created
push deploy works
```

---

# 21. Testing commands

Test app directly:

```bash
curl -I http://127.0.0.1:<APP_PORT>
```

Test Caddy domain locally:

```bash
curl -kI --resolve <DOMAIN>:443:127.0.0.1 https://<DOMAIN>/
```

Test webhook directly:

```bash
curl -i http://127.0.0.1:<HOOK_PORT>/_github/<APP_NAME>
```

Test webhook through Caddy:

```bash
curl -ki --resolve <DOMAIN>:443:127.0.0.1 https://<DOMAIN>/_github/<APP_NAME>
```

Check app logs:

```bash
journalctl -u <SERVICE_NAME> -f
```

Check webhook logs:

```bash
journalctl -u <HOOK_SERVICE_NAME> -f
```

Check Caddy logs:

```bash
journalctl -u caddy -f
```

Check DDNS logs:

```bash
journalctl -u cloudflare-ddns.service -n 100 --no-pager
```

Check DDNS timer:

```bash
systemctl list-timers cloudflare-ddns.timer
```

Check latest repo commit:

```bash
cd <REPO_PATH>
git log -1 --oneline
```

Check DNS:

```bash
dig +short <DOMAIN> @1.1.1.1
```

---

# 22. Reboot survival checklist

These services should be enabled:

```bash
systemctl is-enabled caddy
systemctl is-enabled ssh
systemctl is-enabled tailscaled
systemctl is-enabled cloudflare-ddns.timer
systemctl is-enabled <SERVICE_NAME>
systemctl is-enabled <HOOK_SERVICE_NAME>
```

These services should be active:

```bash
systemctl is-active caddy
systemctl is-active ssh
systemctl is-active tailscaled
systemctl is-active cloudflare-ddns.timer
systemctl is-active <SERVICE_NAME>
systemctl is-active <HOOK_SERVICE_NAME>
```

Full reboot test:

```bash
sudo reboot
```

After reboot:

```bash
ssh home-server
curl -I http://127.0.0.1:<APP_PORT>
curl -kI --resolve <DOMAIN>:443:127.0.0.1 https://<DOMAIN>/
systemctl list-timers cloudflare-ddns.timer
```

---

# 23. SSH and admin access

Use Tailscale for remote SSH.

Suggested laptop SSH config:

```sshconfig
Host home-server
    HostName <SERVER_LAN_IP>
    User <LINUX_USER>
    IdentityFile ~/.ssh/id_ed25519

Host home-server-ts
    HostName <SERVER_TS_IP>
    User <LINUX_USER>
    IdentityFile ~/.ssh/id_ed25519
```

Use at home:

```bash
ssh home-server
```

Use away from home:

```bash
ssh home-server-ts
```

Do not publicly forward SSH unless there is a very specific reason.

---

# 24. Troubleshooting

## Caddy returns 502

The local app is probably not running or not listening where Caddy expects.

Check:

```bash
systemctl status <SERVICE_NAME> --no-pager
journalctl -u <SERVICE_NAME> -n 100 --no-pager
curl -I http://127.0.0.1:<APP_PORT>
```

Check Caddy:

```bash
journalctl -u caddy -n 100 --no-pager
sudo caddy validate --config /etc/caddy/Caddyfile
```

## Caddy env vars are not loading

Check:

```bash
sudo systemctl cat caddy
sudo systemctl show caddy --property=EnvironmentFiles
```

Caddy must load the env file that defines the vars used in the Caddyfile.

After changing env vars:

```bash
sudo systemctl daemon-reload
sudo systemctl restart caddy
```

## DNS points to old WAN IP

Run DDNS manually:

```bash
sudo /usr/local/bin/cloudflare-ddns
```

Check logs:

```bash
journalctl -u cloudflare-ddns.service -n 100 --no-pager
```

Check the timer:

```bash
systemctl list-timers cloudflare-ddns.timer
```

Check DNS:

```bash
dig +short <DOMAIN> @1.1.1.1
```

Check that the record exists in:

```text
/etc/homeserver/cloudflare-ddns.records
```

Check that the API token has access to the Cloudflare zone.

## DDNS timer is not running

Check:

```bash
systemctl status cloudflare-ddns.timer --no-pager
systemctl list-timers cloudflare-ddns.timer
```

Enable:

```bash
sudo systemctl enable --now cloudflare-ddns.timer
```

Run manually:

```bash
sudo systemctl start cloudflare-ddns.service
```

## DDNS script fails

Run manually:

```bash
sudo /usr/local/bin/cloudflare-ddns
```

Check:

```bash
sudo cat /etc/homeserver/cloudflare-ddns.records
sudo ls -l /etc/homeserver/cloudflare-ddns.env
journalctl -u cloudflare-ddns.service -n 100 --no-pager
```

Common causes:

```text
API token missing
API token lacks DNS edit permission
API token cannot access the Cloudflare zone
records file has invalid format
zone_name does not exist in Cloudflare
record_name is misspelled
jq/curl not installed
```

Install dependencies:

```bash
sudo apt update
sudo apt install -y curl jq dnsutils
```

## GitHub webhook says 202 but repo does not update

GitHub 202 only means something responded successfully.

Check the GitHub webhook response body.

Expected for push:

```text
deploy started
```

Check logs:

```bash
journalctl -u <HOOK_SERVICE_NAME> -f
```

Run manual deploy:

```bash
/srv/deploy/<APP_NAME>/deploy.sh
```

Check repo commit:

```bash
cd <REPO_PATH>
git log -1 --oneline
```

## Webhook logs nothing

Test direct hook:

```bash
curl -i http://127.0.0.1:<HOOK_PORT>/_github/<APP_NAME>
```

Test through Caddy:

```bash
curl -ki --resolve <DOMAIN>:443:127.0.0.1 https://<DOMAIN>/_github/<APP_NAME>
```

Check running process:

```bash
ps -o user:20,pid,cmd -C deploy-hook
```

Check service file:

```bash
sudo systemctl cat <HOOK_SERVICE_NAME>
```

Check binary contains expected log text:

```bash
strings /opt/deploy-hook/bin/deploy-hook | grep "request received"
```

## Webhook ignores push events

Correct logic:

```go
if event != "push" {
	// ignore non-push events
}
```

Wrong logic:

```go
if event == "push" {
	// this accidentally ignores real push events
}
```

## Site works from laptop but not from server curl

The server may resolve the domain to the public WAN IP.

That is okay.

Use local Caddy test:

```bash
curl -kI --resolve <DOMAIN>:443:127.0.0.1 https://<DOMAIN>/
```

## Cloudflare proxied mode breaks something

Return the affected DNS records to:

```text
DNS only
```

Then retest:

```bash
dig +short <DOMAIN> @1.1.1.1
curl -kI --resolve <DOMAIN>:443:127.0.0.1 https://<DOMAIN>/
```

Only re-enable proxied mode after the DNS-only setup is healthy.

---

# 25. Future automation ideas

The repeated pieces should eventually become scripts.

Possible scripts:

```text
/srv/admin/add-next-app.sh
/srv/admin/add-go-app.sh
/srv/admin/add-static-site.sh
/srv/admin/add-repo-webhook.sh
/srv/admin/add-cloudflare-record.sh
/srv/admin/check-server.sh
```

A future `add-next-app.sh` could accept:

```bash
sudo /srv/admin/add-next-app.sh \
  --name <APP_NAME> \
  --domain <DOMAIN> \
  --zone <ZONE_NAME> \
  --repo <REPO_PATH> \
  --filter <PACKAGE_FILTER> \
  --env-prefix <APP_ENV_PREFIX> \
  --port <APP_PORT> \
  --hook-port <HOOK_PORT>
```

It should:

```text
1. Add HOST/PORT variables to server.env
2. Add DDNS record to cloudflare-ddns.records
3. Run cloudflare-ddns
4. Create systemd app service
5. Add Caddy block
6. Validate/reload Caddy
7. Create deploy script
8. Add sudoers permission
9. Create deploy-hook env file
10. Create deploy-hook service
11. Print GitHub webhook URL
12. Print validation commands
13. Run local validation tests
```

---

# 26. Final rule

A hosted app is not considered done until all of these work:

```bash
curl -I http://127.0.0.1:<APP_PORT>
curl -kI --resolve <DOMAIN>:443:127.0.0.1 https://<DOMAIN>/
dig +short <DOMAIN> @1.1.1.1
systemctl is-active <SERVICE_NAME>
systemctl is-enabled <SERVICE_NAME>
```

A webhook deploy is not considered done until all of these work:

```bash
/srv/deploy/<APP_NAME>/deploy.sh
curl -i http://127.0.0.1:<HOOK_PORT>/_github/<APP_NAME>
curl -ki --resolve <DOMAIN>:443:127.0.0.1 https://<DOMAIN>/_github/<APP_NAME>
systemctl is-active <HOOK_SERVICE_NAME>
systemctl is-enabled <HOOK_SERVICE_NAME>
```

DDNS is not considered done until all of these work:

```bash
sudo /usr/local/bin/cloudflare-ddns
systemctl is-active cloudflare-ddns.timer
systemctl is-enabled cloudflare-ddns.timer
systemctl list-timers cloudflare-ddns.timer
dig +short <DOMAIN> @1.1.1.1
```

Then perform a real GitHub push and verify:

```bash
journalctl -u <HOOK_SERVICE_NAME> -f
cd <REPO_PATH>
git log -1 --oneline
```
