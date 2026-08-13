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

Validated packages are staged as immutable releases under the tenant addon directory. Activation stops the
tenant, atomically changes the active module link, runs Odoo module maintenance in a transient container,
and restarts the tenant. A failed activation restores the prior code link and reruns maintenance against it.
Superseded releases remain available for operator-triggered rollback. Release changes are recorded in the
central audit event table. Tenant addon modules must be introduced through this release pipeline; unmanaged
module directories are rejected during activation.

The Go provisioning API and React superadmin are the tenant registration authority. Tenant provisioning
installs the management agent and assigns its private credential as part of the initial transaction, so no
separate registration or enrollment path exists.

## Operator configuration

Set `CUSTOMIZATION_ADMIN_TOKEN` in the root `.env` file before starting the Compose stack. The React
superadmin asks for this token in the customization workflow and retains it in browser session storage for
validation, staging, activation, and rollback. Use a long random value and rotate it when operator access changes.

Each tenant is provisioned with a separate random management-agent token that is never serialized by the
superadmin API.
