# Extending the Platform

## Choose the ownership layer

- Put behavior required for every managed tenant in `custom_addons/` as a platform addon.
- Put customer-specific behavior in a separate tenant module delivered through the release workflow.
- Put lifecycle, monitoring, and provisioning behavior in the Go control plane.
- Keep the management-agent HTTP contract small and versioned; do not make central monitoring depend on customer tables.

## Platform Odoo modules

Develop and test against Odoo 18, which is the project runtime. Add a platform module to the init list in `backend/services/docker.go` only when every new tenant should receive it. Updating the init list affects new tenants; it does not batch-upgrade existing tenant databases.

For an existing fleet, define and test an explicit rollout procedure before merging a platform module change.

## Tenant modules

Use a unique technical name, declare dependencies, include access controls, and test installation plus upgrade in a disposable tenant. Package exactly one module directory in the ZIP. Avoid direct Docker access, subprocess execution, shared filesystem assumptions, and irreversible migrations.

## Control-plane APIs

Register routes in `backend/main.go`, keep database models in `backend/models`, and separate handlers from services. Add authentication before expanding administrative API exposure. Update the generated System Map assumptions if route declaration style changes.

## Frontend

The superadmin is a compact operational React interface. Preserve searchable tenant tables, explicit lifecycle confirmations, stable modal dimensions, responsive layouts, and clear failure states. The production frontend build runs ESLint and Vite.

## Mobile

The Flutter app and Go mobile endpoints are prototypes with known integration gaps. Correct per-tenant authentication routing, externalize the base URL, and add real token/session semantics before adding features. Do not build new behavior around the tenant-wide API key as if it were a user-scoped token.

## Required update checklist

- Add focused tests proportional to the change
- Build affected images
- Test on a disposable tenant
- Update Feature Status and the relevant team workflow
- Mark limitations honestly when the complete user path is not connected
