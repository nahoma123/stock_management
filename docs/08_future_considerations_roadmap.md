# Roadmap

This page contains work not implemented in the current system. Completed capabilities belong in [Feature Status](./feature_status.md), not here.

The dependency-ordered execution plan is maintained in [Platform Finalization Plan](./finalization_plan.md).

## Priority: security and operability

- Operator authentication, roles, session management, and audit attribution
- Secret-authenticated or signed internal webhooks
- Persistent Odoo filestores and automated tested backups
- Health checks, metrics, centralized logs, alerting, and resource limits
- Production DNS, TLS, firewall, and database network restrictions
- Provisioning idempotency and a supported retry/recovery workflow

## Priority: finish or remove prototypes

- Route mobile login to the selected tenant container
- Replace tenant-wide API-key access with user-scoped mobile sessions
- Externalize mobile environment configuration
- Integrate FCM/APNs or remove push-notification claims and surfaces
- Decide whether `shopping_portal` is a supported product feature

## Product operations

- Automated license-expiry enforcement
- Billing, subscription plans, payment state, and invoices
- Audit-event UI and operator identity
- Controlled tenant deletion in the React dashboard
- Batch platform-module rollout with canary tenants and failure isolation
- Customer onboarding checklist and import approval records

## Scale and resilience

- Multi-host container scheduling
- PostgreSQL capacity strategy, replication, and recovery objectives
- Object storage or managed volumes for filestores
- Per-tenant CPU, memory, and storage quotas
- Custom domains and automated certificate lifecycle
- Disaster-recovery exercises and documented RPO/RTO

Roadmap priority should follow customer evidence, security risk, and operational load rather than the age of an item.
