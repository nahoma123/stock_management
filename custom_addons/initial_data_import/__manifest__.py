{
    'name': 'Initial Product Data Import',
    'version': '18.0.1.0.0',
    'category': 'Inventory/Inventory',
    'summary': 'Preview and import products, prices, pictures, and opening stock',
    'depends': ['stock', 'product'],
    'data': [
        'security/ir.model.access.csv',
        'views/product_template_views.xml',
        'wizards/initial_data_import_views.xml',
    ],
    'installable': True,
    'application': False,
    'license': 'LGPL-3',
}
