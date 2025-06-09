from odoo import models, fields, api

class ProductAggregator(models.Model):
    _name = 'shopping.portal.product'
    _description = 'Aggregated Products from Multiple Instances'

    name = fields.Char(string='Product Name', required=True)
    instance_id = fields.Many2one('res.company', string='Source Instance', required=True)
    product_id = fields.Integer(string='Original Product ID', required=True)
    price = fields.Float(string='Price', required=True)
    currency_id = fields.Many2one('res.currency', string='Currency', required=True)
    description = fields.Text(string='Description')
    image = fields.Binary(string='Product Image')
    active = fields.Boolean(default=True)
    
    @api.model
    def aggregate_products(self):
        """Aggregate products from all instances"""
        products = self.env['product.product'].search([
            ('sale_ok', '=', True),
            ('active', '=', True)
        ])
        
        for product in products:
            self.create({
                'name': product.name,
                'instance_id': self.env.company.id,
                'product_id': product.id,
                'price': product.list_price,
                'currency_id': self.env.company.currency_id.id,
                'description': product.description_sale,
                'image': product.image_1920,
            }) 