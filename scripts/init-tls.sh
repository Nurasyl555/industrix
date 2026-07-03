#!/bin/bash
set -e

# One-time TLS bootstrap. Run this ONCE on the server, BEFORE ever starting
# the nginx service — nginx's config references cert files that don't exist
# yet, so it won't start until this has run at least once.
#
# Usage: ./scripts/init-tls.sh <domain> <email>
# Example: ./scripts/init-tls.sh industrix.kz admin@industrix.kz

DOMAIN=$1
EMAIL=$2

if [ -z "$DOMAIN" ] || [ -z "$EMAIL" ]; then
  echo "Usage: ./scripts/init-tls.sh <domain> <email>"
  exit 1
fi

echo "Make sure $DOMAIN's DNS A record already points at this server, and"
echo "that YOUR_DOMAIN_HERE in infra/nginx/conf.d/site.conf has been"
echo "replaced with $DOMAIN — press Enter to continue, Ctrl+C to abort."
read -r

# nginx isn't running yet (it can't start without a cert), so port 80 is
# free — reuse the certbot service's own volumes via `compose run` so the
# cert lands exactly where the real certbot/nginx services expect it.
# --entrypoint overrides the service's default renew-loop entrypoint for
# this one-off run.
docker compose -f docker-compose.prod.yml run --rm -p 80:80 --entrypoint certbot \
  certbot certonly --standalone --non-interactive --agree-tos --email "$EMAIL" -d "$DOMAIN"

echo "Certificate obtained. Now run:"
echo "  docker compose -f docker-compose.prod.yml up -d nginx certbot"
