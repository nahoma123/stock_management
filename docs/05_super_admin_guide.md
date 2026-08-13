# Super Admin Guide

The React superadmin at `http://superadmin.localhost` is the single control plane for tenant lifecycle
operations. Its Go API stores tenant metadata, provisions tenant databases and containers, and communicates
with each Odoo instance through the authenticated management-agent contract.

## Create a tenant

1. Open the superadmin dashboard.
2. Enter the company name and a unique lowercase subdomain.
3. Select **Create Tenant**.
4. Follow provisioning progress in the creation log.

Provisioning creates the database, isolated addon directory, management credential, initialized Odoo
modules, and tenant container. A successful tenant becomes `active`; failures become `error` and retain their
creation log for diagnosis.

## Tenant operations

The dashboard supports searching tenants, inspecting creation logs, enabling or disabling a tenant, setting
license expiry, opening monitoring, deploying tenant customizations, and deleting a tenant. Deletion removes
the tenant container, database, addon files, and central tenant record.

Monitoring is available immediately after successful provisioning. The management-agent credential is
private and is never returned by the tenant API.

## Customizations

Tenant Odoo modules are uploaded as ZIP packages from the customization view. Validate a package first, then
stage it as a release. Activation stops the tenant, installs or upgrades the module, and restarts Odoo.
Previous releases remain available for rollback. See `docs/architecture/tenant_customization.md` for the
layering and security model.

Set `CUSTOMIZATION_ADMIN_TOKEN` before using protected customization actions. The UI keeps this token only in
browser session storage.

## Troubleshooting

Check the tenant creation log first, followed by `docker compose logs backend` and the relevant
`odoo_tenant_<id>` container logs. Common causes are database connectivity, invalid Odoo modules, unavailable
Docker access, or incorrect host/network configuration.
