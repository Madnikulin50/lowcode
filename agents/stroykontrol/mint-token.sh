#!/usr/bin/env bash
# Mints a fresh Compose token and writes it to var/dev-token — nothing more.
# Meant to run as a GoLand "Before launch: Run External Tool" step ahead of
# the stroykontrol-web debug configuration (see README's GoLand section),
# so hitting Debug always has a live token, never a manually copy-pasted one
# that quietly goes stale mid-session (the JWT this mints is short-lived,
# ~2h — see mintToken() in compose/helpers.mjs).
#
#   ./mint-token.sh                 # writes var/dev-token
#   TOKEN=<token> ./mint-token.sh   # just echoes that token to the file, no minting
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"
mkdir -p var
node -e "import('./compose/helpers.mjs').then(h=>h.mintToken()).then(t=>process.stdout.write(t))" > var/dev-token
echo "mint-token.sh: wrote $(wc -c < var/dev-token) bytes to var/dev-token" >&2
