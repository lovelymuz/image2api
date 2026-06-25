#!/bin/sh
# Reload nginx whenever the certificate file changes — i.e. once acme.sh has
# issued/renewed the real cert into the shared volume, nginx picks it up within
# ~a minute without a container restart. Runs in the background so it doesn't
# block startup.
D="${DOMAIN:-localhost}"
CERT="/etc/nginx/certs/live/$D/fullchain.pem"
(
  last=""
  while true; do
    sleep 60
    cur="$(stat -c %Y "$CERT" 2>/dev/null || echo '')"
    if [ -n "$cur" ] && [ "$cur" != "$last" ]; then
      # Skip the very first observation (the self-signed bootstrap); only reload
      # on a genuine change (acme.sh overwrote the cert).
      [ -n "$last" ] && nginx -s reload 2>/dev/null || true
      last="$cur"
    fi
  done
) &
