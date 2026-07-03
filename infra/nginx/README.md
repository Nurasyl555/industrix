# nginx + TLS setup

`conf.d/site.conf` is the only active config for the MVP — it terminates
TLS and routes `/api/v1`, `/health`, `/swagger` to the backend and
everything else to the frontend. `conf.d/future/` holds configs for later
phases (chat WebSocket gateway, gRPC gateway, MinIO subdomain) that aren't
loaded — none of those modules exist yet.

## First deploy on a new server

1. Point your domain's DNS A record at the server's IP.
2. Replace every `YOUR_DOMAIN_HERE` in `conf.d/site.conf` with your real
   domain (e.g. `industrix.kz`).
3. Set the `PUBLIC_API_URL` GitHub Actions repo Variable to
   `https://<your-domain>/api/v1` (Settings -> Secrets and variables ->
   Actions -> Variables) — the frontend build bakes this in, so it needs to
   be set before the next CD run.
4. Get the first certificate — nginx won't start without one:
   ```
   chmod +x scripts/init-tls.sh
   ./scripts/init-tls.sh <your-domain> <your-email>
   ```
5. Start everything:
   ```
   docker compose -f docker-compose.prod.yml up -d
   ```

The `certbot` service in `docker-compose.prod.yml` handles renewal
automatically from then on (checks every 12h, renews if within 30 days of
expiry — this is certbot's own default behavior, not something configured
here).
