#!/bin/sh
# Bootstrap a self-signed cert so nginx's 443 server block can start BEFORE
# acme.sh has issued the real certificate. acme.sh later overwrites these files
# in the shared volume; 30-cert-watch.sh reloads nginx when that happens.
set -e
D="${DOMAIN:-localhost}"
CERT_DIR="/etc/nginx/certs/live/$D"
mkdir -p "$CERT_DIR" /var/www/acme
if [ ! -s "$CERT_DIR/fullchain.pem" ] || [ ! -s "$CERT_DIR/privkey.pem" ]; then
  echo "nginx: generating self-signed bootstrap cert for $D"
  openssl req -x509 -newkey rsa:2048 -nodes -days 3650 \
    -keyout "$CERT_DIR/privkey.pem" -out "$CERT_DIR/fullchain.pem" \
    -subj "/CN=$D" >/dev/null 2>&1
fi
