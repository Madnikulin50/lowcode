#!/bin/sh
set -e

# Generate HTML from markdown mounted at /docs
if [ -d /docs ]; then
    echo "Generating HTML from /docs..."
    DOCS_DIR=/docs OUT_DIR=/usr/share/nginx/html python3 /tmp/build-html.py
    echo "Done. Starting nginx..."
else
    echo "WARNING: /docs not found — using prebuilt HTML if available"
fi

exec nginx -g "daemon off;"