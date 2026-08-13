# Team Overview

This is the nontechnical starting point for understanding and working with the Stock Platform. Use the role guides below for day-to-day work. Technical implementation details live separately in Developer Docs.

## What the product does

The Stock Platform provisions an isolated Odoo stock-management environment for each customer company. The central superadmin application creates and monitors tenants, controls access, and manages approved customer customizations.

Each customer receives:

- A separate business database and Odoo environment
- Inventory, sales, stock reporting, and initial product-data import
- A company-specific URL
- Monitoring managed by the platform team
- Optional reviewed custom modules without changing other customers

## Choose your guide

- [Sales & Demos](./sales.md): positioning, demo preparation, and what can be promised
- [Product & Testing](./product-testing.md): acceptance testing, evidence, and release decisions
- [Operations](./operations.md): tenant creation, monitoring, access, and incident handling

## Current product boundary

The platform supports managed tenant provisioning, isolated customer data, product imports, central monitoring, and versioned customer customizations. Billing automation, high-availability deployment, and fully automated disaster recovery are roadmap items and should not be presented as completed capabilities.

The mobile application and push-notification flow are prototypes, not customer-ready features.

## Where to report issues

Record the tenant name, user action, expected result, actual result, time, and a screenshot when appropriate. Product decides priority; developers diagnose implementation details; operations handles service availability and customer access.
