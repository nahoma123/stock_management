# Section 7: Extending the Platform

## 7.1 Introduction

This SaaS platform is designed to be extensible, allowing for growth in terms of the number of tenants it supports and the features it offers. This section outlines various ways the platform can be extended and considerations for doing so.

## 7.2 Adding More Tenants

The primary method of expansion is onboarding new tenants.

*   **Process Recap:** New tenants are added using the "Create New Tenant" wizard found in the Super Admin interface (`SaaS Management > Create New Tenant`). This process provisions a new database and initializes it with the defined set of Odoo modules.

*   **Scalability Considerations:**
    *   **Server Resources:** Each new tenant consumes additional server resources:
        *   **Disk Space:** For their database and Odoo filestore (attachments).
        *   **CPU & RAM:** When users from the tenant are active, they utilize CPU and RAM on the `odoo` application server and the `db` PostgreSQL server.
    *   **Database Server:** With a large number of tenants or very active tenants, the single PostgreSQL service (`db`) might become a bottleneck. Advanced strategies like database server optimization, using a more powerful database server, or exploring PostgreSQL replication/clustering might be necessary in the future.
    *   **Odoo Application Server:** The number of Odoo workers configured for the `odoo` service (`workers` in `odoo.conf`) needs to be adequate for the total number of concurrent users across all tenants. Monitor CPU and memory usage and adjust worker count or consider horizontal scaling of the `odoo` service (more advanced, requiring a load balancer).
    *   Regular monitoring of resource utilization is key to understanding when scaling is needed.

## 7.3 Adding New Features (Developing New Custom Odoo Modules)

The platform's functionality can be enhanced by developing new custom Odoo modules.

### 7.3.1 Workflow Recap
Developing a new Odoo module typically involves:
1.  Creating the module's directory structure (e.g., `custom_addons/my_new_module/`).
2.  Defining Python models (`models/`), XML views (`views/`), security rules (`security/ir.model.access.csv`), menu items, etc.
3.  Writing business logic in Python methods.
4.  Creating QWeb templates for user interfaces.
5.  Thoroughly testing the module in a development environment (ideally a staging tenant).
    *(Refer to Odoo's official developer documentation for detailed module development guides.)*

### 7.3.2 Integrating with SaaS Provisioning
To ensure a new custom module is automatically installed for all newly created tenants:
1.  Identify the technical name of your new module (e.g., `my_new_module`).
2.  Edit the `saas_management_tools/models/saas_tenant.py` file.
3.  Locate the `_create_and_initialize_tenant_db` method.
4.  Add your module's technical name to the `modules_to_install` string or list. For example:
    ```python
    # Inside _create_and_initialize_tenant_db method
    # modules_to_install = 'base,web,boutique_theme,shopping_portal'
    modules_to_install = 'base,web,boutique_theme,shopping_portal,my_new_module'
    ```
5.  Deploy this change to your superadmin Odoo instance. Subsequent tenant provisioning will include this new module.

### 7.3.3 Rolling Out to Existing Tenants
Once a new module is developed and added to the codebase:
*   **Code Deployment:** Deploy the updated `custom_addons` directory to the server.
*   **Restart Services:** Restart the Odoo Docker containers (`docker-compose up -d --no-deps odoo odoo_superadmin` - the `--no-deps` is optional but can speed up if only app servers are changed).
*   **Making Module Available to Existing Tenants:**
    *   **Option 1 (Manual by Tenant Admin or Super Admin):**
        An administrator can log into each existing tenant's Odoo instance, go to the "Apps" menu, search for the new module, and click "Install."
    *   **Option 2 (Super Admin Initiated Batch Install - Future Feature):**
        A more advanced approach (not yet implemented) would be to develop a utility within the `saas_management_tools` module. This utility could:
        *   List all active tenants.
        *   Allow the super admin to select a new module to be installed.
        *   Iterate through the selected tenant databases and programmatically trigger the installation of the new module. This would require careful error handling and potentially background job processing.
*   **Considerations:**
    *   **Dependencies:** If the new module has dependencies, ensure they are also available and listed in its manifest.
    *   **Data Migrations:** If the new module alters existing data structures or requires initial data setup, include appropriate migration scripts or post-init hooks in your module.

## 7.4 Mobile App Integration (Recap)

Odoo's architecture is inherently API-friendly, facilitating mobile app integration.

*   **API First:** Odoo's primary interface is its web interface, but all operations are accessible via its JSON-RPC API (and XML-RPC, though JSON-RPC is generally preferred for modern apps).
*   **Progressive Web App (PWA):** Odoo's web interface is responsive. For a simpler mobile experience, PWAs can be a good starting point, allowing users to "install" the web app to their home screen.
*   **Native/Cross-Platform Development:** For richer mobile experiences:
    *   Develop native (iOS/Android) or cross-platform (e.g., React Native, Flutter, .NET MAUI) applications.
    *   These applications would interact with Odoo as a backend data source and service provider via its JSON-RPC API.
*   **Tenant-Specific API Endpoint:** Each tenant has its own isolated data and thus its own API endpoint:
    `https://<tenant_subdomain>.yourshopsdomain.com/jsonrpc`
    Authentication would be handled using Odoo user credentials for that specific tenant database.

## 7.5 Integrating with Third-Party Services

Many business applications benefit from integrating with external services.

*   **Examples:**
    *   **Payment Gateways:** (e.g., Stripe, PayPal). Odoo often has existing modules for popular gateways. If not, a custom one can be developed.
    *   **Shipping Providers:** (e.g., FedEx, UPS, DHL). Similar to payment gateways, connectors might exist or can be built.
    *   **External SMS Gateways:** As envisioned by the `sms_reporting` example.
    *   **Accounting Software, CRMs, Marketing Automation Tools.**
*   **General Approach:**
    1.  **Identify Need & API:** Determine the third-party service and understand its API capabilities.
    2.  **Develop Custom Odoo Module:** Create a new module (e.g., `payment_stripe_integration`, `sms_twilio_gateway`) in `custom_addons/`.
    3.  **Secure Configuration:**
        *   Store API keys, tokens, and other sensitive configuration data securely. Odoo's system parameters (`ir.config_parameter`) can be used, accessible via the Settings menu. Consider encrypting these values if Odoo's native storage isn't deemed sufficient.
        *   Provide configuration options within the module's settings in Odoo.
    4.  **Implement API Interaction:** The module's Python code will use libraries like `requests` to make API calls to the third-party service.
    5.  **Provide UI/Logic:** Expose functionality through Odoo views, models, and wizards as needed.
    6.  **Installation:** This integration module would then be made available for installation in tenant instances that require it (either during provisioning or installed later).

## 7.6 Updating Odoo Core and Custom Modules

Keeping the platform and its features up-to-date is crucial for security, performance, and access to new functionalities.

### 7.6.1 Odoo Core Updates (Minor/Patch Versions)
Odoo releases regular updates. For a specific major version (e.g., 18.0), these are typically patch releases or minor updates.
1.  **Update Base Docker Image:**
    *   Modify `Dockerfile.odoo` to point to a newer build of the `odoo:18.0` image if specific dated tags are used (e.g., `FROM odoo:18.0.20240301`).
    *   Alternatively, if you're using a floating tag like `odoo:18.0`, you can explicitly pull the latest version of that tag: `docker pull odoo:18.0`.
2.  **Rebuild Custom Odoo Image:**
    ```bash
    docker-compose build odoo odoo_superadmin
    ```
3.  **Restart Services:**
    ```bash
    docker-compose up -d
    ```
    This will stop the old containers and start new ones based on the updated image.
4.  **Thorough Testing:** **Crucial.** After any core update, thoroughly test all aspects of the platform, including superadmin functions and tenant operations, in a staging environment before applying to production. Pay attention to release notes for any breaking changes, though these are less common in patch releases.

### 7.6.2 Custom Module Updates
When you develop new versions of your custom modules in `custom_addons/`:
1.  **Deploy Updated Code:**
    *   Push changes to your Git repository.
    *   On the server, pull the latest code: `git pull`.
2.  **Restart Odoo Services:**
    ```bash
    docker-compose up -d --no-deps odoo odoo_superadmin
    ```
    This ensures Odoo's Python interpreter picks up any code changes.
3.  **Upgrade Modules in Odoo Instances:**
    *   **For each affected tenant instance (and the superadmin instance if its modules were changed):**
        *   Log in as an administrator.
        *   Go to the "Apps" menu.
        *   Find the module(s) that were updated.
        *   Click the "Upgrade" button. This applies any changes to views, models (database schema), and data files.
    *   This can be a manual process per tenant. For a large number of tenants, scripting this process or developing a superadmin tool for batch module upgrades would be a significant enhancement for future development.

Regularly updating and maintaining the platform is an ongoing process essential for its long-term health and success. Always prioritize testing updates in a staging environment before rolling them out to production.
