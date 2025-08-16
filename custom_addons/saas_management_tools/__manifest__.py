{
    'name': 'SaaS Management Tools',
    'version': '1.0',
    'summary': 'Tools for managing SaaS tenants (creation, etc.)',
    'category': 'Tools',
    'author': 'Jules AI Agent',
    'depends': ['base'],
    'data': [
        'security/ir.model.access.csv',
        'data/saas_tenant_sequence.xml',
        'views/saas_tenant_views.xml',
        'views/tenant_creation_wizard_views.xml',
        'views/saas_dashboard_views.xml',
    ],
    'installable': True,
    'application': True, # So it appears in the Apps list of superadmin
    'auto_install': False,
}
