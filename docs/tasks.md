# Project Tasks

## Completing Owner Info Features
- [x] Document task in docs/tasks.md
- [x] Expose API Key in React Dashboard
- [x] Add owner_email field to res_company in daily_sales_report
- [x] Update mail_template.xml to use owner_email
- [x] Create views for owner_email and add to __manifest__.py
- [x] Verify functionality

## Mobile App API & Push Notifications
- [x] Document task in docs/tasks.md
- [x] Update Go Backend Models (MobileDevice, MinNotificationAmount)
- [x] Implement Mobile API endpoints (devices, settings)
- [x] Implement Webhook receiver in Go Backend
- [x] Create mobile_push_notifications Odoo Addon
- [x] Create API Documentation (docs/mobile_api.md)
- [x] Verify functionality

## Enhanced Mobile Dashboard API
- [x] Document task in docs/tasks.md
- [x] Enhance backend/main.go (Yesterday stats & 7-Day Trend)
- [x] Update docs/mobile_api.md
- [x] Verify functionality

## Project Cleanup & Refactoring
- [x] Document task in docs/tasks.md
- [x] Remove Dead Code & Artifacts
- [x] Clean up Orphaned Docker Containers
- [x] Modularize Go Backend
- [x] Verify functionality

## Flutter Mobile Application Scaffolding
- [/] Document task in docs/tasks.md
- [ ] Scaffold Flutter Project Template
- [ ] Configure pubspec.yaml Dependencies
- [ ] Setup Architecture Skeleton (Core & Feature folders)
- [ ] Implement Core Network, Theme, and Storage Utilities
- [ ] Verify Scaffolding via flutter analyze

## Initial Product Data Import
- [x] Add CSV and XLSX upload wizard with preview and validation
- [x] Map products, ETB/USD prices, embedded primary pictures, and opening stock
- [x] Add Machine Code upsert behavior and Model Code product field
- [x] Install Inventory and importer for newly provisioned tenants
- [ ] Verify import against a representative customer workbook

## Managed Tenant Customization Foundation
- [x] Formalize core, platform, and tenant addon layers
- [x] Add authenticated, versioned tenant management agent
- [x] Add central health and installed-module monitoring
- [x] Add isolated customization package validation and deployment
- [x] Add superadmin monitoring and customization controls
- [x] Add versioned staging, automated module maintenance, and rollback releases
- [x] Add central release and operator audit records
