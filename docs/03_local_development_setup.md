# Section 3: Local Development Setup

This section guides you through setting up the Stock Management SaaS project on your local machine for development and testing purposes.

## 3.1 Prerequisites

Before you begin, ensure you have the following software installed on your system:

*   **Docker Engine:** The core Docker platform for running containers.
*   **Docker Compose:** The tool for defining and running multi-container Docker applications. It's often included with Docker Desktop installations.
*   **Git:** For cloning the project repository (or a method to download/extract the project files).
*   **Code Editor:** A text editor or Integrated Development Environment (IDE) for viewing and editing project files (e.g., VS Code, Sublime Text, Atom).

## 3.2 Getting the Code

1.  **Clone the Repository:**
    Open your terminal or command prompt and run:
    ```bash
    git clone <repository_url>
    cd <project_directory_name>
    ```
    Replace `<repository_url>` with the actual URL of the Git repository and `<project_directory_name>` with the name of the folder created by the clone.

2.  **Alternatively, Download/Extract:**
    If you received the project as an archive (e.g., a ZIP file), extract it to a suitable location on your computer.

## 3.3 Directory Structure Overview

Once you have the code, familiarize yourself with the key files and directories:

*   `docker-compose.yml`: Defines all the Docker services, networks, and volumes.
*   `Dockerfile.odoo`: Custom Dockerfile for building the Odoo images with necessary additions like `postgresql-client`.
*   `odoo.conf`: Configuration file for the main Odoo instance that serves tenants. Contains `dbfilter_format`.
*   `superadmin.conf`: Configuration file for the Super Admin Odoo instance.
*   `custom_addons/`: Directory containing all custom Odoo modules developed for this project.
    *   `saas_management_tools/`: The core module for tenant management, database creation, etc.
    *   `boutique_theme/`, `shopping_portal/`: Example tenant-specific modules.
*   `.env` (if present): Might contain environment-specific variables, though in this project, most configurations are directly in `docker-compose.yml` or Odoo config files.

## 3.4 Configuration Review (Optional but Recommended)

Before the first launch, it's good practice to review the default configurations, especially passwords:

*   **`docker-compose.yml`:**
    *   Check `POSTGRES_PASSWORD` for the `db` service (e.g., `nahom_admin`).
    *   Check `POSTGRES_PASSWORD` for the `db_superadmin` service (e.g., `superadmin_db_password`).
    *   These passwords are also used in the respective Odoo service environment variables (`PASSWORD`).
*   **`odoo.conf`:**
    *   `db_password`: Should match `POSTGRES_PASSWORD` of the `db` service in `docker-compose.yml`.
    *   `admin_passwd`: This is the master admin password for this Odoo instance (used for database management operations by Odoo itself, not the OS user). Set a strong password (e.g., `your_tenant_master_password`).
*   **`superadmin.conf`:**
    *   `db_password`: Should match `POSTGRES_PASSWORD` of the `db_superadmin` service in `docker-compose.yml`.
    *   `admin_passwd`: Master admin password for the superadmin Odoo instance (e.g., `your_superadmin_master_password`). This is critical for the `saas_management_tools` to create new tenant databases.

**Note:** For local development, the default passwords might be acceptable, but they **must** be changed for any staging or production deployment.

## 3.5 Building Docker Images

The Odoo services (`odoo` and `odoo_superadmin`) use a custom Docker image defined by `Dockerfile.odoo`. To build this image (and any other images that might require building):

```bash
docker-compose build
```
This command reads `docker-compose.yml`, finds services with a `build` section, and executes the Docker build process as defined in the specified Dockerfile (`Dockerfile.odoo` in this case).

## 3.6 Starting Services

To start all defined services in detached mode (running in the background):

```bash
docker-compose up -d
```

*   `-d`: Detached mode.

To check the status of your running services:
```bash
docker-compose ps
```
This will show you which containers are running, their ports, and their current state.

## 3.7 Viewing Logs

To view the logs of a specific service (useful for troubleshooting):

```bash
docker-compose logs -f <service_name>
```
Replace `<service_name>` with the name of the service from `docker-compose.yml` (e.g., `odoo`, `db`, `odoo_superadmin`).
*   `-f`: Follow log output. Press `Ctrl+C` to stop following.

For example, to view the logs of the superadmin Odoo instance:
```bash
docker-compose logs -f odoo_superadmin
```

## 3.8 `/etc/hosts` Configuration for Tenant Routing

The `odoo` service uses `dbfilter_format = %d_db` in `odoo.conf`. This means Odoo will select the database based on the first part of the hostname (subdomain) used in the URL. For local development, you need to tell your operating system that hostnames like `tenant1.localhost` should resolve to your local machine (`127.0.0.1`).

**Why it's needed:** When you try to access `http://tenant1.localhost:18069`, your browser asks your OS to find `tenant1.localhost`. Without an `/etc/hosts` entry, your OS won't know where to send this request.

**How to edit:** You'll need administrator/root privileges to edit this file.

*   **Linux & macOS:**
    Open a terminal and run:
    ```bash
    sudo nano /etc/hosts
    # Or use another editor like vim: sudo vim /etc/hosts
    ```
*   **Windows:**
    1.  Open Notepad (or your preferred text editor) **as Administrator**.
    2.  Go to `File > Open` and navigate to `C:\Windows\System32\drivers\etc\hosts`. Make sure to select "All Files (*.*)" in the file type dropdown if you don't see the `hosts` file.

**Example entries:**
Add lines like the following to your `hosts` file:
```
127.0.0.1 tenant1.localhost
127.0.0.1 another-tenant.localhost
127.0.0.1 shopxyz.localhost
```
This tells your computer that any request to `tenant1.localhost` (or other defined subdomains) should go to your local IP address `127.0.0.1`.

## 3.9 Accessing Super Admin Instance

1.  **URL:** `http://localhost:18070` (as defined by the port mapping for `odoo_superadmin` in `docker-compose.yml`).
2.  **Login:**
    *   The first time you access it, Odoo might ask you to create a master password (this is the `admin_passwd` from `superadmin.conf`).
    *   After setting up the database (if it's the very first run for the `db_superadmin` volume), the default Odoo login is typically:
        *   **Email:** `admin`
        *   **Password:** `admin` (You should change this immediately).
3.  **Install `saas_management_tools`:**
    *   Once logged in, go to `Apps`.
    *   Remove the default "Apps" filter in the search bar.
    *   Search for "SaaS Management Tools".
    *   Click "Install" on the `saas_management_tools` module.

## 3.10 Creating a Test Tenant

1.  After installing `saas_management_tools` in the superadmin instance, navigate to `SaaS Management > Create New Tenant`.
2.  Fill in the details:
    *   **Shop Name:** e.g., "My Test Shop"
    *   **Requested Subdomain:** e.g., `testshop` (This will be used to form the `db_name` like `testshop_db` and the URL like `testshop.localhost`).
3.  Click "Create Tenant".
4.  You will be taken to the newly created tenant record. It will initially be in "Draft" state.
5.  Click the "Create Database" button. The state should change to "Creating".
    *   Monitor the logs in `saas_management_tools/models/saas_tenant.py` (via `docker-compose logs -f odoo_superadmin`) and the `creation_log` field on the tenant form for progress.
    *   This process involves creating a new PostgreSQL database and then initializing it with Odoo modules, which can take a few minutes.
6.  Once successful, the tenant state will change to "Active".

**Remember:** For each new tenant subdomain (e.g., `testshop`), you need to add a corresponding entry to your `/etc/hosts` file:
```
127.0.0.1 testshop.localhost
```

## 3.11 Accessing a Test Tenant Instance

1.  **URL:** Assuming you created a tenant with subdomain `testshop` and your `/etc/hosts` file is configured: `http://testshop.localhost:18069` (using the port `18069` for the main `odoo` service).
2.  **Login:**
    *   The database is initialized without demo data. The default admin credentials for a new Odoo database are typically:
        *   **Email:** `admin`
        *   **Password:** `admin`
    *   You will be prompted to change this password on first login and set up company details.

## 3.12 Stopping Services

*   To stop all running services:
    ```bash
    docker-compose down
    ```
*   To stop services AND remove the named volumes (effectively deleting all Odoo data, including tenant databases and filestores - useful for starting completely fresh):
    ```bash
    docker-compose down -v
    ```
    **Caution:** Use `docker-compose down -v` with care, as it will delete all data stored in the Docker volumes associated with this project.

## 3.13 Troubleshooting Common Local Setup Issues

*   **Port Conflicts:**
    *   **Symptom:** Error messages like "port is already allocated" or "bind: address already in use".
    *   **Solution:** Check if another application is using the ports defined in `docker-compose.yml` (e.g., `18069`, `18070`, `5433`). Either stop the conflicting application or change the host port mapping in `docker-compose.yml` (e.g., change `"18069:8069"` to `"18099:8069"` and access Odoo on `http://localhost:18099`).
*   **Docker Build Failures (`docker-compose build`):**
    *   **Symptom:** Errors during the image building process.
    *   **Solution:** Carefully read the error messages. Common issues include:
        *   Network problems preventing package downloads (e.g., during `apt-get update`).
        *   Syntax errors in `Dockerfile.odoo`.
        *   Base image not found (e.g., `odoo:18.0` if there's a typo or Docker Hub is inaccessible).
*   **Odoo Not Starting (Container exits or is unhealthy):**
    *   **Symptom:** `docker-compose ps` shows a service as "Exited" or "Restarting".
    *   **Solution:**
        1.  Check logs: `docker-compose logs -f <service_name>` (e.g., `odoo` or `odoo_superadmin`).
        2.  Common Odoo startup issues:
            *   **Database connection problems:** Ensure the corresponding PostgreSQL service (`db` or `db_superadmin`) is running and healthy. Verify `HOST`, `USER`, `PASSWORD` in Odoo's environment variables and Odoo config files match the PostgreSQL service configuration.
            *   **Incorrect `addons_path` or missing critical modules.**
            *   **File permission issues** with mounted volumes (less common with default Docker setups).
*   **Tenant Creation Failures:**
    *   **Symptom:** Tenant stays in "Creating" state for too long, or changes to "Error" state.
    *   **Solution:**
        1.  Check the `creation_log` field on the tenant's form in the Super Admin Odoo instance.
        2.  Check the logs of the `odoo_superadmin` container: `docker-compose logs -f odoo_superadmin`. Look for errors related to `createdb` or the Odoo initialization (`odoo --init`) subprocess calls.
        3.  Ensure `postgresql-client` is correctly installed in the `odoo_superadmin` image (verify `Dockerfile.odoo` and rebuild if necessary).
        4.  Verify network connectivity between `odoo_superadmin` container and the `db` (tenant PostgreSQL) container.
        5.  Check that the `admin_passwd` in `superadmin.conf` is correctly set and that the `db_user` (`odoo`) and `db_password` (`nahom_admin`) for the tenant PostgreSQL service (`db`) are correct in `saas_tenant.py`'s `_create_and_initialize_tenant_db` method. (Ideally, these credentials for the tenant DB service should be configurable rather than hardcoded in the Python file).
