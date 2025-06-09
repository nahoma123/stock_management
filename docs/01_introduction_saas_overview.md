# Section 1: Introduction & SaaS Overview

## 1.1 Purpose of the Project

This project aims to develop a Software as a Service (SaaS) application for Stock Management. The system will allow multiple businesses (tenants) to manage their inventory, stock levels, product tracking, and potentially orders and suppliers, all from a centrally hosted platform. Each tenant will have a dedicated, isolated environment while benefiting from shared infrastructure and updates.

The primary goal is to provide a scalable, robust, and easy-to-use stock management solution that can be quickly provisioned for new clients, offering a cost-effective alternative to on-premise or custom-built systems.

## 1.2 Core SaaS Principles

The development of this Stock Management SaaS adheres to fundamental SaaS principles:

*   **Multi-Tenancy:** The architecture is designed to serve multiple tenants (customers/businesses) from a single instance of the application. While the application instance is shared, each tenant's data is isolated and secured from other tenants. In our Odoo-based implementation, this is achieved by provisioning separate databases for each tenant, all managed by a common Odoo application server (for tenant-facing operations) and a super-admin Odoo instance for administrative tasks.
*   **Central Management & Provisioning:** A central super-administrator interface is a core component. This interface is responsible for:
    *   Tenant lifecycle management (creation, suspension, deletion).
    *   Monitoring tenant health and resource usage.
    *   Managing master configurations and application updates.
    *   Automated or semi-automated provisioning of new tenant instances.
*   **Scalability:** The system is designed with scalability in mind, leveraging Docker to allow for horizontal scaling of application servers and database services as the number of tenants and their load increases.
*   **Customization (where applicable):** While the core application is shared, tenants may have options for certain customizations (e.g., themes, specific workflows if supported by the underlying Odoo modules) without affecting the core codebase.
*   **Subscription-Based Model:** (Implied) SaaS solutions are typically offered on a subscription basis, though the implementation of billing and subscription management is outside the scope of the current technical tasks but is a consideration for a production system.

## 1.3 High-Level Architecture

The system employs a tiered architecture, containerized using Docker for deployment and scalability:

1.  **Super Admin Instance:**
    *   **Odoo Super Admin Service (`odoo_superadmin`):** An Odoo instance dedicated to administrative functions. This is where the `saas_management_tools` module resides, allowing administrators to create and manage tenants.
    *   **Super Admin Database (`db_superadmin`):** A PostgreSQL database exclusively for the `odoo_superadmin` service. It stores information about tenants (metadata, state, configuration) but not the tenants' actual business data.
    *   **Configuration:** Uses `superadmin.conf` for its Odoo settings.

2.  **Tenant Instances (Template & Actual):**
    *   **Odoo Tenant Service (`odoo`):** The primary Odoo application server that will handle requests for all active tenant instances. It uses `dbfilter_format = %d_db` in its `odoo.conf` to dynamically select the correct tenant database based on the request's subdomain.
    *   **Tenant Databases (`db` - PostgreSQL Service):** A PostgreSQL service that hosts the individual databases for each tenant (e.g., `tenant1_db`, `tenant2_db`). Each tenant's data is isolated within its own database.
        *   New tenant databases are created on this PostgreSQL service by the `saas_management_tools` module running in the `odoo_superadmin` instance.
    *   **Configuration:** The `odoo` service uses `odoo.conf` which is configured to serve as a template and routing mechanism for tenant databases.

3.  **Docker & Containerization:**
    *   **`docker-compose.yml`:** Defines and manages all services (Odoo instances, PostgreSQL databases), their configurations, volumes, ports, and networks.
    *   **`Dockerfile.odoo`:** A custom Dockerfile that builds upon the official `odoo:18.0` image to include necessary dependencies like `postgresql-client` (used by the superadmin for creating tenant databases). Both `odoo` and `odoo_superadmin` services use the image built from this Dockerfile.
    *   **Volumes:** Persistent storage for Odoo application data (filestores) and PostgreSQL databases is managed using Docker named volumes (e.g., `odoo-data`, `postgres-data`, `odoo-superadmin-data`, `postgres-superadmin-data`).

4.  **Web Server/Reverse Proxy (Conceptual - Not yet implemented):**
    *   In a production environment, a reverse proxy (e.g., Nginx, Traefik) would be placed in front of the Odoo services. It would handle SSL termination, routing requests based on subdomains (e.g., `tenant1.yourdomain.com` to the `odoo` service), and potentially load balancing. This component is not part of the current setup but is a crucial part of a complete SaaS architecture.

This architecture allows for a separation of concerns: the superadmin manages the platform and tenants, while the main Odoo service, in conjunction with the tenant PostgreSQL service, delivers the stock management application to each tenant in an isolated manner.
