# Tenant Customization Architecture

Every managed tenant has three addon layers, in precedence order:

1. `/mnt/tenant-addons` contains company-owned custom modules. Each tenant has a separate directory.
2. `/mnt/platform-addons` contains read-only platform modules maintained by the operator.
3. Odoo's packaged addon directory contains pinned core applications.

Tenant packages must contain exactly one top-level Odoo module directory with a valid module name and
`__manifest__.py`. The superadmin validates archive structure, rejects links, traversal, platform-name
collisions, oversized files, and obvious host-process access before deployment.

The `tenant_management_agent` module is the stable boundary between the platform and customized tenant
schemas. Its `/tenant-agent/v1/health` contract is authenticated by a per-tenant token and versioned
independently from tenant modules. Central monitoring must use this contract instead of querying custom
tenant tables directly.

Customization packages are privileged executable Python code. Only trusted operator-reviewed packages
should be deployed. Validation reduces accidental and common unsafe packages; it is not a sandbox.

## Operator configuration

Set `CUSTOMIZATION_ADMIN_TOKEN` in the root `.env` file before starting the Compose stack. The React
superadmin asks for this token only when validating or deploying a customization and retains it in browser
session storage. Use a long random value and rotate it when operator access changes.

New tenants are provisioned with a separate random management-agent token that is never serialized by the
superadmin API. Existing tenants created before this architecture require a future enrollment/reprovisioning
step before their monitoring endpoint becomes available.
