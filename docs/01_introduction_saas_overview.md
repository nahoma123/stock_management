# System Overview

## Purpose

The Stock Platform is a managed, multi-company inventory and sales platform built on Odoo 18. A React/Go control plane provisions a separate Odoo container, PostgreSQL database, and tenant-addon directory for every customer company.

## Runtime architecture

The checked-in Compose stack contains:

- `db`: PostgreSQL 16 for control-plane tables and tenant databases
- `backend`: Gin/GORM API for tenant lifecycle, monitoring, mobile data, and customization releases
- `frontend`: React superadmin served by Nginx
- `documentation`: this Markdown-driven documentation site
- `traefik`: local hostname routing on host port `8090`

Tenant Odoo containers are created dynamically through the Docker socket and are not static Compose services. Their names are `odoo_tenant_<id>`. Initialization and module maintenance use transient containers.

## Tenant isolation

Each tenant has:

- A database named from its subdomain, such as `acme_db`
- A long-running Odoo container
- A tenant-owned addon directory under `tenants/<subdomain>/custom_addons`
- A private management-agent token
- A tenant API key used by the mobile API

Shared platform addons are mounted read-only at `/mnt/platform-addons`. Tenant releases are mounted at `/mnt/tenant-addons`. Odoo packaged modules form the core layer.

## Provisioned applications

New tenants initialize Odoo `base`, `web`, Sales, Inventory, Daily Sales Report, Initial Product Data Import, and Tenant Management Agent. The exact list is defined in `DockerCreateAndStartInitContainer` and appears automatically on the documentation System Map.

`shopping_portal` and `mobile_push_notifications` exist in the repository but are not installed by default.

## Control plane

The React dashboard can create, search, enable, disable, monitor, customize, and set expiry for tenants. Creation progress arrives over WebSocket. The Go API also has a tenant-deletion endpoint, but the current React table does not expose a delete button.

Only customization mutation endpoints use `CUSTOMIZATION_ADMIN_TOKEN`. The remaining control-plane endpoints currently have no operator authentication; see [Security and Limitations](./security_and_limitations.md).

## Current status

Use [Feature Status](./feature_status.md) as the concise source of truth for available, prototype, and incomplete capabilities.
