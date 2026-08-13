# Production Deployment Guide

This project currently provides a single-host Docker Compose deployment. Production operation requires a
Linux host with Docker Engine, Compose, persistent storage, DNS, TLS termination, and a backup destination.

## Configuration

1. Point wildcard tenant DNS and the superadmin hostname at the deployment host.
2. Set strong PostgreSQL credentials and a long random `CUSTOMIZATION_ADMIN_TOKEN` in the deployment
   environment.
3. Set `HOST_PROJECT_PATH` to the absolute checkout path used for tenant addon bind mounts.
4. Restrict access to the Docker socket, database port, and superadmin route at the host/network level.
5. Configure Traefik with production host rules and TLS certificates.

Never commit `.env`, API keys, database passwords, agent tokens, or customer customization packages.

## Deploy

```bash
docker compose build
docker compose up -d
docker compose ps
```

Confirm the database health check passes, the backend can reach Docker, the frontend is available through
Traefik, and a disposable tenant can complete provisioning and monitoring.

## Updates

Build and test images before replacement. Apply control-plane updates first, then replace tenant containers
when platform Odoo code changes. Customer modules should move through validate, stage, and activate; use the
recorded prior release for rollback when an activation fails operational checks.

## Backups

Back up the PostgreSQL data volume and `tenants/` directory together. The database contains both control-plane
records and tenant databases; `tenants/` contains customer addon releases. Test full restoration regularly.
Odoo filestore persistence must also be included before enabling features that store attachments outside the
database.

## Operations

Monitor backend, PostgreSQL, Traefik, and tenant container health. Central audit events record customization
operations, while tenant creation logs capture provisioning. Establish resource limits, log retention,
database maintenance, certificate renewal, and tested incident procedures before onboarding customers.

The current design is appropriate for an initial managed deployment. Multi-host scheduling, high availability,
off-host secrets, automated backups, billing, and disaster recovery automation remain production roadmap work.
