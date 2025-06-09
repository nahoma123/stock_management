from odoo import models, fields

class TenantCreationWizard(models.TransientModel):
    _name = 'saas.tenant.creation.wizard'
    _description = 'Tenant Creation Wizard'

    name = fields.Char(string='Shop Name', required=True)
    subdomain = fields.Char(string='Requested Subdomain', required=True)
    # Add admin user details for the new tenant if needed

    def action_confirm_creation(self):
        SaasTenant = self.env['saas.tenant']
        # Basic validation (e.g., check if subdomain already exists in saas.tenant)
        existing_tenant = SaasTenant.search([('subdomain', '=', self.subdomain)], limit=1)
        if existing_tenant:
            raise models.UserError(f"A tenant with subdomain '{self.subdomain}' already exists or is being processed.")

        new_tenant = SaasTenant.create({
            'name': self.name,
            'subdomain': self.subdomain,
            'state': 'draft',
        })
        # Trigger the creation process (which is currently a placeholder)
        new_tenant.action_create_tenant_database()
        return {
            'type': 'ir.actions.act_window',
            'res_model': 'saas.tenant',
            'view_mode': 'form',
            'res_id': new_tenant.id,
            'target': 'current',
        }
