# Deployment

Rubrical uses the shared **homeserver** platform and root [`homeserver.yaml`](../homeserver.yaml).

| Host | Role |
| --- | --- |
| `rubrical.spencerls.dev` | App + deploy webhook |

```bash
sudo /srv/repos/homeserver/admin/install-platform.sh --install-deps
cd /srv/repos && git clone <rubrical-remote> rubrical
sudo /srv/repos/homeserver/admin/install-pack.sh /srv/repos/rubrical
```

`install-pack` provisions Postgres (`databases.rubrical`), generates the encryption key, and wires units/Caddy/DDNS.

```bash
git push origin main
# → homeserver-deploy rubrical
```

Purge: `rubrical-purge.timer`.
