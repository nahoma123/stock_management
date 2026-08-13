# Tenant Customization

The platform supports operator-reviewed tenant Odoo modules through a managed release workflow. Tenants do not directly edit server files or upload packages from inside Odoo.

## Addon layers

1. Tenant releases at `/mnt/tenant-addons`
2. Read-only platform addons at `/mnt/platform-addons`
3. Odoo packaged addons

A tenant package cannot use a module name already present in the platform layer.

## Package contract

Upload one ZIP of at most 25 MB containing exactly one top-level module directory. The directory name must match the allowed lowercase Odoo module pattern and contain `__manifest__.py`.

Validation rejects traversal, absolute paths, backslash paths, symlinks, multiple top-level directories, oversized expansion, and a small set of obvious host-process access strings. Validation is not a Python sandbox; approved code runs inside the tenant Odoo process.

## Release lifecycle

1. **Validate** inspects the package without storing a release.
2. **Deploy** creates a central release record and stages immutable files under `.releases/<module>/<release-id>`.
3. **Activate** stops Odoo, switches the active relative symlink, runs an isolated Odoo install/update container, and recreates the daemon.
4. The prior active release becomes `superseded` and remains available for rollback.
5. A failed activation restores the prior link and attempts maintenance/restart against it.

Release states include `staging`, `staged`, `active`, `superseded`, and `failed`. Stage, activation, and rollback outcomes create central audit events.

## Rollback boundary

Rollback restores code and runs Odoo module maintenance. It cannot generically reverse destructive data migrations or external side effects. Tenant module authors must use compatible schema evolution and provide explicit migration recovery when needed.

## Authentication

Validate, deploy, activate, and rollback require `CUSTOMIZATION_ADMIN_TOKEN`. Release listing is currently readable through an unauthenticated control-plane endpoint, so network restriction is still required.

See [Customization Architecture](./architecture/tenant_customization.md) for implementation boundaries.
