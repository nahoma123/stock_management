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
        return super(SaasTenant, self).create(vals)
    subdomain = fields.Char(string='Subdomain', required=True)
    db_name = fields.Char(string='Database Name', compute='_compute_db_name', store=True, readonly=True)
    state = fields.Selection([
        ('draft', 'Draft'),
        ('creating', 'Creating'),
        ('active', 'Active'),
        ('error', 'Error'),
    ], string='Status', default='draft', copy=False)
    creation_log = fields.Text(string="Creation Log", readonly=True)
    license_expiry_date = fields.Date(string='License Expiry Date', copy=False)
    notes = fields.Text(string='Internal Notes')
    dashboard_summary = fields.Text(string="Dashboard Summary", compute='_compute_dashboard_summary')

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
        self.ensure_one()
        self._log_message("Starting database creation process...")

        # Fetch necessary passwords from current (superadmin) Odoo config
        # admin_passwd is used by Odoo to authorize database creation/deletion
        master_admin_password = self.env['ir.config_parameter'].sudo().get_param('admin_passwd')
        if not master_admin_password:
            self._log_message("Error: admin_passwd (master admin password) is not set in the superadmin Odoo configuration.")
            self.state = 'error'
            return

        # Database connection parameters for the *tenant's database server* (from docker-compose 'db' service)
        # These are assumptions based on the docker-compose.yml provided earlier.
        # Ideally, these should come from a more secure config source.
        tenant_db_host = 'db' # Service name of the tenant PostgreSQL container
        tenant_db_user = 'odoo' # User for the tenant PostgreSQL service
        tenant_db_password = 'nahom_admin' # Password for the tenant PostgreSQL service

        try:
            # 1. Create the new database
            self._log_message(f"Attempting to create database: {self.db_name}")
            # To use create_database, the current Odoo instance (superadmin) must be configured
            # with db_host, db_user, db_password that has CREATEDB privileges on the target server (tenant_db_host).
            # Odoo's create_database uses the connection of the *current* instance if not specified otherwise.
            # This is tricky because superadmin's odoo.conf points to db_superadmin, not the tenant 'db' service.
            # For simplicity and directness, we'll use a subprocess call to 'createdb'
            # This requires 'createdb' to be available in the odoo_superadmin container
            # and the PostgreSQL server 'db' to be configured to accept connections from 'odoo_superadmin' for user 'odoo'.

            # Ensure the odoo_superadmin container has psql client tools:
            # Add 'RUN apt-get update && apt-get install -y postgresql-client && rm -rf /var/lib/apt/lists/*' to Dockerfile.backend for odoo_superadmin if not present

            create_db_command = [
                'createdb',
                '--host', tenant_db_host,
                '--port', '5432', # Default PG port
                '--username', tenant_db_user,
                '--owner', tenant_db_user, # Tenant DB user should own the DB
                self.db_name
            ]
            _logger.info(f"Executing createdb command: {' '.join(create_db_command)}")
            env = {'PGPASSWORD': tenant_db_password, **self.env.context.get('env', {})} # Pass password via env

            process = subprocess.Popen(create_db_command, stdout=subprocess.PIPE, stderr=subprocess.PIPE, env=env)
            stdout, stderr = process.communicate(timeout=60) # 60 seconds timeout

            if process.returncode == 0:
                self._log_message(f"Database {self.db_name} created successfully.")
                _logger.info(f"createdb stdout: {stdout.decode()}")
            else:
                error_msg = f"Failed to create database {self.db_name}. Return code: {process.returncode}. Error: {stderr.decode()}"
                self._log_message(error_msg)
                _logger.error(error_msg)
                self.state = 'error'
                return

            # 2. Initialize the new database with Odoo modules
            self._log_message(f"Initializing Odoo in database {self.db_name}...")
            modules_to_install = 'base,web,boutique_theme,shopping_portal'

            # The odoo-bin command needs to run from the odoo_superadmin container,
            # but target the new database on the 'db' (tenant) postgresql service.
            # It should use the addons path available to the odoo_superadmin service.
            # The superadmin.conf might point to db_superadmin, so we explicitly provide tenant db params.
            init_command = [
                'odoo', # Assuming 'odoo' is alias for 'odoo-bin' or it's in PATH
                '-c', '/etc/odoo/odoo.conf', # Use superadmin's config for addons_path etc.
                '--database', self.db_name,
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

            # This command runs Odoo which might take a while.
            process_init = subprocess.Popen(init_command, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
            stdout_init, stderr_init = process_init.communicate(timeout=600) # 10 minutes timeout for init

            if process_init.returncode == 0:
                self._log_message(f"Database {self.db_name} initialized successfully with modules: {modules_to_install}.")
                _logger.info(f"Odoo init stdout: {stdout_init.decode()}")
                self.state = 'active'
            else:
                error_msg = f"Failed to initialize modules in {self.db_name}. Return code: {process_init.returncode}. Error: {stderr_init.decode()}"
                self._log_message(error_msg)
                _logger.error(error_msg)
                self.state = 'error'
                # Consider dropping the database if init fails to allow retry
                # self._drop_tenant_db() # Implement this method if needed

        except subprocess.TimeoutExpired:
            error_msg = "Database creation/initialization process timed out."
            self._log_message(error_msg)
            _logger.error(error_msg)
            self.state = 'error'
        except Exception as e:
            error_msg = f"An unexpected error occurred: {str(e)}"
            self._log_message(error_msg)
            _logger.exception("Tenant DB Creation/Init Unexpected Error")
            self.state = 'error'
        finally:
            self.env.cr.commit() # Commit changes to state and log

    def _compute_dashboard_summary(self):
        # This method computes for each record, which is not ideal for a global summary.
        # For a true dashboard, this logic would be on a separate model or via read_group.
        # Here, we'll just make it available on any saas.tenant record,
        # and the dashboard view will show it from a dummy record or the first one.
        # A better approach for a real dashboard: use a transient model.
        # For this subtask, we'll simplify and assume this text will be shown on a generic dashboard view.

        tenants = self.env['saas.tenant'].search([]) # Search all tenants
        total_tenants = len(tenants)
        active_tenants = len(tenants.filtered(lambda t: t.state == 'active'))
        error_tenants = len(tenants.filtered(lambda t: t.state == 'error'))
        draft_tenants = len(tenants.filtered(lambda t: t.state == 'draft'))
        creating_tenants = len(tenants.filtered(lambda t: t.state == 'creating'))

        summary = f"Total Tenants: {total_tenants}\n"
        summary += f"Active: {active_tenants}\n"
        summary += f"Error: {error_tenants}\n"
        summary += f"Draft: {draft_tenants}\n"
        summary += f"Creating: {creating_tenants}"

        for record in self: # Assign the same summary to all records queried by the ORM for this compute
            record.dashboard_summary = summary

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
            threaded_create = threading.Thread(target=self.sudo(tenant.id)._create_and_initialize_tenant_db)
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
