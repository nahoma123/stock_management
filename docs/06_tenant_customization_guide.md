# Section 6: Tenant Customization Guide

## 6.1 Understanding Tenant Customization

A core strength of this SaaS platform is providing each tenant with their own isolated Odoo environment. This isolation is primarily at the database level. While all tenants share the same underlying Odoo application codebase and custom addon modules (from `custom_addons/`), their data, configurations, and activated themes are unique to their instance.

Customizations, therefore, are applied at the Odoo application level *within* a tenant's specific environment, leveraging the flexibility of Odoo's module system and theming engine.

## 6.2 Theme Customization (`boutique_theme`)

The `boutique_theme` module is an example of how tenants can achieve a unique look and feel for their Odoo instance, distinct from the standard Odoo UI or other tenants.

### 6.2.1 Overview
The `boutique_theme` provides visual customizations such as:
*   Custom logos for the backend, login screen, and browser favicon.
*   Altered color schemes and typography.
*   Modified layouts for certain views or web pages.

### 6.2.2 How it Works
*   **Location:** The theme's code resides in `custom_addons/boutique_theme/`.
*   **Key Components:**
    *   **XML Views:** Odoo's QWeb templates are modified by inheriting existing templates and using XPath expressions to alter specific elements. These are found in the `views/` directory (e.g., `web_layout.xml`, `product_views.xml`).
    *   **SCSS/CSS Files:** Styling changes are primarily managed through SCSS files (e.g., `static/src/scss/theme.scss`), which are compiled into CSS. These allow for changing colors, fonts, spacing, etc.
    *   **Static Assets:** Custom images (like logos, background images) and fonts are stored in the `static/` directory.

### 6.2.3 Common Customizations (Recap)

Here's a brief recap of how typical theme customizations are implemented within `boutique_theme`:

*   **Changing Logos:**
    *   **Backend Logo (Top-left):** Modify the `web.menu_secondary` template (or a more specific one if Odoo versions differ) to replace the default Odoo logo.
        *   Example (conceptual snippet for an XML view):
            ```xml
            <xpath expr="//header//a[@class='o_menu_brand']" position="replace">
                <a href="/web" class="o_menu_brand">
                    <img src="/boutique_theme/static/src/img/custom_backend_logo.png" alt="Company Logo"/>
                </a>
            </xpath>
            ```
    *   **Login Screen Logo:** Inherit the `web.login` template.
        *   Example:
            ```xml
            <xpath expr="//img[contains(@class,'o_logo')]" position="attributes">
                <attribute name="src">/boutique_theme/static/src/img/custom_login_logo.png</attribute>
            </xpath>
            ```
    *   **Favicon:** Replace or modify the `web.layout` template to point to a custom favicon.
        *   Example:
            ```xml
            <xpath expr="//link[@rel='shortcut icon']" position="attributes">
                <attribute name="href">/boutique_theme/static/src/img/custom_favicon.ico</attribute>
            </xpath>
            ```
    *   Ensure your custom logo files (e.g., `custom_backend_logo.png`, `custom_login_logo.png`, `custom_favicon.ico`) are placed in the theme's `static/src/img/` directory.

*   **Changing Colors & Fonts:**
    *   Modify `custom_addons/boutique_theme/static/src/scss/theme.scss`.
    *   Define or override SCSS variables (e.g., `$primary-color`, `$font-family-sans-serif`).
    *   Add custom CSS rules to target specific Odoo elements.
        ```scss
        // Example in theme.scss
        $primary-color: #3AADA7;
        $secondary-color: #E87A5A;
        $font-family-base: 'Georgia', serif;

        body {
            font-family: $font-family-base;
        }
        .o_navbar_brand {
            color: $primary-color;
        }
        ```

*   **Modifying Text & Layout:**
    *   Inherit existing QWeb templates (XML views) from base Odoo modules or other custom modules.
    *   Use XPath expressions to select elements and modify their content, attributes, or position.
    *   Example: Changing a page title or adding a custom footer to website layouts.

### 6.2.4 Applying Theme Changes

*   **For Developers during Development/Staging:**
    1.  Make code changes to the `boutique_theme` module files.
    2.  Ensure the Odoo server is started with the updated code.
    3.  In a specific tenant's Odoo instance (that you're using for testing):
        *   Go to `Apps`.
        *   Find `boutique_theme`.
        *   Click "Upgrade" (or uninstall and then reinstall if it's a fresh test).
        *   Clear your browser cache and refresh the page to see changes.

*   **For All Tenants (Production Rollout):**
    1.  Once theme changes are tested and stable, the updated `boutique_theme` code (as part of the shared `custom_addons` directory) needs to be deployed to the production server (e.g., via `git pull`).
    2.  The Docker containers for `odoo` and `odoo_superadmin` need to be rebuilt and restarted to use the new code:
        ```bash
        docker-compose build
        docker-compose up -d # This will recreate containers using the new image
        ```
    3.  **Important:** For existing tenants to see the theme changes:
        *   **CSS/JS Changes:** Often, a hard refresh (Ctrl+F5 or Cmd+Shift+R) in the tenant's browser is enough after the server restarts, as Odoo recompiles assets.
        *   **XML/View Changes:** If XML templates defining the structure of views were changed, an administrator would typically need to log into each tenant's Odoo instance (or have an automated script do so) and perform an "Upgrade" operation on the `boutique_theme` module via the Apps menu. This step is necessary for Odoo to re-evaluate the view inheritance and apply structural changes.
        *   The current SaaS setup does **not** automatically push theme *structural* updates (like XML changes) into live, running tenant databases without this explicit module upgrade step within each tenant. Simpler asset changes (CSS, JS, images) are usually picked up after server restart and browser cache clearing.

## 6.3 Functional Customization (Via Modules)

Tenants receive specific functionalities beyond the Odoo core through custom modules installed during their provisioning.

### 6.3.1 Overview
Examples include:
*   `shopping_portal`: Adds e-commerce related features.
*   Future modules like `sms_reporting` could add SMS-based reporting capabilities.

### 6.3.2 How it Works
1.  **Module Development:** Custom modules are developed and placed in the `custom_addons/` directory.
2.  **Provisioning Configuration:** The `saas_management_tools` module (specifically in the `_create_and_initialize_tenant_db` method within `saas_tenant.py`) defines which modules are installed by default for new tenants (e.g., `modules_to_install = 'base,web,boutique_theme,shopping_portal'`).

### 6.3.3 Modifying Existing Custom Modules (for Developers)
If you need to add a new feature or fix a bug in a module like `shopping_portal`:
1.  **Edit Module Code:** Make your changes to the Python files, XML views, static assets, etc., within the specific module's directory (e.g., `custom_addons/shopping_portal/`).
2.  **Test Thoroughly:** Use a development or staging tenant instance to install/upgrade the modified module and test its functionality extensively.
3.  **Deployment and Impact:**
    *   **New Tenants:** After deploying the updated module code to the server and restarting the Odoo services, newly provisioned tenants will automatically get the latest version of the module.
    *   **Existing Tenants:** For existing tenants to benefit from the changes, the updated module must be upgraded within their individual Odoo instances. This can be done manually by an administrator logging into each tenant or potentially via a scripted batch upgrade process (more advanced).

### 6.3.4 Tenant-Specific Configurations
Many Odoo modules, both standard and custom, offer configuration options that tenants can adjust within their own isolated environments. These settings are stored in their respective databases.
*   **Examples:**
    *   Configuring system parameters in Odoo's general settings.
    *   Setting up chart of accounts, taxes, fiscal localizations.
    *   Defining product categories, attributes, and price lists in `stock` or `product` modules.
    *   Configuring payment gateways or shipping methods if using Odoo's e-commerce (`website_sale`).
    *   Any specific settings provided by your custom modules (e.g., API keys for `sms_reporting`, display preferences for `shopping_portal`).
*   Tenants access these settings through the standard Odoo interface (usually under `Settings`, `Configuration` menus within relevant applications).

## 6.4 Limitations of Customization in a SaaS Model

While Odoo is highly customizable, a SaaS environment imposes certain limitations to ensure stability, security, and maintainability for all tenants:

*   **No Direct Code Access for Tenants:** Tenants cannot upload their own custom Odoo modules, directly edit backend Python code, or modify server-side files. This is a fundamental security and operational principle of SaaS.
*   **Standardized Updates:** All tenants generally run on the same version of the core Odoo application and the shared custom addon codebase. Customization is achieved through module installation, theme application, and per-tenant configuration settings stored in their database.
*   **Feature Requests:** If a tenant requires a significant new feature or a deep customization that isn't achievable through existing configurations:
    *   They would typically submit this request to the SaaS provider (you).
    *   The provider then evaluates if the feature is beneficial for multiple tenants and can be incorporated into the standard offering.
    *   Alternatively, for very specific or large-scale needs, the provider might offer it as a separate, premium extension if the platform's architecture is designed to support such pluggable, tenant-specific modules (this is an advanced architectural consideration).

The goal is to strike a balance between providing flexibility for tenants and maintaining a manageable, secure, and scalable platform for the SaaS provider.
