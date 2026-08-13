# Local Development

## Prerequisites

- Docker Engine with Compose
- Git
- Node 20 when running either frontend outside Docker
- Flutter SDK only when working on `mobile_app/`

The backend build uses Go 1.21 in Docker and vendored dependencies. Tenant Odoo runs on Odoo 18.

## Start the platform

Set `CUSTOMIZATION_ADMIN_TOKEN` in the ignored root `.env`, confirm `HOST_PROJECT_PATH` in Compose matches the checkout, then run:

```bash
docker compose build
docker compose up -d
docker compose ps
```

Open `http://superadmin.localhost:8090`. Create a disposable tenant and open it at `http://<subdomain>.localhost:8090` after its state becomes active.

## Documentation development

```bash
cd documentation
npm install
npm run dev
```

Open `http://localhost:4174`. Changes under `docs/`, Compose, backend routes, and the default Odoo module list hot reload into the site.

## Checks

```bash
docker compose build backend frontend documentation
docker compose config --quiet
```

Run backend tests in the repository’s Go image:

```bash
docker run --rm -v "$PWD/backend:/app" -w /app golang:1.21-alpine \
  sh -lc '/usr/local/go/bin/go test -mod=vendor ./...'
```

The frontend Docker build runs ESLint and Vite. The documentation build runs Vite. Odoo addon verification requires initializing or upgrading the relevant module in a disposable database.

## Logs

```bash
docker compose logs -f backend
docker logs -f odoo_tenant_<id>
docker logs odoo_init_<id>
```

Initialization containers are normally removed after completion, so use the dashboard creation log first.
