# Docker Configuration

`docker-compose.yml` defines one PostgreSQL service, the Go backend, React/Nginx frontend, Traefik, and the
Odoo image used for tenant containers. The backend creates tenant init and daemon containers through the
Docker socket; there is no separate Odoo superadmin service.

The backend mounts:

- `./custom_addons` at `/mnt/extra-addons` for read-only platform modules in tenant containers.
- `./tenants` at `/app/tenants` for isolated tenant addon directories and releases.
- `/var/run/docker.sock` to create, stop, and replace tenant containers.

Tenant Odoo containers receive two addon mounts: their own addon directory at `/mnt/tenant-addons` and the
shared platform directory at `/mnt/platform-addons:ro`. PostgreSQL holds the control-plane tables and one
database per tenant. Traefik routes `superadmin.localhost` to the frontend and tenant subdomains to their
individual Odoo containers.

Important environment variables are `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_USER`, `POSTGRES_PASSWORD`,
`HOST_PROJECT_PATH`, `DOCKER_NETWORK`, and `CUSTOMIZATION_ADMIN_TOKEN`.
