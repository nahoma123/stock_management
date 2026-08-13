from odoo import fields, models


class ProductTemplate(models.Model):
    _inherit = 'product.template'

    model_code = fields.Char(index=True)
