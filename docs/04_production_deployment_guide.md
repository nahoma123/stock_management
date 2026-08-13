# Production Readiness

The current Compose deployment is a development/initial managed-deployment foundation, not a production-ready reference architecture.

## Blockers before customer production

- Add authentication and authorization to the superadmin UI and all control-plane lifecycle endpoints
- Remove default database credentials and external database exposure
- Persist and back up Odoo filestores as well as PostgreSQL and tenant releases
- Authenticate internal sale webhooks
- Add resource limits, health checks, logging, metrics, and alerting
- Configure production DNS, HTTPS, and restricted network access
- Test tenant provisioning, upgrade, rollback, backup, and restore procedures
- Decide whether to finish or disable mobile login and push-notification surfaces

## Secrets

Keep `.env`, database credentials, tenant API keys, agent tokens, operator tokens, and customer packages outside Git. `CUSTOMIZATION_ADMIN_TOKEN` is necessary but does not replace operator authentication for the rest of the control plane.

## Backups

Back up PostgreSQL and `tenants/` consistently. Add persistent tenant filestore mounts before relying on Odoo attachments, imported images, or documents, then include those mounts in backups. A backup is not accepted until restoration into a disposable environment has been tested.

## Deployment sequence

1. Build and test backend, frontend, documentation, and Odoo images.
2. Back up current data and releases.
3. Apply central database migrations through backend startup.
4. Replace control-plane services.
5. Upgrade affected platform modules in tenant databases through a tested procedure.
6. Validate monitoring and primary business flows on a canary tenant.
7. Roll out to remaining tenants and retain a recovery plan.

The current project does not automate step 5 across all tenants.

## Customer customization releases

Use validate, stage, and activate. Activation stops the tenant, switches the module release, runs Odoo maintenance, and restarts it. Verify monitoring and the affected workflow. Rollback restores a previous code release and reruns maintenance, but it cannot guarantee reversal of every module-specific data migration; module authors must design reversible upgrades.
