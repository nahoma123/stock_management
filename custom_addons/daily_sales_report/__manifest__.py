{
    'name': 'Daily Sales Report',
    'version': '1.0',
    'category': 'Sales',
    'summary': 'Sends a daily summary of sales to the company owner',
    'depends': ['sale_management', 'mail'],
    'data': [
        'data/mail_template.xml',
        'data/ir_cron.xml',
        'views/res_company_views.xml',
    ],
    'installable': True,
    'application': False,
    'license': 'LGPL-3',
}
