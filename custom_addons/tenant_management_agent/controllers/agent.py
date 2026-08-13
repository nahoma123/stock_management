import os
import secrets
from datetime import datetime, timezone

import odoo
from odoo import http
from odoo.http import request
from odoo.modules.module import get_module_path


CONTRACT_VERSION = '1.0'


class TenantManagementAgent(http.Controller):
    @http.route('/tenant-agent/v1/health', type='http', auth='none', methods=['GET'], csrf=False)
    def health(self, **kwargs):
        expected_token = os.environ.get('TENANT_AGENT_TOKEN', '')
        supplied_token = request.httprequest.headers.get('X-Tenant-Agent-Token', '')
        if not expected_token or not secrets.compare_digest(expected_token, supplied_token):
            return request.make_json_response({'error': 'unauthorized'}, status=401)

        installed_modules = request.env['ir.module.module'].sudo().search_read(
            [('state', '=', 'installed')],
            ['name', 'shortdesc', 'installed_version'],
            order='name',
        )
        modules = []
        for module in installed_modules:
            path = get_module_path(module['name']) or ''
            if path.startswith('/mnt/tenant-addons'):
                layer = 'tenant'
            elif path.startswith('/mnt/platform-addons'):
                layer = 'platform'
            else:
                layer = 'core'
            modules.append({
                'name': module['name'],
                'label': module['shortdesc'],
                'version': module['installed_version'],
                'layer': layer,
            })

        company = request.env.company.sudo()
        payload = {
            'contract_version': CONTRACT_VERSION,
            'checked_at': datetime.now(timezone.utc).isoformat(),
            'status': 'healthy',
            'tenant_id': os.environ.get('MANAGED_TENANT_ID'),
            'database': request.env.cr.dbname,
            'company': company.name,
            'odoo_version': odoo.release.version,
            'users': request.env['res.users'].sudo().search_count([('active', '=', True), ('share', '=', False)]),
            'products': request.env['product.template'].sudo().search_count([('active', '=', True)]),
            'warehouses': request.env['stock.warehouse'].sudo().search_count([]),
            'modules': modules,
        }
        return request.make_json_response(payload)
