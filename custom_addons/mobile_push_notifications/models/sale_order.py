# -*- coding: utf-8 -*-
from odoo import models, api
import requests
import logging
import threading

_logger = logging.getLogger(__name__)

class SaleOrder(models.Model):
    _inherit = 'sale.order'

    def action_confirm(self):
        res = super(SaleOrder, self).action_confirm()
        
        for order in self:
            payload = {
                "db_name": self.env.cr.dbname,
                "order_id": order.id,
                "order_name": order.name,
                "amount_total": order.amount_total,
                "customer_name": order.partner_id.name or "Unknown Customer"
            }
            
            # Send webhook asynchronously to not block the Odoo transaction
            def send_webhook(data):
                try:
                    # In docker-compose, 'backend' resolves to the Go backend container. 
                    # If running outside, this might need to be configurable.
                    url = "http://backend:8080/api/webhooks/odoo/sale"
                    response = requests.post(url, json=data, timeout=5)
                    _logger.info("Sent mobile push notification webhook: %s", response.status_code)
                except Exception as e:
                    _logger.error("Failed to send mobile push notification webhook: %s", e)
            
            thread = threading.Thread(target=send_webhook, args=(payload,))
            thread.start()
            
        return res
