# Mobile API

The Go backend exposes a prototype mobile API under `/api/mobile`. Except for tenant discovery and login, endpoints accept a tenant API key in `X-API-Key` or `api_key` query parameter.

## Status warning

The stats and settings handlers are implemented. End-to-end mobile login is currently broken because it authenticates against `http://odoo:8069`, while the platform creates `odoo_tenant_<id>` containers. The returned API key is tenant-wide rather than user-scoped. Do not treat this API as production-ready authentication.

## Tenant discovery

`GET /api/mobile/tenants` requires no API key and returns `id`, `name`, and `subdomain` for active tenants.

## Login

`POST /api/mobile/login` accepts `tenant_id`, `email`, and `password`, intends to authenticate against Odoo JSON-RPC, and returns the tenant API key. Correct per-tenant routing and user-scoped session design are required.

## Sales dashboard

`GET /api/mobile/stats` requires the API key. It returns:

- Tenant name, subdomain, state, and expiry
- Confirmed sales totals/order counts for all time, current month, today, and yesterday
- Seven-day revenue/order trend for days with orders
- Five most recent orders
- Five products ranked by sold quantity

When `sale_order` is not yet initialized, it returns an initializing response instead of sales data.

## Device endpoints

- `POST /api/mobile/devices`: register `device_token` and `platform`
- `DELETE /api/mobile/devices/:token`: remove a token for the authenticated tenant
- `PUT /api/mobile/settings`: set `min_notification_amount`

These records do not result in real push delivery yet.

## Sale webhook

`POST /api/webhooks/odoo/sale` accepts database name, order identity, amount, and customer. When the amount reaches the tenant threshold, the backend finds registered devices and writes dummy push messages to logs.

The webhook has no authentication or signature. The `mobile_push_notifications` addon that calls it is not installed for new tenants by default.

## Flutter client

`mobile_app/` includes store selection, Odoo credential form, persisted API key, sales summary cards, weekly chart, recent orders, top products, pull-to-refresh, and notification-threshold settings. Its base URL is hardcoded to `http://superadmin.localhost:8090/api`.
