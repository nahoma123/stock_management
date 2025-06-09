from odoo import models, fields, api

class ShoppingCart(models.Model):
    _name = 'shopping.portal.cart'
    _description = 'Shopping Cart'

    user_id = fields.Many2one('res.users', string='User', required=True)
    line_ids = fields.One2many('shopping.portal.cart.line', 'cart_id', string='Cart Lines')
    total_amount = fields.Float(string='Total Amount', compute='_compute_total_amount')
    currency_id = fields.Many2one('res.currency', string='Currency', default=lambda self: self.env.company.currency_id)
    
    @api.depends('line_ids.price_subtotal')
    def _compute_total_amount(self):
        for cart in self:
            cart.total_amount = sum(cart.line_ids.mapped('price_subtotal'))

class ShoppingCartLine(models.Model):
    _name = 'shopping.portal.cart.line'
    _description = 'Shopping Cart Line'

    cart_id = fields.Many2one('shopping.portal.cart', string='Cart', required=True)
    product_id = fields.Many2one('shopping.portal.product', string='Product', required=True)
    quantity = fields.Float(string='Quantity', default=1.0)
    price_unit = fields.Float(string='Unit Price', related='product_id.price')
    price_subtotal = fields.Float(string='Subtotal', compute='_compute_price_subtotal')
    
    @api.depends('quantity', 'price_unit')
    def _compute_price_subtotal(self):
        for line in self:
            line.price_subtotal = line.quantity * line.price_unit 