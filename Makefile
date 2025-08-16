# Makefile for Odoo SaaS Platform

# Variables
COMPOSE = docker compose

# Targets
.PHONY: up down logs logs-odoo logs-superadmin ps shell-odoo shell-superadmin restart build refresh update-saas-module

up:
	@echo "Starting Odoo SaaS Platform..."
	@$(COMPOSE) up -d

down:
	@echo "Stopping Odoo SaaS Platform..."
	@$(COMPOSE) down

logs:
	@echo "Tailing logs for all services..."
	@$(COMPOSE) logs -f

logs-odoo:
	@echo "Tailing logs for the odoo service..."
	@$(COMPOSE) logs -f odoo

logs-superadmin:
	@echo "Tailing logs for the odoo_superadmin service..."
	@$(COMPOSE) logs -f odoo_superadmin

ps:
	@echo "Showing status of running containers..."
	@$(COMPOSE) ps

shell-odoo:
	@echo "Opening a shell inside the odoo container..."
	@$(COMPOSE) exec odoo /bin/bash

shell-superadmin:
	@echo "Opening a shell inside the odoo_superadmin container..."
	@$(COMPOSE) exec odoo_superadmin /bin/bash

restart:
	@echo "Restarting all services..."
	@$(COMPOSE) restart

build:
	@echo "Rebuilding Docker images..."
	@$(COMPOSE) build

refresh: down up update-saas-module

update-saas-module:
	@echo "Updating saas_management_tools module..."
	@$(COMPOSE) exec odoo_superadmin odoo -c /etc/odoo/odoo.conf -d your_superadmin_db -u saas_management_tools --stop-after-init --no-http