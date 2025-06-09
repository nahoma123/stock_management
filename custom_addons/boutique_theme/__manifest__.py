# ./custom_addons/boutique_theme/__manifest__.py
{
    'name': 'Boutique Theme & Core UI',
    'version': '1.0',
    'summary': 'Core UI customizations and branding for the Boutique ERP.',
    'sequence': 1, # Load early
    'description': """
Primary Theme and UI Modifications:
- Custom Branding (Logo, Colors)
- Simplified Navigation
- Hides unnecessary Odoo elements
    """,
    'category': 'Theme/Backend',
    'author': 'Your Company Name', # Put your name/company here
    'website': 'Your Website (optional)',
    'license': 'LGPL-3', # Or 'OEEL-1' if basing on Enterprise & keeping private
    'depends': [
        'base',         # Core Odoo framework
        'web',          # Required for backend UI changes
        'stock',        # For Inventory views
        'point_of_sale',# For POS views
        'crm',          # For CRM views
        # Add other core modules you intend to heavily modify later
    ],
    'data': [
        # We will add XML files here later for views
        'views/web_layout.xml',
        'views/menu_items.xml',
    ],
    'assets': {
        'web.assets_backend': [
            # We will add SCSS/CSS/JS files here later
            'boutique_theme/static/src/scss/theme.scss',
        ],
        'point_of_sale.assets': [
            # For POS specific styling later
            # 'boutique_theme/static/src/scss/pos_theme.scss',
        ],
    },
    'installable': True,
    'application': True, # Makes it easily findable in Apps
    'auto_install': False,
}