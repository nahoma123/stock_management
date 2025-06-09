{
    'name': 'Shopping Portal',
    'version': '1.0',
    'summary': 'Multi-instance shopping portal with user registration',
    'sequence': 1,
    'description': """
Shopping Portal Features:
- User Registration and Authentication
- Product Aggregation from Multiple Instances
- Shopping Cart Management
- Order Processing
    """,
    'category': 'Website',
    'author': 'Your Company Name',
    'website': 'Your Website',
    'license': 'LGPL-3',
    'depends': [
        'base',
        'web',
        'website',
        'auth_signup',
        'portal',
        'sale',
        'product',
    ],
    'data': [
        'security/ir.model.access.csv',
        'views/portal_templates.xml',
        'views/portal_my_home.xml',
        'views/portal_layout.xml',
        'views/registration_templates.xml',
        'views/product_templates.xml',
    ],
    'assets': {
        'web.assets_frontend': [
            'shopping_portal/static/src/scss/portal.scss',
            'shopping_portal/static/src/js/portal.js',
        ],
    },
    'installable': True,
    'application': True,
    'auto_install': False,
} 