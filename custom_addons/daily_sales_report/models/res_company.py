from odoo import models, fields, api
from datetime import datetime, time
import logging

_logger = logging.getLogger(__name__)

class ResCompany(models.Model):
    _inherit = 'res.company'

    owner_email = fields.Char(string='Owner Email', help='Email address of the company owner for daily sales reports.')

    @api.model
    def _send_daily_sales_report(self):
        """
        Calculates daily sales and sends an email report for each company.
        This method is meant to be called by a scheduled action (cron job).
        """
        for company in self.search([]):
            today_start = datetime.combine(datetime.today(), time.min)
            today_end = datetime.combine(datetime.today(), time.max)
            
            domain = [
                ('company_id', '=', company.id),
                ('state', 'in', ['sale', 'done']),
                ('date_order', '>=', today_start),
                ('date_order', '<=', today_end),
            ]
            orders = self.env['sale.order'].search(domain)
            total_sales = sum(orders.mapped('amount_total'))
            
            _logger.info(f"Daily sales for {company.name}: {len(orders)} orders, total {total_sales}")
            
            template = self.env.ref('daily_sales_report.email_template_daily_sales', raise_if_not_found=False)
            if template:
                # Pass data via context so the template can render it
                ctx = dict(self.env.context, total_sales=total_sales, order_count=len(orders))
                template.with_context(ctx).send_mail(company.id, force_send=True)
            else:
                _logger.warning("Daily sales report template not found!")
