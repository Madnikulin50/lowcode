#!/usr/bin/env bash
# Simple launcher for local dev: mints a fresh Compose token automatically
# (via compose/helpers.mjs's mintToken(), same DSN probing seed.mjs uses)
# and starts stroykontrol-web against the real, running Compose instance.
# No fixtures, no manual token copy-pasting, no debugger env vars.
#
#   ./run.sh                                    # defaults below
#   TOKEN=<token> ./run.sh                       # skip minting, use this token
#   NAMESPACE_ID=<id> COMPOSE_API=<url> ./run.sh # override defaults
#   ./run.sh --listen=:9000                      # extra args pass through
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

: "${COMPOSE_API:=http://localhost:3333/compose}"
: "${NAMESPACE_ID:=512312736219004929}" # "Стройконтроль ПТО" dev namespace

if [ -z "${TOKEN:-}" ]; then
  echo "minting a Compose token (mintToken(), see compose/helpers.mjs)..." >&2
  TOKEN=$(node -e "import('./compose/helpers.mjs').then(h=>h.mintToken()).then(t=>process.stdout.write(t))")
fi

go build -o bin/stroykontrol-web .
exec ./bin/stroykontrol-web --api="$COMPOSE_API" --token="$TOKEN" --namespace="$NAMESPACE_ID" "$@"
