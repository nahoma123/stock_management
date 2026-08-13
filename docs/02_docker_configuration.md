# Docker Configuration

## Compose services

`docker-compose.yml` defines PostgreSQL, the Go backend, React frontend, documentation site, and Traefik. The System Map page reads this file directly and lists the current services.

Traefik listens on host port `8090`; its dashboard is exposed on `8091`. Local URLs therefore include `:8090`:

- Superadmin: `http://superadmin.localhost:8090`
- Documentation: `http://docs.localhost:8090`
- Tenant example: `http://acme.localhost:8090`

## Dynamic tenant containers

The backend mounts `/var/run/docker.sock` and creates:

- `odoo_init_<id>` for initial database/module setup
- `odoo_maintenance_<id>` for a customization install or update
- `odoo_tenant_<id>` for the long-running tenant

All use the `odoo_project-odoo:latest` image and `saas_net` network.

## Addon mounts

The backend sees `./tenants` at `/app/tenants` and `./custom_addons` at `/mnt/extra-addons`. Tenant containers receive:

- `tenants/<subdomain>/custom_addons:/mnt/tenant-addons`
- `custom_addons:/mnt/platform-addons:ro`

Their addon precedence is tenant, platform, then packaged Odoo addons.

## Data persistence

`postgres-data` persists PostgreSQL. Tenant release files persist in `tenants/`. The current dynamic tenant-container configuration does not define a dedicated persistent `/var/lib/odoo` filestore volume, which is a production-readiness gap for attachment-heavy workflows.

## Required configuration

- `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_USER`, `POSTGRES_PASSWORD`
- `HOST_PROJECT_PATH`: absolute host checkout used for Docker bind mounts
- `DOCKER_NETWORK`: normally `saas_net`
- `CUSTOMIZATION_ADMIN_TOKEN`: protects validation and release mutation endpoints

Compose currently defaults the database password to `odoo`, exposes PostgreSQL on host port `5433`, and allows an empty customization token. Replace these defaults outside local development.
