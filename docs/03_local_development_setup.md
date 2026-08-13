# Local Development Setup

## Start the stack

1. Set a non-empty `CUSTOMIZATION_ADMIN_TOKEN` in `.env`.
2. Run `docker compose build`.
3. Run `docker compose up -d`.
4. Open `http://superadmin.localhost`.

Traefik routes `*.localhost` automatically in modern browsers. A tenant created with subdomain `acme` is
available at `http://acme.localhost` after provisioning finishes.

## Create a test tenant

Use the **Create Tenant** form in the React dashboard. The Go backend creates the tenant record, database,
isolated addon directory, Odoo initialization container, and long-running tenant container. Watch progress in
the dashboard creation log or with:

```bash
docker compose logs -f backend
```

## Development checks

```bash
docker compose build backend frontend
docker compose ps
```

The frontend image runs ESLint and the Vite production build. The backend image compiles from vendored Go
dependencies. Backend unit tests can be run with the repository's Go version in a container.

## Troubleshooting

Inspect `docker compose ps`, backend logs, and the relevant `odoo_tenant_<id>` logs. Confirm PostgreSQL is
healthy, the backend can access the Docker socket, `HOST_PROJECT_PATH` matches the checkout, and all services
use the configured Docker network.
