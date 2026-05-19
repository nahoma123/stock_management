# Gemini Project Context: Odoo SaaS Platform

This document provides a summary of the Odoo SaaS platform project to give context for future development sessions.

## 1. Project Overview

This project is a multi-tenant Odoo SaaS platform. It uses Docker for containerization and is designed to be a flexible and scalable solution for hosting multiple Odoo 17 instances. The platform includes a superadmin instance for managing tenants and several custom Odoo addons to provide specific functionalities.

## 2. Technologies

*   **Odoo Backend:** Odoo (Python), version 17.
*   **Superadmin Backend:** Go, Gin Framework, Gorilla WebSocket
*   **Superadmin Frontend:** React, Vite
*   **Database:** PostgreSQL
*   **Containerization:** Docker, Docker Compose

## 3. Directory Structure

*   `custom_addons/`: Contains custom Odoo modules.
    *   `boutique_theme/`: A custom theme for the Odoo backend.
    *   `saas_management_tools/`: Tools for managing SaaS tenants.
    *   `shopping_portal/`: A multi-instance shopping portal.
*   `backend/`: Contains the Go backend for the superadmin panel.
*   `frontend/`: Contains the React frontend for the superadmin panel.
*   `docs/`: Project documentation.
*   `docker-compose.yml`: Defines the Docker services for the Odoo application, database, and superadmin.
*   `Dockerfile.odoo`: Dockerfile for the Odoo application.
*   `Dockerfile.backend`: Dockerfile for the Go backend.
*   `Dockerfile.frontend`: Dockerfile for the React frontend.
*   `odoo.conf`: Configuration file for the Odoo tenant instances.

## 4. Key Files

### `docker-compose.yml`

This file defines the following services:

*   **odoo:** The main Odoo application service for tenants.
*   **db:** The PostgreSQL database service for the tenant template.
*   **backend:** The Go backend service for the superadmin panel.
*   **frontend:** The React frontend service for the superadmin panel.

### `odoo.conf`

This is the configuration file for the Odoo tenant instances. It specifies database connection details, addons paths, and other Odoo-specific settings.

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

---

## 8. Key Learnings & Developer Notes

*   **Odoo 17 View Types:** The `tree` view type has been renamed to `list`. When creating list views in XML for Odoo 17+, the root tag of the view's `arch` must be `<list>` and the `view_mode` in the corresponding `ir.actions.act_window` must use `'list'`. Using the legacy `<tree>` tag will cause a `ValueError: Wrong value for ir.ui.view.type: 'tree'` during module installation or upgrade.
*   **Odoo 17 View Attributes:** The `attrs` attribute is deprecated for dynamically changing view element properties. Use the `invisible` attribute with a domain directly on the element instead (e.g., `invisible="[('state', '!=', 'draft')]"`).
*   **Odoo XML Loading Order:** Odoo parses data files in the order they are listed in `__manifest__.py`. If an XML record refers to an ID in another file (e.g., an action using a `view_id`), the file containing the referenced ID must be listed first.
*   **Odoo Caching:** When making changes to `__manifest__.py` or other non-Python files, a simple browser refresh is not enough. You must either use the "Update Apps List" feature or, more reliably, restart the Odoo server container to ensure changes are loaded.

## 9. Recent Activity Log (What happened last time?)

*   **When:** September 15, 2025
*   **What:** Updated the project documentation to reflect the new superadmin architecture.
*   **Why:** The previous documentation was outdated and referred to the old Odoo-based superadmin.
*   **How:**
    1.  Updated the "Technologies" section to include the new Go backend and React frontend.
    2.  Updated the "Directory Structure" section to include the new `backend` and `frontend` directories.
    3.  Updated the "Key Files" section to reflect the new `docker-compose.yml` services.
    4.  Added a new entry to the "Recent Activity Log" to summarize the architectural changes.

*   **When:** August 15, 2025
*   **What:** Successfully debugged and installed the `saas_management_tools` custom Odoo module.
*   **Why it was failing:** The module was written with syntax from older Odoo versions, causing multiple compatibility errors with the project's Odoo 17 environment.
*   **How it was solved:** We performed an iterative debugging session:
    1.  Isolated the initial, misleading `ValueError` by commenting out the tree view.
    2.  Fixed a deprecated `attrs` attribute in the form view.
    3.  Fixed multiple "External ID not found" and `FileNotFoundError` errors by correcting the file loading order and file paths in `__manifest__.py`.
    4.  Finally diagnosed the root `ValueError` as being caused by the `tree` view tag, which was corrected to `list`.