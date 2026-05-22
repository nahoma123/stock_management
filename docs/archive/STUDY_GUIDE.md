# Odoo SaaS Platform Study Guide

This guide provides a comprehensive overview of the Odoo SaaS platform to help you understand its architecture and key components.

## 1. Introduction

This project is a multi-tenant Odoo SaaS platform that allows you to host and manage multiple Odoo 17 instances. It uses Docker for containerization, with a Go backend and React frontend for the superadmin panel.

## 2. Technologies

*   **Odoo Backend:** Odoo (Python), version 17
*   **Superadmin Backend:** Go, Gin Framework, Gorilla WebSocket
*   **Superadmin Frontend:** React, Vite
*   **Database:** PostgreSQL
*   **Containerization:** Docker, Docker Compose

## 3. Databases

There is **one** database instance in this project:

*   `db`: This is the primary PostgreSQL database used by:
    *   The main Odoo application (`odoo` service), which serves the tenants.
    *   The Go backend (`backend` service) for the superadmin panel.

## 4. Directory Structure

*   `custom_addons/`: Contains custom Odoo modules.
*   `backend/`: The Go backend for the superadmin panel.
*   `frontend/`: The React frontend for the superadmin panel.
*   `docs/`: Project documentation.
*   `docker-compose.yml`: Defines all the Docker services.
*   `Dockerfile.*`: Dockerfiles for each service.
*   `*.conf`: Odoo configuration files.

## 5. Key Components

### Odoo Service (`odoo`)

*   The core Odoo application that serves the tenants.
*   It connects to the `db` database.
*   Custom addons from the `custom_addons` directory are loaded into this service.

### Superadmin Backend (`backend`)

*   A Go application that provides the API for the superadmin panel.
*   It uses the Gin framework and Gorilla WebSocket for real-time communication.
*   It connects to the `db` database to manage tenants.

### Superadmin Frontend (`frontend`)

*   A React application built with Vite that provides the user interface for the superadmin panel.
*   It communicates with the `backend` service to manage tenants.

## 6. Custom Addons

*   `boutique_theme`: A custom theme for the Odoo backend.
*   `saas_management_tools`: Tools for creating and managing SaaS tenants.
*   `shopping_portal`: A multi-instance shopping portal.

## 7. Getting Started

To set up the local development environment, follow the instructions in the `03_local_development_setup.md` file in the `docs` directory.

## 8. Further Reading

For more detailed information, please refer to the documentation in the `docs` directory:

*   `01_introduction_saas_overview.md`
*   `02_docker_configuration.md`
*   `04_production_deployment_guide.md`
*   `05_super_admin_guide.md`
*   `06_tenant_customization_guide.md`
*   `07_extending_the_platform.md`
*   `08_future_considerations_roadmap.md`
*   `Ollama_Integration_Guide.md`
