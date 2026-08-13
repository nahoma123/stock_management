# Platform Finalization Plan

This plan orders the work needed to make the superadmin, customer-company model, tenant lifecycle, and deployment architecture coherent and production-ready. Complete phases in order; later work depends on the ownership and security decisions in Phase 1.

## Target relationship model

Use these concepts consistently:

- **Platform organization**: the company operating the SaaS platform
- **Operator**: an authenticated platform staff member with an assigned role
- **Customer company**: the legal/business customer and commercial account
- **Tenant**: one isolated Odoo business environment owned by a customer company
- **Deployment**: the runtime database, container, hostname, storage, version, and health state for a tenant
- **Subscription**: the commercial plan, lifecycle, limits, and renewal state for a customer company or tenant
- **Contact**: customer owner, administrator, billing, and support contacts

Initially, one customer company may own one tenant, but the schema and API should allow one company to own multiple tenants later without duplicating customer identity or billing data.

## Phase 1: Architecture decisions

- [ ] Write an architecture decision record for the customer-company, tenant, deployment, subscription, contact, and operator boundaries
- [ ] Decide whether subscriptions attach to a customer company, individual tenant, or both
- [ ] Define operator roles: platform owner, operations, support, product/testing, and read-only
- [ ] Define which actions each role can perform and which actions require confirmation or elevated approval
- [ ] Define tenant lifecycle states and allowed transitions: provisioning, active, suspended, failed, deleting, and deleted
- [ ] Separate commercial status, desired lifecycle state, and observed runtime health instead of storing all meanings in one `state` string
- [ ] Define immutable identifiers separately from mutable company name, display name, and hostname
- [ ] Record single-host deployment as the current boundary and the abstractions required for a future scheduler

**Complete when:** an approved relationship diagram, state machine, role matrix, and decision record exist, with no ambiguous ownership fields.

## Phase 2: Central data model

- [ ] Add central models and migrations for customer companies, contacts, operators, role assignments, deployments, and subscriptions
- [ ] Make `Tenant` reference a customer company instead of representing both customer and environment
- [ ] Add database constraints for unique customer identity, tenant slug/hostname, database name, and active deployment
- [ ] Replace free-form release states and tenant states with validated constants and transition services
- [ ] Add created-by, updated-by, and correlation identifiers to lifecycle and audit records
- [ ] Store credential fingerprints and rotation metadata without serializing secrets
- [ ] Define deletion behavior for company, tenant, subscription, audit, and release records
- [ ] Add migration and rollback tests for the central schema

**Complete when:** the database enforces the relationship model and invalid ownership/state combinations cannot be persisted.

## Phase 3: Operator authentication and authorization

- [ ] Add secure operator login, logout, session expiry, password policy, and credential recovery
- [ ] Add authorization middleware to every superadmin HTTP and WebSocket route
- [ ] Protect tenant listing, creation, lifecycle, monitoring, audit, release listing, and deletion
- [ ] Replace the shared customization token with role-based operator authorization or a narrowly scoped deployment credential
- [ ] Add CSRF protection where cookie sessions are used
- [ ] Add rate limiting and failed-login controls
- [ ] Attribute every administrative audit event to an operator
- [ ] Add authorization tests for every role/action combination

**Complete when:** no administrative endpoint or WebSocket data is reachable anonymously and audit records identify the acting operator.

## Phase 4: Superadmin company workflow

- [ ] Create customer-company onboarding before tenant provisioning
- [ ] Capture legal/display name, primary contact, billing contact, support contact, locale, timezone, and default currency
- [ ] Show all tenants, subscriptions, contacts, and incidents from the customer-company view
- [ ] Create tenants from within a customer company and prevent orphan tenants
- [ ] Add operator-safe edit rules for company identity, tenant display name, and hostname
- [ ] Add a company/tenant handoff checklist for sales, product, and operations
- [ ] Add search and filtering across company, tenant, contact, state, and subscription
- [ ] Add audit history to the UI

**Complete when:** an operator can onboard a customer, provision its tenant, inspect ownership, and trace every change without direct database work.

## Phase 5: Durable provisioning and lifecycle

- [ ] Replace the in-process provisioning goroutine with a durable job model and worker
- [ ] Make every provisioning step idempotent and safely retryable
- [ ] Persist step status, attempt count, timestamps, failure reason, and recovery action
- [ ] Add compensating cleanup for partially created databases, directories, and containers
- [ ] Validate hostname, database, storage, image, and network availability before provisioning
- [ ] Reconcile desired tenant state with actual Docker/runtime state
- [ ] Make enable recreate a missing tenant container when the deployment record is valid
- [ ] Add deliberate suspend, resume, reprovision, and delete workflows
- [ ] Add typed confirmation and retention policy to destructive deletion
- [ ] Enforce subscription/expiry policy through the lifecycle service rather than a display-only date

**Complete when:** provisioning survives backend restart, retry does not duplicate resources, and every failure has a supported recovery path.

## Phase 6: Storage, secrets, and isolation

- [ ] Add persistent Odoo filestore storage per tenant
- [ ] Back up PostgreSQL, filestore, and tenant release files as one recoverable unit
- [ ] Test full tenant restoration and document RPO/RTO
- [ ] Move database, agent, API, and operator secrets to an external secret mechanism
- [ ] Add credential rotation for tenant API keys and management-agent tokens
- [ ] Use separate or least-privilege database credentials where practical
- [ ] Remove external PostgreSQL exposure from production configuration
- [ ] Add CPU, memory, process, and storage limits per tenant container
- [ ] Review tenant customization runtime permissions and filesystem mounts

**Complete when:** a tenant can be restored with attachments intact, secrets can rotate without code changes, and one tenant cannot consume unbounded host resources.

## Phase 7: Customization and platform releases

- [ ] Define separate release tracks for platform addons and tenant-owned addons
- [ ] Add operator identity, approval state, release notes, compatibility, and required Odoo version to releases
- [ ] Add checksum uniqueness and prevent duplicate or concurrent activation
- [ ] Add a per-tenant deployment lock
- [ ] Define safe database migration and rollback contracts for customer modules
- [ ] Add canary activation and post-activation health/business checks
- [ ] Build a controlled platform-addon rollout workflow for existing tenants
- [ ] Expose audit events and failure recovery in the superadmin UI
- [ ] Retain releases according to a documented storage policy

**Complete when:** releases are reviewable, serialized, observable, compatible, and recoverable without filesystem intervention.

## Phase 8: Observability and support

- [ ] Add health checks for backend, PostgreSQL, Traefik, documentation, and tenant containers
- [ ] Collect structured logs with tenant, company, operator, job, and request correlation IDs
- [ ] Add metrics for provisioning duration/failure, container health, database capacity, and customization outcomes
- [ ] Add alerts with owner, severity, acknowledgement, and escalation rules
- [ ] Create an incident record linked to customer company and tenant
- [ ] Add support-safe diagnostic views without exposing credentials
- [ ] Define log and audit retention
- [ ] Add operational runbooks for provisioning failure, tenant outage, database restore, and failed release

**Complete when:** operations can detect, diagnose, communicate, and recover from common failures using supported interfaces and runbooks.

## Phase 9: Product completion

- [ ] Validate initial import with representative customer workbooks and preserve import approval evidence
- [ ] Configure and test outgoing email and daily sales-report delivery
- [ ] Decide whether to finish or remove the Flutter mobile application
- [ ] If retained, route login to the selected tenant, issue user-scoped sessions, externalize environment configuration, and test logout/revocation
- [ ] Decide whether to finish or remove push notifications
- [ ] If retained, authenticate webhooks, install the addon intentionally, and integrate FCM/APNs
- [ ] Decide whether `shopping_portal` is supported, optional, or experimental
- [ ] Add billing only after the subscription ownership model is approved

**Complete when:** every user-visible surface is either supported and tested or explicitly excluded from the product.

## Phase 10: Verification and release readiness

- [ ] Add API contract tests, authorization tests, migration tests, and provisioning integration tests
- [ ] Add disposable-tenant end-to-end tests for create, monitor, suspend, resume, customize, rollback, backup, restore, and delete
- [ ] Add frontend tests for company onboarding and destructive confirmations
- [ ] Run tenant-isolation and customization security reviews
- [ ] Create staging and production configuration profiles
- [ ] Perform a canary customer onboarding rehearsal with sales, product, operations, and engineering
- [ ] Freeze a release candidate and record all image/module versions
- [ ] Obtain product, operations, security, and engineering sign-off

**Complete when:** the release candidate passes the full lifecycle rehearsal without undocumented manual repair.

## Tomorrow's recommended starting sequence

1. Draw the relationship diagram and approve the terminology.
2. Write the tenant lifecycle state machine.
3. Write the operator role/action matrix.
4. Turn those decisions into central model migrations and API boundaries.
5. Implement authentication before adding more superadmin capabilities.
