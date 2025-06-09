from odoo import models, fields, api
from odoo.exceptions import ValidationError

class PortalUser(models.Model):
    _inherit = 'res.users'

    is_portal_user = fields.Boolean(string='Is Portal User', default=False)
    instance_ids = fields.Many2many('res.company', string='Accessible Instances')
    
    @api.model
    def create_portal_user(self, vals):
        """Create a new portal user with registration"""
        if not vals.get('login') or not vals.get('password'):
            raise ValidationError('Login and password are required for registration')
            
        # Create the user
        user = self.create({
            'name': vals.get('name', ''),
            'login': vals.get('login'),
            'password': vals.get('password'),
            'is_portal_user': True,
            'groups_id': [(6, 0, [self.env.ref('base.group_portal').id])],
        })
        
        return user 