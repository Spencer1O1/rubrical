#!/usr/bin/env bash
# Auto-deploy entrypoint for the home server.
# Invoked by deploy-hook; safe to run manually:
#   /srv/repos/rubrical/deploy/homeserver/deploy.sh
set -euo pipefail

REPO="$(cd "$(dirname "$0")/../.." && pwd)"
BRANCH="${DEPLOY_BRANCH:-main}"
LOCK_FILE="/tmp/rubrical-deploy.lock"
ENV_FILE="${RUBRICAL_ENV_FILE:-/etc/homeserver/apps/rubrical.env}"
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

  set -a
  # shellcheck disable=SC1090
  source "$ENV_FILE"
  set +a

  cd "$REPO"

  git fetch origin "$BRANCH"
  git reset --hard "origin/$BRANCH"

  corepack enable || true
  pnpm install --frozen-lockfile

  make css templ
  make migrate-up
  make build

  sudo /usr/bin/systemctl restart "$SERVICE"

  echo "=== Rubrical deploy finished at $(date) ==="
) 9>"$LOCK_FILE"
