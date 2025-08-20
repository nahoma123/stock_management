import subprocess
import threading
import os
from odoo import models, fields, api, service, exceptions
import logging

_logger = logging.getLogger(__name__)

class SaasTenant(models.Model):
    _name = 'saas.tenant'
    _description = 'SaaS Tenant'

    name = fields.Char(string='Tenant Name', required=True, copy=False, readonly=True, default='New')

    @api.model
    def create(self, vals):
        if vals.get('name', 'New') == 'New':
            vals['name'] = self.env['ir.sequence'].next_by_code('saas.tenant') or '/'
        # Explicitly set show_create_button based on initial state.
        # The 'state' field has a default of 'draft', but this handles cases
        # where an initial state is provided during creation.
        if vals.get('state'):
            vals['show_create_button'] = (vals['state'] == 'draft')
        else:
            # If no state is provided, it will default to 'draft', so the button should be shown.
            vals['show_create_button'] = True
        return super(SaasTenant, self).create(vals)

    def write(self, vals):
        if 'state' in vals:
            vals['show_create_button'] = (vals['state'] == 'draft')
        return super(SaasTenant, self).write(vals)
    subdomain = fields.Char(string='Subdomain', required=True)
    db_name = fields.Char(string='Database Name', compute='_compute_db_name', store=True, readonly=True)
    state = fields.Selection([
        ('draft', 'Draft'),
        ('creating', 'Creating'),
        ('active', 'Active'),
        ('error', 'Error'),
    ], string='Status', default='draft', copy=False)
    show_create_button = fields.Boolean(string="Show Create Button", default=True, store=True)
    creation_log = fields.Text(string="Creation Log", readonly=True)
    license_expiry_date = fields.Date(string='License Expiry Date', copy=False)
    notes = fields.Text(string='Internal Notes')

    _sql_constraints = [
        ('subdomain_uniq', 'unique (subdomain)', 'Subdomain must be unique!'),
        ('db_name_uniq', 'unique (db_name)', 'Database name must be unique!'),
    ]

    @api.depends('subdomain')
    def _compute_db_name(self):
        for record in self:
            if record.subdomain:
                record.db_name = f"{record.subdomain.lower().replace('.', '_')}_db"
            else:
                record.db_name = False

    def _log_message(self, message):
        self.ensure_one()
        _logger.info(f"Tenant {self.name} ({self.db_name}): {message}")
        self.write({'creation_log': f"{self.creation_log or ''}\n{message}"})

    def _create_and_initialize_tenant_db(self):
        # This method runs in a separate thread.
        # It's crucial to create a new cursor and environment for this thread.
        with self.pool.cursor() as cr:
            env = self.env(cr=cr)
            tenant = env['saas.tenant'].browse(self.id)

            try:
                tenant._log_message("Starting database creation process in new thread...")
                config_param = env['ir.config_parameter'].sudo()

                # --- 1. Get Tenant DB connection details from config ---
                db_host = config_param.get_param('saas.tenant_db_host')
                db_port = config_param.get_param('saas.tenant_db_port')
                db_user = config_param.get_param('saas.tenant_db_user')
                db_password = config_param.get_param('saas.tenant_db_password')

                if not all([db_host, db_port, db_user, db_password]):
                    missing_params = [p for p in ['host', 'port', 'user', 'password'] if not locals().get(f'db_{p}')]
                    raise exceptions.UserError(f"The following tenant database settings are missing from System Parameters: {', '.join(missing_params)}")

                # --- 2. Create the new database using createdb command ---
                tenant._log_message(f"Attempting to create database: {tenant.db_name} on host '{db_host}'")

                env_vars = os.environ.copy()
                env_vars['PGPASSWORD'] = db_password

                create_db_command = [
                    'createdb', '--host', db_host, '--port', db_port,
                    '--username', db_user, '--owner', db_user, tenant.db_name
                ]
                
                _logger.info(f"Executing createdb command: {' '.join(create_db_command)}")
                process_create = subprocess.Popen(
                    create_db_command, env=env_vars, stdout=subprocess.PIPE,
                    stderr=subprocess.PIPE, text=True
                )
                stdout, stderr = process_create.communicate(timeout=60)

                if process_create.returncode != 0:
                    error_msg = f"Failed to create database via createdb. STDERR: {stderr}"
                    tenant._log_message(error_msg)
                    _logger.error(f"createdb for {tenant.db_name} failed:\n{stderr}")
                    tenant.state = 'error'
                    return
                
                tenant._log_message(f"Database {tenant.db_name} created successfully.")

                # --- 3. Initialize the new database with Odoo modules ---
                tenant._log_message(f"Initializing Odoo in database {tenant.db_name}...")
                modules_to_install = 'base,web,boutique_theme,shopping_portal'
                
                init_command = [
                    'odoo', '--database', tenant.db_name, '--db_host', db_host,
                    '--db_port', db_port, '--db_user', db_user, '--db_password', db_password,
                    '--init', modules_to_install, '--without-demo=all', '--stop-after-init',
                    '--no-xmlrpc', '--no-longpolling', '--logfile', '/dev/null',
                ]
                
                _logger.info(f"Executing Odoo init command: {' '.join(init_command)}")
                process_init = subprocess.Popen(init_command, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
                stdout, stderr = process_init.communicate(timeout=600)
                
                _logger.info(f"Odoo init for {tenant.db_name} STDOUT:\n{stdout}")
                if stderr:
                    _logger.error(f"Odoo init for {tenant.db_name} STDERR:\n{stderr}")

                if process_init.returncode == 0:
                    tenant._log_message(f"Database initialized with modules: {modules_to_install}.")
                    tenant.state = 'active'
                else:
                    error_msg = f"Failed to initialize modules in {tenant.db_name}. See logs for details."
                    tenant._log_message(error_msg)
                    tenant.state = 'error'

            except subprocess.TimeoutExpired:
                error_msg = "Database initialization process timed out."
                tenant._log_message(error_msg)
                _logger.error(error_msg, exc_info=True)
                tenant.state = 'error'
            except Exception as e:
                error_msg = f"An unexpected error occurred: {str(e)}"
                tenant._log_message(error_msg)
                _logger.error("Tenant DB Creation/Init Unexpected Error", exc_info=True)
                tenant.state = 'error'
            finally:
                cr.commit()

    def action_create_tenant_database(self):
        for tenant in self:
            if tenant.state != 'draft':
                raise exceptions.UserError("Database creation can only be initiated for tenants in 'Draft' state.")

            tenant.write({
                'state': 'creating',
                'creation_log': 'Initiating tenant creation...'
            })
            # The explicit commit below was removed as it is a bad practice and likely caused the issue.
            # self.env.cr.commit()

            # Run in a separate thread to avoid blocking the UI for too long.
            # For a robust solution, use Odoo's job queue or a dedicated job runner.
            threaded_create = threading.Thread(target=tenant.sudo()._create_and_initialize_tenant_db)
            threaded_create.start()

            # Provide immediate feedback to the user
            return {
                'type': 'ir.actions.client',
                'tag': 'display_notification',
                'params': {
                    'title': 'Tenant Creation Started',
                    'message': f'Database creation for "{tenant.name}" has started in the background. Refresh to see status updates.',
                    'sticky': False,
                    'type': 'info',
                }
            }
