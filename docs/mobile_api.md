# Mobile App API Documentation

This document describes the API endpoints provided by the Superadmin Go backend for consumption by the mobile application.

## Authentication
All endpoints require an API Key associated with a specific tenant.
Pass the API Key via the `X-API-Key` HTTP Header or as a query parameter `?api_key=...`.

**Header Example:**
```http
X-API-Key: tenant_key_1234567890abcdef
```

---

## Endpoints

### 1. Get Mobile Stats
Retrieves sales statistics and top products for the tenant's Odoo database.

- **URL**: `/api/mobile/stats`
- **Method**: `GET`
- **Success Response**: `200 OK`
```json
{
  "stats": {
    "tenant_name": "My Shop",
    "subdomain": "myshop",
    "state": "active",
    "license_expiry_date": "2027-01-01",
    "total_sales_all_time": 150000.0,
    "order_count_all_time": 120,
    "total_sales_today": 5000.0,
    "order_count_today": 2,
    "total_sales_yesterday": 4000.0,
    "order_count_yesterday": 4,
    "total_sales_month": 35000.0,
    "order_count_month": 15
  },
  "weekly_trend": [
    {
      "date": "2026-05-15",
      "total_revenue": 1000.0,
      "order_count": 1
    },
    {
      "date": "2026-05-16",
      "total_revenue": 4000.0,
      "order_count": 4
    }
  ],
  "recent_orders": [
    {
      "name": "S00120",
      "date_order": "2026-05-21T14:30:00Z",
      "amount_total": 1250.0,
      "state": "sale"
    }
  ],
  "top_products": [
    {
      "name": "Product A",
      "quantity": 50,
      "total_revenue": 25000.0
    }
  ]
}
```

---

### 2. Register Mobile Device (Push Notifications)
Registers a mobile device token (FCM/APNs) to receive push notifications for significant business events.

- **URL**: `/api/mobile/devices`
- **Method**: `POST`
- **Request Body**:
```json
{
  "device_token": "fcm_token_string_here",
  "platform": "android" // or "ios"
}
```
- **Success Response**: `200 OK`
```json
{
  "message": "Device registered successfully"
}
```

---

### 3. Unregister Mobile Device
Removes a device token to stop receiving push notifications.

- **URL**: `/api/mobile/devices/:token`
- **Method**: `DELETE`
- **Success Response**: `200 OK`
```json
{
  "message": "Device unregistered successfully"
}
```

---

### 4. Update Notification Settings
Updates the user's notification preferences, such as the minimum sale amount required to trigger a push notification.

- **URL**: `/api/mobile/settings`
- **Method**: `PUT`
- **Request Body**:
```json
{
  "min_notification_amount": 1000.0
}
```
- **Success Response**: `200 OK`
```json
{
  "message": "Settings updated successfully",
  "min_notification_amount": 1000.0
}
```

---

## Odoo Internal Webhooks (Not for Mobile App)

### 1. Sale Webhook
Triggered internally by the `mobile_push_notifications` Odoo addon when a sale order is confirmed.

- **URL**: `/api/webhooks/odoo/sale`
- **Method**: `POST`
- **Authentication**: None (Internal only, authenticated via `db_name` lookup)
- **Request Body**:
```json
{
  "db_name": "tenant1_db",
  "order_id": 42,
  "order_name": "S00042",
  "amount_total": 1500.0,
  "customer_name": "John Doe"
}
```
