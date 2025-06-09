# Section 4: Production Deployment Guide

This section provides a comprehensive guide for deploying the Stock Management SaaS application to a production server. Moving from a local development setup to a live environment requires careful attention to security, configuration, and long-term maintenance.

## 4.1 Introduction

Deploying to production involves more than just copying files. It requires setting up a dedicated server, securing it, configuring the application for performance and security, and establishing routines for backups and maintenance. This guide assumes a Linux-based server environment, specifically Ubuntu LTS.

## 4.2 Server Preparation

### 4.2.1 Choosing a Server/Provider
Select a reputable cloud provider (e.g., AWS, Google Cloud, Azure, DigitalOcean, Linode) or a dedicated server. Consider factors like:
*   CPU, RAM, and Storage requirements (start modest, scale as needed).
*   Data center location.
*   Backup services.
*   Pricing.

### 4.2.2 Operating System
It's recommended to use a stable Long-Term Support (LTS) version of a Linux distribution. This guide will assume **Ubuntu LTS** (e.g., Ubuntu 22.04 LTS).

### 4.2.3 Installing Docker and Docker Compose
Connect to your server via SSH.
1.  **Update package lists:**
    ```bash
    sudo apt update
    ```
2.  **Install Docker Engine:**
    Follow the official Docker installation guide for Ubuntu: [https://docs.docker.com/engine/install/ubuntu/](https://docs.docker.com/engine/install/ubuntu/)
3.  **Install Docker Compose:**
    Follow the official Docker Compose installation guide: [https://docs.docker.com/compose/install/](https://docs.docker.com/compose/install/) (Usually involves downloading the binary).
    ```bash
    # Example for a specific version - check official docs for latest
    sudo curl -L "https://github.com/docker/compose/releases/download/v2.20.2/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    sudo chmod +x /usr/local/bin/docker-compose
    ```
4.  **Add your user to the `docker` group (optional, to run Docker commands without `sudo`):**
    ```bash
    sudo usermod -aG docker $USER
    newgrp docker # Apply group changes, or log out and log back in
    ```

### 4.2.4 Basic Security (Firewall)
Use `ufw` (Uncomplicated Firewall) to manage firewall rules.
1.  **Allow SSH (so you don't get locked out):**
    ```bash
    sudo ufw allow OpenSSH
    # or sudo ufw allow 22/tcp
    ```
2.  **Allow HTTP and HTTPS:**
    ```bash
    sudo ufw allow http  # Port 80
    sudo ufw allow https # Port 443
    ```
3.  **Enable ufw:**
    ```bash
    sudo ufw enable
    sudo ufw status verbose # To check
    ```

## 4.3 Code Deployment

1.  **Using Git:**
    On your production server, clone your project repository:
    ```bash
    git clone <your_repository_url>
    cd <project_directory_name>
    ```
    For updates, you'll typically use `git pull` within this directory.
2.  **Ensure all necessary files are on the server:**
    This includes your `docker-compose.yml`, `Dockerfile.odoo`, Odoo configuration files (`odoo.conf`, `superadmin.conf`), and the `custom_addons` directory.

## 4.4 Production Configuration (Crucial)

This is the most critical step for security and stability. **DO NOT USE DEVELOPMENT PASSWORDS IN PRODUCTION.**

### 4.4.1 Passwords
*   **PostgreSQL Passwords:**
    *   In `docker-compose.yml`: Change `POSTGRES_PASSWORD` for both `db` and `db_superadmin` services to strong, unique passwords.
*   **Odoo Database Passwords:**
    *   In `odoo.conf`: Update `db_password` to match the new password for the `db` service.
    *   In `superadmin.conf`: Update `db_password` to match the new password for the `db_superadmin` service.
*   **Odoo Master Admin Passwords:**
    *   In `odoo.conf`: Set a strong, unique `admin_passwd`.
    *   In `superadmin.conf`: Set a different strong, unique `admin_passwd`. This password is vital for the superadmin's ability to create tenant databases.

### 4.4.2 Secret Management (Recommendation)
Hardcoding passwords directly in configuration files is not ideal for high-security environments. For true production, consider:
*   **Docker Secrets:** For Swarm mode.
*   **Environment Variables from a Vault:** Using tools like HashiCorp Vault to inject secrets as environment variables at runtime.
*   **`.env` file (with restricted permissions):** Store secrets in an `.env` file and reference them in `docker-compose.yml`. Ensure this file is not committed to Git and has strict file permissions.
For this project, ensure you at least change the hardcoded defaults.

### 4.4.3 Odoo Configuration (`odoo.conf`, `superadmin.conf`)
Review and update the following settings for production:
*   `proxy_mode = True`: **Essential.** Tells Odoo it's running behind a reverse proxy, so it correctly handles `X-Forwarded-*` headers for scheme, host, and client IP.
*   `workers = (2 * CPU_CORES) + 1`: A common starting point for the number of Odoo worker processes. Adjust based on your server's CPU cores and monitoring.
*   `limit_time_cpu = 600`: Max CPU time in seconds for a request.
*   `limit_time_real = 1200`: Max wall clock time in seconds for a request.
*   `limit_memory_hard = 2684354560` (e.g., 2.5GB): Max virtual memory per worker.
*   `limit_memory_soft = 2147483648` (e.g., 2GB): Memory limit after which a worker may be recycled. Adjust these based on available RAM and observed usage.
*   `log_level = info` (or `warn`): Reduce log verbosity compared to `debug` in development.
*   `addons_path = /mnt/extra-addons,/usr/lib/python3/dist-packages/odoo/addons`: Ensure this path is correct and includes your custom addons.
*   `dbfilter_format = %d_db`: (In `odoo.conf` only) Ensure this is set for the tenant-facing instance.
*   **Remove or comment out `admin_passwd` after initial setup if you manage databases externally and want to disable the web-based database manager for that instance.** For the superadmin instance, `admin_passwd` is critical for the tenant creation process.

### 4.4.4 `docker-compose.yml` for Production
*   **Port Mappings:** You might remove direct port mappings for database services (e.g., `"5433:5432"` for `db_superadmin`) if they are not meant to be accessed directly from outside the Docker host. The Odoo services will access them via the internal Docker network. The Odoo service ports (`18069`, `18070`) will be accessed via the reverse proxy, not directly from the internet.
*   **Volume Paths:** Ensure named volumes are used, which is the default in the provided `docker-compose.yml`. If using host-bound paths (not recommended for databases), ensure they are correct for the server environment.
*   **Restart Policy:** `restart: unless-stopped` or `restart: always` is generally good for production services.

## 4.5 Building and Starting Services

1.  **Build the custom Odoo image:**
    ```bash
    docker-compose build
    ```
2.  **Start all services in detached mode:**
    ```bash
    docker-compose up -d
    ```
3.  **Check status:**
    ```bash
    docker-compose ps
    ```

## 4.6 DNS Configuration

You'll need to configure DNS records with your DNS provider:

*   **Super Admin:** Create an `A` record for your superadmin interface.
    *   Example: `superadmin.yourdomain.com` -> `YOUR_SERVER_IP`
*   **Tenants (Wildcard):** Create a wildcard `A` record to handle all tenant subdomains.
    *   Example: `*.yourshopsdomain.com` -> `YOUR_SERVER_IP`
    *   And potentially a root record for the domain itself if you plan to have a landing page there: `yourshopsdomain.com` -> `YOUR_SERVER_IP`

Allow time for DNS propagation.

## 4.7 Reverse Proxy Setup (Nginx Example)

A reverse proxy is essential for a production Odoo setup.

### 4.7.1 Why Use a Reverse Proxy?
*   **SSL/HTTPS Termination:** Handles incoming HTTPS connections, encrypting/decrypting data, and passing unencrypted HTTP traffic to Odoo locally.
*   **Load Balancing:** (Not covered in this basic setup) Can distribute traffic across multiple Odoo instances.
*   **Security:** Acts as an additional layer, can serve static files, implement rate limiting, etc.
*   **Clean URLs:** Hides port numbers from users.

### 4.7.2 Installation
```bash
sudo apt update
sudo apt install nginx -y
sudo systemctl start nginx
sudo systemctl enable nginx
```

### 4.7.3 Basic Nginx Configuration
Create Nginx server block configuration files in `/etc/nginx/sites-available/`.

**1. Super Admin Server Block (`/etc/nginx/sites-available/superadmin`):**
```nginx
server {
    listen 80;
    server_name superadmin.yourdomain.com;

    # Redirect HTTP to HTTPS (Certbot will handle this better later)
    # location / {
    # return 301 https://$host$request_uri;
    # }

    # For Certbot verification (if needed before SSL)
    location /.well-known/acme-challenge/ {
        root /var/www/html;
    }

    location / {
        proxy_pass http://localhost:18070; # Or your mapped superadmin Odoo port
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header X-Forwarded-Host $host; # Odoo needs this
        proxy_read_timeout 600s; # Increase timeout for long Odoo operations
        proxy_send_timeout 600s;
    }

    location /longpolling/ { # Required for Odoo live chat/notifications
        proxy_pass http://localhost:18070/longpolling/; # Ensure port matches
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

**2. Tenant Server Block (`/etc/nginx/sites-available/tenants`):**
This will handle `yourshopsdomain.com` and `*.yourshopsdomain.com`.
```nginx
server {
    listen 80;
    server_name yourshopsdomain.com *.yourshopsdomain.com;

    # For Certbot verification
    location /.well-known/acme-challenge/ {
        root /var/www/html;
    }

    location / {
        proxy_pass http://localhost:18069; # Or your mapped tenant Odoo port
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header X-Forwarded-Host $host;
        proxy_read_timeout 600s;
        proxy_send_timeout 600s;
    }

    location /longpolling/ {
        proxy_pass http://localhost:18069/longpolling/; # Ensure port matches
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

**Enable these sites:**
```bash
sudo ln -s /etc/nginx/sites-available/superadmin /etc/nginx/sites-enabled/
sudo ln -s /etc/nginx/sites-available/tenants /etc/nginx/sites-enabled/

sudo nginx -t # Test Nginx configuration
sudo systemctl reload nginx
```

### 4.7.4 SSL/HTTPS Setup (Let's Encrypt with Certbot)
1.  **Install Certbot and Nginx plugin:**
    ```bash
    sudo apt install certbot python3-certbot-nginx -y
    ```
2.  **Obtain SSL Certificates:**
    For the superadmin domain:
    ```bash
    sudo certbot --nginx -d superadmin.yourdomain.com
    ```
    For the tenant domains (use the main domain and a wildcard if your DNS provider supports it via Certbot, otherwise you might need to list main domains or use DNS challenge for wildcards):
    ```bash
    sudo certbot --nginx -d yourshopsdomain.com -d *.yourshopsdomain.com
    ```
    *   Follow the prompts. Certbot will modify your Nginx configuration to include SSL settings and set up automatic redirects from HTTP to HTTPS.
    *   If `*.yourshopsdomain.com` with `--nginx` plugin causes issues (some DNS providers require DNS validation for wildcards), you might need to issue certs for specific tenant subdomains as they are created or use a more advanced Certbot setup (DNS challenge). For simplicity, you could initially omit the wildcard and add it later or handle SSL for tenants differently. A common approach is to get a wildcard cert using DNS validation.

3.  **Verify Auto-Renewal:**
    Certbot should set up a cron job or systemd timer for automatic renewal. Test it:
    ```bash
    sudo certbot renew --dry-run
    ```

## 4.8 Initial Startup and Testing

1.  **Access Super Admin:** `https://superadmin.yourdomain.com`. Log in, ensure `saas_management_tools` is installed.
2.  **Create a Test Tenant:** Use a subdomain like `test1.yourshopsdomain.com`.
3.  **Verify Tenant Access:** `https://test1.yourshopsdomain.com`. Log in.
4.  **Test Core Functionalities:** Basic stock operations, user management within the tenant.

## 4.9 Backup and Maintenance (Critical Overview)

This is a brief overview. Implement a robust strategy.

### 4.9.1 What to Back Up
*   **PostgreSQL Databases:**
    *   All tenant databases (dynamically discover or use `pg_dumpall`).
    *   The `db_superadmin` database.
    *   Tool: `pg_dump` or `pg_dumpall`.
    Example for a single DB:
    ```bash
    docker exec <db_container_name> pg_dump -U <user> -Fc <database_name> > backup_file.dump
    ```
    You'll need to script this for all tenant DBs.
*   **Odoo Filestores (Volumes):**
    *   The content of Docker named volumes: `odoo-data`, `odoo-superadmin-data`.
    *   You can back up the Docker volume paths directly on the host (e.g., `/var/lib/docker/volumes/<volume_name>/_data`).
*   **Configuration Files:** `docker-compose.yml`, Odoo configs, Nginx configs.

### 4.9.2 Backup Strategy
*   **Frequency:** Daily for databases and filestores is common.
*   **Retention:** Keep several days/weeks of backups (e.g., 7 daily, 4 weekly).
*   **Off-site Storage:** **Crucial.** Store backups on a separate server/service (e.g., AWS S3, Backblaze B2).
*   **Test Restores:** Regularly test your backup restoration process.

### 4.9.3 Maintenance
*   **Odoo Updates:** Plan for applying Odoo patches and minor version updates. This involves updating the base image and potentially testing custom modules.
*   **System Updates:** Regularly update the server's operating system and packages (`sudo apt update && sudo apt upgrade`).
*   **Security Monitoring:** Keep an eye on security bulletins for Odoo, PostgreSQL, Nginx, and Ubuntu.

## 4.10 Monitoring (Brief Mention)

Implement monitoring for:
*   **Server Resources:** CPU, RAM, disk space, network I/O (e.g., using `htop`, `vmstat`, or cloud provider tools).
*   **Docker Container Health:** `docker ps`, check for restarting containers.
*   **Odoo Application Logs:** Check Odoo logs (forwarded to Docker logs) for errors.
*   **Nginx Logs:** `/var/log/nginx/access.log` and `/var/log/nginx/error.log`.
Consider tools like Prometheus, Grafana, or cloud provider monitoring services for more advanced monitoring.

This guide provides a foundational approach to deploying the SaaS application. Each production environment may have unique requirements, so adapt these guidelines accordingly.
