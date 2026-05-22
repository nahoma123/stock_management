# Project Backlog

This file contains a checklist of tasks to mature the Odoo SaaS platform from its current proof-of-concept state to a production-ready application.

### Phase 1: Core SaaS Management & Tenant Provisioning

- [ ] **Refactor Database Creation:** Replace `subprocess` calls in `saas_tenant.py` with a direct database connection library (e.g., `psycopg2`).
- [ ] **Secure Credential Management:** Remove hardcoded passwords from Python files and load them from environment variables.
- [ ] **Implement a Robust Job Queue:** Replace the `threading` model with Odoo's built-in job queue (`@job` decorator) for tenant creation.
- [ ] **Automate Multi-Company Setup:** Extend the tenant creation script to automatically enable and configure multi-company mode.
- [ ] **Develop Tenant Decommissioning Logic:** Create a feature to safely deactivate and delete tenants and their databases.
- [ ] **Improve Error Handling:** Enhance error logging and reporting for the tenant provisioning process.

### Phase 2: Networking & Infrastructure

- [ ] **Implement a Reverse Proxy:** Set up and configure a reverse proxy (e.g., Traefik, Nginx) to handle subdomain routing.
- [ ] **Automate SSL/TLS Certificates:** Configure the reverse proxy for automatic SSL certificate generation and renewal (e.g., with Let's Encrypt).
- [ ] **Create Production Docker Configuration:** Create a `docker-compose.prod.yml` file optimized for production (e.g., with logging drivers, resource limits).

### Phase 3: Point of Sale (POS) & Offline Capabilities

- [ ] **Configure and Test Offline POS:** Thoroughly test and document the setup for the offline mode of the Odoo POS module.
- [ ] **Package POS as a PWA:** Add a web app manifest and service worker to make the POS installable as a Progressive Web App.

### Phase 4: Mobile Owner App (New Feature)

- [ ] **Design Mobile API:** Define and document the API endpoints in Odoo for the mobile app.
- [ ] **Implement Mobile API:** Develop the API endpoints in a new custom Odoo module.
- [ ] **Develop Mobile App (iOS & Android):** Create the mobile application that consumes the new API.

### Phase 5: Documentation and Testing

- [ ] **Enhance Developer Documentation:** Update the project's documentation to reflect the new architecture and setup procedures.
- [ ] **Create Administrator's Guide:** Write a guide for system administrators on how to manage the SaaS platform.
- [ ] **Add Automated Tests:** Implement a suite of automated tests for the custom addons.
