import subprocess
import threading
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
        # Create a new environment and cursor for this thread
        with self.pool.cursor() as cr:
            env = self.env(cr=cr)
            tenant = env['saas.tenant'].browse(self.id)

            tenant._log_message("Starting database creation process...")

            # Fetch necessary passwords from current (superadmin) Odoo config
            # admin_passwd is used by Odoo to authorize database creation/deletion
            master_admin_password = env['ir.config_parameter'].sudo().get_param('admin_passwd')
            if not master_admin_password:
                tenant._log_message("Error: admin_passwd (master admin password) is not set in the superadmin Odoo configuration.")
                tenant.state = 'error'
                return

            # Database connection parameters for the *tenant's database server* (from docker-compose 'db' service)
            # These are assumptions based on the docker-compose.yml provided earlier.
            # Ideally, these should come from a more secure config source.
            tenant_db_host = 'db' # Service name of the tenant PostgreSQL container
            tenant_db_user = 'odoo' # User for the tenant PostgreSQL service
            tenant_db_password = 'nahom_admin' # Password for the tenant PostgreSQL service

            try:
                # 1. Create the new database
                tenant._log_message(f"Attempting to create database: {tenant.db_name}")

                create_db_command = [
                    'createdb',
                    '--host', tenant_db_host,
                    '--port', '5432', # Default PG port
                    '--username', tenant_db_user,
                    '--owner', tenant_db_user, # Tenant DB user should own the DB
                    tenant.db_name
                ]
                _logger.info(f"Executing createdb command: {' '.join(create_db_command)}")
                env_vars = {'PGPASSWORD': tenant_db_password}
                process = subprocess.Popen(create_db_command, stdout=subprocess.PIPE, stderr=subprocess.PIPE, env=env_vars)
                stdout, stderr = process.communicate(timeout=60) # 60 seconds timeout

                if process.returncode == 0:
                    tenant._log_message(f"Database {tenant.db_name} created successfully.")
                    _logger.info(f"createdb stdout: {stdout.decode()}")
                else:
                    error_msg = f"Failed to create database {tenant.db_name}. Return code: {process.returncode}. Error: {stderr.decode()}"
                    tenant._log_message(error_msg)
                    _logger.error(error_msg)
                    tenant.state = 'error'
                    return

                # 2. Initialize the new database with Odoo modules
                tenant._log_message(f"Initializing Odoo in database {tenant.db_name}...")
                modules_to_install = 'base,web,boutique_theme,shopping_portal'

                init_command = [
                    'odoo', # Assuming 'odoo' is alias for 'odoo-bin' or it's in PATH
                    '-c', '/etc/odoo/odoo.conf', # Use superadmin's config for addons_path etc.
                    '--database', tenant.db_name,
                    '--db_host', tenant_db_host,
                    '--db_port', '5432',
                    '--db_user', tenant_db_user,
                    '--db_password', tenant_db_password,
                    '--init', modules_to_install,
                    '--without-demo=all',
                    '--stop-after-init',
                    '--no-xmlrpc', # Avoid starting services
                    '--no-longpolling',
                    '--logfile', '/dev/null' # Or a specific log file for init
                ]
                _logger.info(f"Executing Odoo init command: {' '.join(init_command)}")

                process_init = subprocess.Popen(init_command, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
                stdout_init, stderr_init = process_init.communicate(timeout=600) # 10 minutes timeout for init

                if process_init.returncode == 0:
                    tenant._log_message(f"Database {tenant.db_name} initialized successfully with modules: {modules_to_install}.")
                    _logger.info(f"Odoo init stdout: {stdout_init.decode()}")
                    tenant.state = 'active'
                else:
                    error_msg = f"Failed to initialize modules in {tenant.db_name}. Return code: {process_init.returncode}. Error: {stderr_init.decode()}"
                    tenant._log_message(error_msg)
                    _logger.error(error_msg)
                    tenant.state = 'error'

            except subprocess.TimeoutExpired:
                error_msg = "Database creation/initialization process timed out."
                tenant._log_message(error_msg)
                _logger.error(error_msg)
                tenant.state = 'error'
            except Exception as e:
                error_msg = f"An unexpected error occurred: {str(e)}"
                tenant._log_message(error_msg)
                _logger.exception("Tenant DB Creation/Init Unexpected Error")
                tenant.state = 'error'
            finally:
                # Commit changes made within this thread's transaction
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
