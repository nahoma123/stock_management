# Section 5: Super Admin Guide

## 5.1 Introduction

The Super Admin interface is a dedicated Odoo instance designed for the overall management of the SaaS platform and its tenants. It provides tools to create new tenant instances, monitor their status, and manage their lifecycle. The primary custom tool within this interface is the "SaaS Management Tools" module.

## 5.2 Accessing the Super Admin

*   **Local Development URL:** `http://localhost:18070`
    *   (As defined by the port mapping for the `odoo_superadmin` service in `docker-compose.yml`).
*   **Production URL:** `https://superadmin.yourdomain.com` (or your configured production URL).
*   **Login Credentials:**
    *   Upon first setup (if the database for `db_superadmin` is new), Odoo may prompt for master password creation (this is the `admin_passwd` from `superadmin.conf`).
    *   The default Odoo administrator login is typically:
        *   **Email:** `admin`
        *   **Password:** `admin`
    *   **It is crucial to change this default password immediately after the first login, especially in production.**

## 5.3 Super Admin Dashboard (`saas_management_tools`)

After logging into the Super Admin interface and ensuring the "SaaS Management Tools" module is installed (see Section 3.9), you will find a "SaaS Management" menu item in the main Odoo application list.

### 5.3.1 Overview of the "SaaS Management" Menu
*   **Dashboard:** Provides an at-a-glance summary of tenant statistics and quick actions.
*   **Tenants:** A list view of all managed tenants, allowing access to their detailed records.
*   **Create New Tenant:** A shortcut to launch the tenant creation wizard.

### 5.3.2 Explanation of the Dashboard View
The dashboard is the default view when you click on the "SaaS Management" menu.
*   **Summary Statistics:**
    *   A text block displays key metrics about the tenant population:
        *   `Total Tenants`: The total number of tenant records in the system.
        *   `Active`: Number of tenants currently in the "Active" state.
        *   `Error`: Number of tenants that encountered an error during creation or other processes.
        *   `Draft`: Number of tenants created but whose database provisioning has not yet been initiated.
        *   `Creating`: Number of tenants currently undergoing the database provisioning process.
*   **Shortcuts:**
    *   **Create New Tenant:** A button that directly opens the tenant creation wizard.
    *   **View All Tenants:** A button that navigates to the tree view listing all tenants.

## 5.4 Managing Tenants

### 5.4.1 Creating a New Tenant
1.  **Navigation:** Click on "Create New Tenant" from the dashboard or the "SaaS Management" menu.
2.  **Tenant Creation Wizard:** A pop-up form will appear with the following fields:
    *   **Shop Name:** Enter a descriptive name for the tenant (e.g., "Green Valley Groceries", "Main Street Boutique"). This is for display purposes.
    *   **Requested Subdomain:** Enter the desired subdomain for the tenant (e.g., `greenvalley`, `mainstboutique`).
        *   This subdomain will be used to construct the tenant's unique URL (e.g., `greenvalley.yourshopsdomain.com`).
        *   It will also be used to generate the tenant's database name (e.g., `greenvalley_db`).
        *   Keep it short, alphanumeric, and unique. The system will check for uniqueness against existing `saas.tenant` records.
3.  **Provisioning Process:**
    *   After clicking "Create Tenant" in the wizard, a new `saas.tenant` record is created in the "Draft" state.
    *   You will be redirected to the form view of this new tenant record.
    *   To initiate the actual database creation and Odoo initialization, click the **"Create Database"** button (visible when the tenant is in "Draft" state).
    *   The tenant's state will change:
        *   `Draft` -> `Creating`: The system is now attempting to create the PostgreSQL database and initialize it with the base Odoo modules. This is a background task.
        *   `Creating` -> `Active`: If the process completes successfully.
        *   `Creating` -> `Error`: If any step in the provisioning fails.
    *   You can monitor the progress via the `creation_log` field on the tenant's form view and the `odoo_superadmin` container logs.
    *   **Remember:** For the new tenant to be accessible via its subdomain in a local development environment, you must add an entry to your `/etc/hosts` file (e.g., `127.0.0.1 greenvalley.localhost`). In production, DNS (wildcard A record) should handle this.

### 5.4.2 Viewing Tenant List
*   Navigate to `SaaS Management > Tenants`.
*   This view displays a list (tree view) of all tenant records with key information:
    *   `Name`: The descriptive name of the tenant.
    *   `Subdomain`: The unique subdomain assigned.
    *   `DB Name`: The computed database name.
    *   `State`: Current status (Draft, Creating, Active, Error).
    *   `Create Date`: Date and time when the tenant record was created in the superadmin system.
    *   `License Expiry Date`: The configured license expiry for the tenant.
*   **Filtering and Searching:** Standard Odoo search and filter capabilities can be used to find specific tenants based on any of these fields.

### 5.4.3 Viewing Tenant Details (Form View)
Clicking on any tenant record in the list view opens its detailed form view.
*   **Fields Explained:**
    *   `Name`: Tenant's descriptive name.
    *   `Subdomain`: Tenant's unique subdomain.
    *   `DB Name`: (Read-only) The computed name of the tenant's PostgreSQL database.
    *   `State`: (Read-only, except for actions) Current status, with a status bar for visual tracking.
    *   `Create Date`: (Read-only) Date and time the tenant record was created.
    *   `License Expiry Date`: Date field to manage tenant license validity.
    *   `Internal Notes`: A text field for administrators to keep internal notes about the tenant.
    *   `Creation Log`: (Read-only) A log of messages generated during the tenant's database creation and initialization process. This is crucial for tracking progress and diagnosing failures. It will show messages like "Starting database creation process...", "Database ... created successfully.", "Initializing Odoo in database ...", or any error messages encountered.

### 5.4.4 Modifying Tenant Information
Some tenant information can be modified directly from the form view:
*   **Editable Fields:**
    *   `Tenant Name`: Can be changed as needed.
    *   `Subdomain`: **Caution:** Changing the subdomain after a tenant's database has been created will **not** automatically rename the database or update URLs without further custom logic (which is not implemented in the current version). This could lead to a mismatch and make the tenant inaccessible if not handled carefully. Currently, it's best to set the subdomain correctly at creation.
    *   `License Expiry Date`: Can be updated to reflect new subscription periods.
    *   `Internal Notes`: Can be updated at any time.
*   **Re-triggering Creation:** The "Create Database" button is only available when the tenant is in the "Draft" state. There is no built-in mechanism to automatically "re-provision" an existing database through this button. If a tenant is in an "Error" state, manual intervention (checking logs, potentially dropping the faulty database, and then possibly resetting the state to "Draft" via Odoo's server actions or direct database manipulation if necessary) would be required before attempting creation again.

### 5.4.5 (Placeholder for Future) Managing Tenant Lifecycle
The current version of `saas_management_tools` focuses on tenant creation. A production-ready system would require additional features for managing the full tenant lifecycle:
*   **Suspending a Tenant:** Temporarily disabling access to a tenant's instance (e.g., for non-payment). This might involve:
    *   Changing the tenant's state to "Suspended".
    *   Potentially adding Nginx rules or other mechanisms to block access to their subdomain.
    *   Stopping tenant-specific workers or processes (if applicable in a more advanced setup).
*   **Reactivating a Tenant:** Reversing a suspension.
*   **Deleting a Tenant:** Permanently removing a tenant and their data. This is a critical operation and should involve:
    *   Changing the state to "To Be Deleted".
    *   A grace period or confirmation step.
    *   Dropping the tenant's PostgreSQL database.
    *   Archiving or deleting the tenant's filestore from the `odoo-data` volume.
    *   Removing associated configurations (e.g., Nginx if specific rules were added).

## 5.5 Troubleshooting Tenant Issues (from Super Admin perspective)

1.  **Check `creation_log`:** The primary source for diagnosing issues during tenant provisioning is the `Creation Log` field on the tenant's form view in the Super Admin interface. It often contains the direct error message from `createdb` or `odoo --init` commands.
2.  **Check `odoo_superadmin` Container Logs:** For more detailed stack traces or system-level errors that might not be fully captured in the `creation_log`, view the Docker logs for the `odoo_superadmin` service:
    ```bash
    docker-compose logs -f odoo_superadmin
    ```
    (In production, these logs might be aggregated elsewhere).
    Look for Python tracebacks or error messages around the time the tenant creation was attempted. Common issues include:
    *   Incorrect database credentials for the tenant PostgreSQL service (`db`) being used by the `createdb` or `odoo --init` commands (check `saas_tenant.py`).
    *   `postgresql-client` not found (if `Dockerfile.odoo` was not set up correctly).
    *   Network issues preventing `odoo_superadmin` from reaching the `db` service.
    *   Problems with `admin_passwd` in `superadmin.conf` preventing Odoo from authorizing database operations.
    *   Insufficient permissions for the PostgreSQL user specified for tenant DB creation.

## 5.6 Customizing the Super Admin Interface (Brief)

*   The Super Admin panel is a fully functional Odoo instance. Therefore, advanced customizations can be made by developing or installing additional Odoo modules specifically within its environment.
*   This could include new models for managing billing, support tickets, more detailed analytics, etc.
*   However, the core tenant provisioning and management functionalities described in this guide are primarily handled by the `saas_management_tools` module. Any direct enhancements to tenant lifecycle management would likely involve modifying this module.

This guide should provide a solid understanding of how to use the Super Admin interface to manage tenants within the SaaS platform.
