#!/usr/bin/env bash
# Template: copy to /srv/deploy/rubrical/deploy.sh (same layout as spencerls).
# Do not run from the git checkout — git reset --hard would rewrite this file mid-deploy.
#
#   sudo mkdir -p /srv/deploy/rubrical
#   sudo cp /srv/repos/rubrical/deploy/homeserver/deploy.sh /srv/deploy/rubrical/deploy.sh
#   sudo chown "$USER:$USER" /srv/deploy/rubrical/deploy.sh
#   chmod +x /srv/deploy/rubrical/deploy.sh
set -euo pipefail

REPO="${RUBRICAL_REPO:-/srv/repos/rubrical}"
BRANCH="${DEPLOY_BRANCH:-main}"
LOCK_FILE="/tmp/rubrical-deploy.lock"
ENV_FILE="${RUBRICAL_ENV_FILE:-/etc/homeserver/apps/rubrical.env}"
SERVER_ENV="${RUBRICAL_SERVER_ENV:-/etc/homeserver/server.env}"
SERVICE="${RUBRICAL_SERVICE:-rubrical.service}"

(
  flock -n 9 || {
    echo "Deploy already running; exiting."
    exit 0
  }

  echo "=== Rubrical deploy started at $(date) ==="
  echo "repo=$REPO branch=$BRANCH"

  if [[ ! -f "$ENV_FILE" ]]; then
    echo "missing env file: $ENV_FILE" >&2
    exit 1
  fi
  if [[ ! -f "$SERVER_ENV" ]]; then
    echo "missing server env: $SERVER_ENV" >&2
    exit 1
  fi

  set -a
  # shellcheck disable=SC1090
  source "$SERVER_ENV"
  # shellcheck disable=SC1090
  source "$ENV_FILE"
  set +a

  cd "$REPO"

  git fetch origin "$BRANCH"
  git reset --hard "origin/$BRANCH"

  corepack enable || true
  pnpm install --frozen-lockfile

  # templ before css so Tailwind scans generated *_templ.go as well as .templ
  make templ css
  make migrate-up
  make build
  make extension-package

  sudo /usr/bin/systemctl restart "$SERVICE"

  echo "=== Rubrical deploy finished at $(date) ==="
) 9>"$LOCK_FILE"
