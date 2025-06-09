# Section 2: Docker Configuration Explained

This section details the Docker setup used to containerize and manage the various components of the SaaS application. Docker and Docker Compose are fundamental to ensuring consistent environments, simplifying deployment, and enabling scalability.

## 2.1 Role of `docker-compose.yml`

The `docker-compose.yml` file is the cornerstone of the Docker configuration. It defines and orchestrates all the services that make up the application. Its key responsibilities include:

*   **Service Definition:** Specifies each service (container) required, such as Odoo application servers and PostgreSQL databases.
*   **Image Building:** Instructs Docker on how to build custom images (e.g., using `Dockerfile.odoo`) or which pre-built images to pull from a registry.
*   **Port Mapping:** Exposes container ports to the host machine, allowing access to the Odoo instances from a web browser.
*   **Volume Management:** Defines and mounts volumes for persistent data storage, ensuring that data like database files and Odoo filestores are not lost when containers are stopped or restarted.
*   **Environment Variables:** Sets environment variables for services, crucial for configuring database connections, passwords, and other parameters.
*   **Networking:** Automatically sets up a default network, enabling seamless communication between the defined services using their service names as hostnames.
*   **Dependencies:** Manages the startup order of services using `depends_on`, ensuring that critical services like databases are running before application services that rely on them.

## 2.2 Service Explanations

The `docker-compose.yml` file defines the following core services:

### 2.2.1 `odoo` (Tenant Odoo Service)

*   **Build Context:**
    ```yaml
    build:
      context: .
      dockerfile: Dockerfile.odoo
    ```
    This service is built using the custom `Dockerfile.odoo` located in the project's root directory.
*   **Ports:**
    *   `"18069:8069"`: Maps port `18069` on the host to port `8069` (Odoo's default HTTP port) inside the container. This is the main access point for tenant users.
*   **Volumes:**
    *   `odoo-data:/var/lib/odoo`: Mounts the named volume `odoo-data` to `/var/lib/odoo` inside the container. This directory is Odoo's default location for its filestore (attachments, binary data).
    *   `./custom_addons:/mnt/extra-addons`: Mounts the local `custom_addons` directory into the container, making custom Odoo modules available to this instance.
    *   `./odoo.conf:/etc/odoo/odoo.conf`: Mounts the `odoo.conf` file from the host into the container, providing the specific configuration for this Odoo service (including `dbfilter_format`).
*   **Environment Variables:**
    *   `HOST=db`: Specifies that the database host is the service named `db`.
    *   `PORT=5432`: The port for the PostgreSQL database.
    *   `USER=odoo`: The username for connecting to the tenant PostgreSQL service.
    *   `PASSWORD=nahom_admin`: The password for the `odoo` user on the tenant PostgreSQL service. (Note: This should be a secure password in a production environment).
*   **Dependencies:**
    *   `depends_on: - db`: Ensures that the `db` service (tenant PostgreSQL) is started before this Odoo service attempts to connect to it.

### 2.2.2 `db` (Tenant PostgreSQL Service)

*   **Image:**
    *   `image: postgres:16`: Uses the official PostgreSQL version 16 image from Docker Hub.
*   **Ports:** (No ports are explicitly mapped to the host by default for this service, as it's primarily accessed by other services within the Docker network. Direct host access can be added for debugging if needed.)
*   **Volumes:**
    *   `postgres-data:/var/lib/postgresql/data`: Mounts the named volume `postgres-data` to `/var/lib/postgresql/data`, which is the default data directory for PostgreSQL. This ensures that all tenant databases are persisted.
*   **Environment Variables:**
    *   `POSTGRES_USER=odoo`: The username that will be created in PostgreSQL for Odoo to connect with.
    *   `POSTGRES_PASSWORD=nahom_admin`: The password for the `POSTGRES_USER`. This must match the `PASSWORD` used by the `odoo` service.
    *   `POSTGRES_DB=postgres`: Specifies the default database to create (though Odoo will manage its own databases).
*   **Dependencies:** None explicitly listed, as it's a base service.

### 2.2.3 `odoo_superadmin` (Super Admin Odoo Service)

*   **Build Context:**
    ```yaml
    build:
      context: .
      dockerfile: Dockerfile.odoo
    ```
    This service is also built using the custom `Dockerfile.odoo`.
*   **Ports:**
    *   `"18070:8069"`: Maps port `18070` on the host to port `8069` in the container. This provides access to the super admin Odoo interface.
*   **Volumes:**
    *   `odoo-superadmin-data:/var/lib/odoo`: Mounts the named volume `odoo-superadmin-data` for this Odoo instance's filestore.
    *   `./custom_addons:/mnt/extra-addons`: Mounts the same `custom_addons` directory, ensuring the `saas_management_tools` module is available.
    *   `./superadmin.conf:/etc/odoo/odoo.conf`: Mounts the `superadmin.conf` file, providing specific configuration for the super admin Odoo instance.
*   **Environment Variables:**
    *   `HOST=db_superadmin`: Specifies that its database host is the service named `db_superadmin`.
    *   `USER=odoo_superadmin`: The username for connecting to its dedicated PostgreSQL service.
    *   `PASSWORD=superadmin_db_password`: The password for the `odoo_superadmin` user.
*   **Dependencies:**
    *   `depends_on: - db_superadmin`: Ensures the `db_superadmin` service is started before this Odoo service.

### 2.2.4 `db_superadmin` (Super Admin PostgreSQL Service)

*   **Image:**
    *   `image: postgres:16`: Uses the official PostgreSQL version 16 image.
*   **Ports:**
    *   `"5433:5432"`: (Optional, for debugging) Maps port `5433` on the host to port `5432` in the container, allowing direct database access if needed.
*   **Volumes:**
    *   `postgres-superadmin-data:/var/lib/postgresql/data`: Mounts the named volume `postgres-superadmin-data` for persisting the super admin database.
*   **Environment Variables:**
    *   `POSTGRES_USER=odoo_superadmin`: The username for the super admin's Odoo instance.
    *   `POSTGRES_PASSWORD=superadmin_db_password`: The password for this user.
    *   `POSTGRES_DB=postgres`: Default database.
*   **Dependencies:** None explicitly listed.

## 2.3 Role of `Dockerfile.odoo`

The `Dockerfile.odoo` file provides a custom Docker image definition for the Odoo services (`odoo` and `odoo_superadmin`).

*   **Base Image:** `FROM odoo:18.0` specifies that our custom image is built upon the official Odoo version 18.0 image. This provides the core Odoo application and its environment.
*   **`postgresql-client` Installation:**
    ```dockerfile
    USER root
    RUN apt-get update && \
        apt-get install -y --no-install-recommends postgresql-client && \
        rm -rf /var/lib/apt/lists/*
    USER odoo
    ```
    These instructions are crucial for the `odoo_superadmin` service. The `postgresql-client` package provides command-line tools like `createdb`, which the `saas_management_tools` module uses (via `subprocess`) to create new databases for tenants on the `db` (tenant PostgreSQL) service. The `USER` directives switch to `root` for installation permissions and then back to the `odoo` user, which is the default non-privileged user in the base Odoo image.

## 2.4 Data Persistence with Named Volumes

Docker named volumes (`odoo-data`, `postgres-data`, `odoo-superadmin-data`, `postgres-superadmin-data`) are used to persist data generated by and used by Docker containers.

*   **Independence from Container Lifecycle:** Data in named volumes is managed by Docker and stored outside the container's writable layer. This means that even if a container is stopped, removed, or recreated, the data in the volume remains intact.
*   **Easy Backup and Management:** Named volumes can be more easily listed, backed up, or migrated compared to bind-mounted host directories in some scenarios.

For this project:
*   `odoo-data` and `odoo-superadmin-data` store the Odoo filestores (e.g., attachments, images).
*   `postgres-data` and `postgres-superadmin-data` store the actual PostgreSQL database files.

## 2.5 Inter-Service Communication (Default Network)

When `docker-compose up` is executed, Docker Compose automatically creates a default bridge network for all the services defined in the `docker-compose.yml` file.

*   **Service Discovery:** Services can reach each other using their service names as hostnames (e.g., the `odoo` service can connect to the `db` service using the hostname `db`). Docker handles the DNS resolution within this private network.
*   **Isolation:** This network is isolated from the host machine's network by default, unless ports are explicitly published (mapped).

This simplifies the configuration, as services don't need to know each other's IP addresses, which can change if containers are restarted.
