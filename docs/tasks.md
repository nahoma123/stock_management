# Implementation Tracker

This tracker summarizes current source status. Use [Feature Status](./feature_status.md) for behavior and limitations.

## Completed foundations

- [x] Go/React tenant control plane
- [x] Per-tenant database, Odoo container, and addon directory provisioning
- [x] WebSocket tenant state and creation-log updates
- [x] Enable, disable, expiry, monitoring, and backend deletion operations
- [x] Authenticated tenant management-agent contract
- [x] Central monitoring metrics and module-layer inventory
- [x] CSV/XLSX product import with preview, ETB/USD prices, embedded primary image, and opening stock
- [x] Daily sales summary email addon and owner-email setting
- [x] Versioned customization validation, staging, activation, audit, and rollback
- [x] Team/developer documentation website and generated System Map

## Acceptance still required

- [ ] Verify product import against a representative customer workbook
- [ ] Verify outgoing email and daily report delivery in a configured tenant
- [ ] Test customization activation and rollback with a representative tenant module and data migration
- [ ] Test complete tenant backup and restore after persistent filestore design is added

## Partially implemented

- [x] Flutter login, dashboard, chart, recent orders, top products, and settings screens exist
- [ ] Route Flutter/Odoo authentication to the selected `odoo_tenant_<id>` container
- [ ] Replace hardcoded mobile base URL with environment configuration
- [x] Device registration, notification threshold, sale webhook, and Odoo hook exist
- [ ] Authenticate the sale webhook
- [ ] Install the notification addon through an intentional rollout if the feature is retained
- [ ] Connect a real FCM/APNs provider; current delivery is log-only

## Production blockers

- [ ] Add operator authentication and authorization to the superadmin
- [ ] Add persistent tenant filestore storage and tested backups
- [ ] Replace development credentials and restrict exposed services
- [ ] Add observability, resource limits, and production health checks
- [ ] Configure production DNS/TLS and network policy
- [ ] Add durable provisioning recovery/idempotency

## Product roadmap

- [ ] Automated license-expiry enforcement
- [ ] Billing and subscription management
- [ ] Audit-event UI with operator attribution
- [ ] Batch platform-addon rollout to existing tenants
- [ ] Customer domains and automated certificate lifecycle
- [ ] High availability and multi-host scheduling
