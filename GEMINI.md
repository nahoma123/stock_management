# Gemini Project Context: Odoo SaaS Platform

This document provides a summary of the Odoo SaaS platform project to give context for future development sessions.

## 1. Project Overview

This project is a multi-tenant Odoo SaaS platform. It uses Docker for containerization and is designed to be a flexible and scalable solution for hosting multiple Odoo instances. The platform includes a superadmin instance for managing tenants and several custom Odoo addons to provide specific functionalities.

## 2. Technologies

*   **Backend:** Odoo (Python)
*   **Frontend:** Node.js, Vite, Tailwind CSS
*   **Database:** PostgreSQL
*   **Containerization:** Docker, Docker Compose

## 3. Directory Structure

*   `custom_addons/`: Contains custom Odoo modules.
    *   `boutique_theme/`: A custom theme for the Odoo backend.
    *   `saas_management_tools/`: Tools for managing SaaS tenants.
    *   `shopping_portal/`: A multi-instance shopping portal.
*   `docs/`: Project documentation.
*   `docker-compose.yml`: Defines the Docker services for the Odoo application, database, and superadmin.
*   `Dockerfile.odoo`: Dockerfile for the Odoo application.
*   `Dockerfile.backend`: Dockerfile for the Node.js backend.
*   `Dockerfile.frontend`: Dockerfile for the Node.js frontend.
*   `odoo.conf`: Configuration file for the Odoo tenant instances.
*   `superadmin.conf`: Configuration file for the Odoo superadmin instance.

## 4. Key Files

### `docker-compose.yml`

This file defines the following services:

*   **odoo:** The main Odoo application service for tenants.
*   **db:** The PostgreSQL database service for the tenant template.
*   **odoo_superadmin:** The Odoo superadmin service for managing tenants.
*   **db_superadmin:** The PostgreSQL database for the superadmin service.

### `odoo.conf` & `superadmin.conf`

These are the configuration files for the Odoo tenant and superadmin instances, respectively. They specify database connection details, addons paths, and other Odoo-specific settings.

### `__manifest__.py` Files

These files are present in each custom addon directory and define the addon's metadata, dependencies, and data files.

## 5. Custom Addons

### `boutique_theme`

*   **Purpose:** Provides a custom theme and UI for the Odoo backend.
*   **Dependencies:** `base`, `web`, `stock`, `point_of_sale`, `crm`
*   **Key Files:** `views/web_layout.xml`, `views/menu_items.xml`, `static/src/scss/theme.scss`

### `saas_management_tools`

*   **Purpose:** Provides tools for creating and managing SaaS tenants.
*   **Dependencies:** `base`
*   **Key Files:** `views/saas_tenant_views.xml`, `wizards/tenant_creation_wizard.py`

### `shopping_portal`

*   **Purpose:** A multi-instance shopping portal with user registration and product aggregation.
*   **Dependencies:** `base`, `web`, `website`, `auth_signup`, `portal`, `sale`, `product`
*   **Key Files:** `views/portal_templates.xml`, `views/registration_templates.xml`, `static/src/js/portal.js`

## 6. Documentation

The `docs/` directory contains detailed documentation for the project, including setup guides, deployment instructions, and information on extending the platform. The main entry point for the documentation is `docs/README.md`.

## 7. Odoo Module Structure

The following is the standard structure for an Odoo module, as found in `module_info.txt`:

```
my_module/
├── __init__.py           # Python package initialization
├── __manifest__.py       # Module manifest file (metadata)
├── models/               # Data models
│   ├── __init__.py
│   └── models.py
├── views/                # XML view definitions
│   └── views.xml
├── security/            # Access rights and rules
│   └── ir.model.access.csv
├── static/               # Static assets (CSS, JS, images)
│   ├── description/
│   │   └── icon.png
│   └── src/
└── data/                 # Demo data and default configurations
    └── demo.xml
```
