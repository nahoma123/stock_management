# Security and Limitations

## Control-plane access

Tenant listing, creation, enable, disable, expiry, deletion, monitoring, audit listing, and release listing currently have no operator authentication middleware. Only customization validation and release mutations compare `CUSTOMIZATION_ADMIN_TOKEN`.

The superadmin must remain behind trusted network access until authentication and authorization are implemented.

## Credential scope

- Tenant API keys are serialized to the React dashboard and used for mobile stats/settings/device endpoints.
- Mobile login returns the tenant-wide API key after Odoo authentication; it is not a user-scoped token.
- Management-agent tokens are private and excluded from JSON responses.
- The operator customization token is retained in browser session storage.

## Internal webhook

`POST /api/webhooks/odoo/sale` trusts the supplied database name and has no secret or signature. Do not expose it to untrusted networks. Push delivery is currently dummy logging only.

## Custom code

Customization ZIP validation blocks common archive hazards and a few obvious dangerous strings. It does not sandbox Python or prove code safety. A tenant module can access that tenant’s Odoo process, database permissions, and mounted paths.

## Data and persistence

Tenant databases are separate but share one PostgreSQL server and credentials. Dynamic tenant containers do not currently declare persistent filestore volumes. Imported images and other attachments therefore require a persistence design before production reliance.

## Lifecycle limitations

- Expiry is informational and not automatically enforced.
- Enable starts an existing container; it does not repair or recreate a missing one.
- Provisioning is asynchronous but not a durable job queue.
- A failed database creation can leave resources requiring operator cleanup.
- Generic customization rollback cannot undo every data migration.

## Mobile limitations

Mobile authentication currently targets a nonexistent shared `odoo` hostname instead of `odoo_tenant_<id>`. The app base URL is hardcoded to local development. Device registration exists, but no production push provider is connected.
