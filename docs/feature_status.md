# Feature Status

Source audit date: **2026-08-13**.

This matrix is based on the current branch implementation. “Available” means the primary code path is connected; it does not imply production hardening.

## Available

| Capability | Current implementation |
| --- | --- |
| Tenant provisioning | React form creates a central record; Go creates the database, addon directory, init container, and tenant container asynchronously. |
| Live creation progress | Backend broadcasts tenant state and log events over WebSocket. |
| Tenant lifecycle | Enable, disable, expiry update, and backend deletion endpoints exist. The UI exposes enable, disable, and expiry; deletion is API-only. |
| Tenant isolation | Separate PostgreSQL database, Odoo container, and tenant addon directory per tenant. |
| Central monitoring | Authenticated management agent reports status, Odoo version, company, users, products, warehouses, and installed modules by layer. |
| Initial product import | Stock managers can preview and import CSV/XLSX products, ETB/USD prices, one embedded image, and STOCK-1 through STOCK-3. |
| Daily sales email | Daily cron calculates confirmed daily orders and total sales and emails the configured owner/company address. Mail delivery still depends on Odoo outgoing-mail configuration. |
| Tenant customization releases | ZIP validation, immutable staging, activation, Odoo install/update maintenance, audit events, release history, and rollback. |
| Sales dashboard API | API-key endpoint returns all-time, month, today, yesterday, seven-day trend, five recent orders, and five top products. |
| Documentation site | Team and developer areas render maintained Markdown and generate a System Map from Compose and Go sources. |

## Prototype or incomplete

| Capability | Limitation |
| --- | --- |
| Flutter mobile app | Screens, local credential storage, dashboard, chart, recent orders, top products, and settings are implemented. Login currently calls `http://odoo:8069`, but tenants run as `odoo_tenant_<id>`, so authentication must be corrected before end-to-end use. The app base URL is hardcoded for local development. |
| Push notifications | Device registration and threshold settings exist. Sale webhooks only log a dummy notification; no FCM/APNs provider sends a real push. The webhook has no secret/signature, and its Odoo addon is not installed by default. |
| Shopping portal | An optional addon with registration and product templates exists, but it is not part of default tenant provisioning and is not covered by the managed release UI as a platform feature. |
| License expiry | Expiry can be stored and displayed. There is no automated enforcement that disables an expired tenant. |
| Audit viewing | Audit records are stored and available from the API. The React dashboard does not currently show an audit view. |

## Not implemented

- Billing and subscription automation
- Operator login, roles, or access control for the superadmin API/UI
- Real push delivery
- Automated backups and restore workflows
- High availability or multi-host scheduling
- Customer-owned domains and production TLS automation
- Resource quotas and usage metering
- Batch rollout of platform module upgrades to all existing tenants

## Required validation before customer readiness

- Test product import with a representative customer workbook
- Add and test control-plane authentication
- Test backup and full restore, including Odoo filestores
- Run a complete disposable-tenant provisioning and deletion exercise
- Establish outgoing email and verify daily report delivery
- Resolve or explicitly exclude the mobile login and push prototype
