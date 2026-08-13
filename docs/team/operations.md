# Operations

Use this guide for tenant onboarding, routine monitoring, access control, and first response. Escalate implementation changes to developers through the Product & Testing handoff.

## Create a tenant

1. Open the superadmin dashboard.
2. Enter the company name and approved lowercase subdomain.
3. Create the tenant and follow its creation log.
4. Wait for the state to become active.
5. Open monitoring and confirm the management agent responds.
6. Verify the customer URL and administrator login before handoff.

Do not create a second tenant to hide a failed creation. Capture the log and resolve or delete the failed tenant deliberately.

## Initial customer data

Obtain the approved CSV or XLSX workbook and retain an unchanged copy. Confirm column meaning, currency, product identifiers, warehouse mapping, and the effective opening-stock date with the customer. Run a preview, resolve validation errors, and have the customer or product owner approve totals.

## Routine checks

Check tenant state, monitoring availability, recent provisioning or customization audit events, license expiry, and reported user access. A healthy container alone does not prove the business workflow works.

## Customization release

Only operator-reviewed Odoo module packages may be uploaded. Validate, stage, and activate a release through the superadmin workflow. Confirm monitoring and the affected customer workflow afterward. Use rollback when the new release fails its operational check.

## Incident first response

Record the tenant, start time, affected users, visible error, recent changes, and business impact. Check the central tenant state and creation or audit logs before container logs. Avoid direct database edits or unmanaged addon files. Escalate security, data isolation, data loss, repeated provisioning failures, and failed rollback immediately.

## Customer handoff

Provide the customer URL, approved administrator contact, enabled scope, import outcome, known limitations, and support path. Never send platform operator tokens or management-agent credentials.
