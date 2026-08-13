# Section 1: Introduction & SaaS Overview

## 1.1 Purpose of the Project

This project aims to develop a Software as a Service (SaaS) application for Stock Management. The system will allow multiple businesses (tenants) to manage their inventory, stock levels, product tracking, and potentially orders and suppliers, all from a centrally hosted platform. Each tenant will have a dedicated, isolated environment while benefiting from shared infrastructure and updates.

The primary goal is to provide a scalable, robust, and easy-to-use stock management solution that can be quickly provisioned for new clients, offering a cost-effective alternative to on-premise or custom-built systems.

## 1.2 Core SaaS Principles

The development of this Stock Management SaaS adheres to fundamental SaaS principles:

*   **Multi-Tenancy:** The architecture serves multiple tenants while isolating each tenant's data and custom code. Every tenant receives a separate PostgreSQL database, addon directory, and Odoo container managed by the central Go control plane.
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

1.  **Super Admin Control Plane:**
    *   **React frontend:** Provides tenant creation, lifecycle, monitoring, and customization controls.
    *   **Go backend:** Stores tenant metadata and performs provisioning through PostgreSQL and Docker APIs.

2.  **Tenant Instances:**
    *   **Odoo Tenant Containers:** Each active tenant runs in its own Odoo container with a dedicated database and tenant-addon mount.
    *   **Tenant Databases (`db` - PostgreSQL Service):** A PostgreSQL service that hosts the individual databases for each tenant (e.g., `tenant1_db`, `tenant2_db`). Each tenant's data is isolated within its own database.
        *   New tenant databases are created by the central Go provisioning API.
    *   **Configuration:** The Go backend creates initialization, maintenance, and long-running tenant containers through the Docker API.

3.  **Docker & Containerization:**
    *   **`docker-compose.yml`:** Defines and manages all services (Odoo instances, PostgreSQL databases), their configurations, volumes, ports, and networks.
    *   **`Dockerfile.odoo`:** Builds the Odoo 18 image used by tenant initialization, maintenance, and daemon containers.
    *   **Volumes:** PostgreSQL data uses a named volume. Tenant-specific addon releases use isolated directories under `tenants/`.

4.  **Web Server/Reverse Proxy:**
    *   Traefik routes the superadmin, documentation, and tenant hostnames to their containers. Production configuration must add managed DNS and TLS.

This architecture separates the control plane from tenant business applications while preserving per-company data and customization isolation.
