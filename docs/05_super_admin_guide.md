# Superadmin Guide

The React dashboard at `http://superadmin.localhost:8090` is the central tenant control plane in local development.

## Dashboard capabilities

- Create tenants with a company name and validated lowercase subdomain
- Receive creation state and log updates over WebSocket
- Search tenants and filter by state
- Copy the tenant API key displayed in the table
- Disable an active tenant or enable a disabled tenant
- Set and display license expiry
- View agent monitoring for active tenants
- Validate, stage, activate, and roll back tenant customization releases

The backend has deletion and audit-list endpoints, but the current dashboard does not expose delete or audit controls.

## Tenant creation

Creation immediately returns a tenant in `creating` state, then runs in a Go goroutine. The backend creates the database, isolated addon directory, init container, and daemon container. Success changes state to `active`; an error changes it to `error` and appends a creation log.

The init container installs `base`, `web`, `sale_management`, `stock`, `daily_sales_report`, `initial_data_import`, and `tenant_management_agent`.

## Monitoring

The management agent is installed during provisioning and authenticated with a private per-tenant token. It reports contract version, health, database, company, Odoo version, internal users, products, warehouses, and installed modules classified as tenant, platform, or core.

## Lifecycle behavior

Disable stops the tenant container and marks the record disabled. Enable starts the existing container. Setting expiry only stores a date; it does not automatically disable access. API deletion force-removes the container, removes the tenant directory, drops the database, and deletes the central record.

## Customization

Customization mutation requests require `X-Customization-Admin-Token`. The browser stores the entered value in session storage. Read-only release listing and audit endpoints currently do not require that token.

## Security warning

General superadmin routes currently have no login or role enforcement. Keep this interface network-restricted and do not treat it as internet-ready. See [Security and Limitations](./security_and_limitations.md).
