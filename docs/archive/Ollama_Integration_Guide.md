# Integrating Ollama with Your Odoo Project

This guide provides a step-by-step walkthrough for integrating a local Ollama instance with your Odoo application. We will create a practical feature that automatically generates a product description using an AI model running on Ollama.

## Prerequisites

Before you begin, ensure you have the following:

1.  A running Odoo instance.
2.  A running Ollama instance on the same machine.
3.  The Ollama API endpoint is accessible at `http://localhost:11434`.
4.  At least one model is installed in Ollama (e.g., `llama3.1:8b`).
5.  A custom Odoo addon to modify. We will use the `boutique_theme` addon in this guide.

## Step 1: Add `requests` to Dependencies

To communicate with the Ollama API from Python, we need the `requests` library.

1.  Open the manifest file for your addon: `custom_addons/boutique_theme/__manifest__.py`.
2.  Add `'requests'` to the `depends` list.

```python
# custom_addons/boutique_theme/__manifest__.py
{
    'name': 'Boutique Theme',
    'summary': 'A custom theme for the Odoo backend.',
    'version': '1.0',
    'category': 'Theme/Backend',
    'author': 'Your Name',
    'website': 'https://www.yourcompany.com',
    'depends': [
        'base',
        'web',
        'stock',
        'point_of_sale',
        'crm',
        'requests',  # Add this line
    ],
    'data': [
        'security/ir.model.access.csv',
        'views/web_layout.xml',
        'views/menu_items.xml',
        'views/product_views.xml',
    ],
    'assets': {
        'web.assets_backend': [
            'boutique_theme/static/src/scss/theme.scss',
        ],
    },
    'installable': True,
    'application': True,
}
```

## Step 2: Add the "Generate Description" Button

Next, we'll add a button to the product form that will trigger our AI generation function.

1.  Open the view file: `custom_addons/boutique_theme/views/product_views.xml`.
2.  Add the following `xpath` to insert a button into the header of the form view.

```xml
<!-- custom_addons/boutique_theme/views/product_views.xml -->
<?xml version="1.0" encoding="utf-8"?>
<odoo>
    <!-- Inherit the product template form view -->
    <record id="product_template_form_view_inherit_boutique" model="ir.ui.view">
        <field name="name">product.template.form.inherit.boutique</field>
        <field name="model">product.template</field>
        <field name="inherit_id" ref="product.product_template_only_form_view"/>
        <field name="arch" type="xml">
            <!-- Add the button to the header -->
            <xpath expr="//header" position="inside">
                <button name="generate_ai_description" type="object" string="Generate Description with AI" class="oe_highlight"/>
            </xpath>

            <!-- Hide the 'Routes' field -->
            <xpath expr="//field[@name='route_ids']" position="attributes">
                <attribute name="invisible">1</attribute>
            </xpath>

            <!-- Hide the 'Weight' and 'Volume' fields on the inventory tab -->
             <xpath expr="//group[@name='group_lots_and_weight']/group[2]" position="attributes">
                 <attribute name="invisible">1</attribute>
             </xpath>
        </field>
    </record>
</odoo>
```

## Step 3: Create the Python Method

Now, we'll create the Python method that the button will call. This method will connect to Ollama and generate the description.

1.  Create a new directory `models` inside `custom_addons/boutique_theme/`.
2.  Create a new file named `product_template.py` inside the new `models` directory.
3.  Add the following code to the new file:

```python
# custom_addons/boutique_theme/models/product_template.py
import requests
import json
from odoo import models, fields, api

class ProductTemplate(models.Model):
    _inherit = 'product.template'

    def generate_ai_description(self):
        for product in self:
            if not product.name:
                continue

            prompt = f"Generate a short, engaging product description for: {product.name}"

            data = {
                "model": "llama3.1:8b", # Or any other model you have
                "prompt": prompt,
                "stream": False # We want the full response at once
            }

            try:
                response = requests.post("http://localhost:11434/api/generate", json=data, timeout=60)
                response.raise_for_status()  # Raise an exception for bad status codes

                full_response = response.json()
                ai_description = full_response.get('response', '').strip()

                if ai_description:
                    product.description_sale = ai_description

            except requests.exceptions.RequestException as e:
                # Handle connection errors, timeouts, etc.
                # You could log this error or show a notification to the user
                pass # For now, we'''ll just ignore errors

```

## Step 4: Update `__init__.py` Files

We need to make sure our new Python file is loaded by Odoo.

1.  Create a new file `__init__.py` inside the `custom_addons/boutique_theme/models/` directory.

```python
# custom_addons/boutique_theme/models/__init__.py
from . import product_template
```

2.  Update the main `__init__.py` file for the addon to import the `models` directory.

```python
# custom_addons/boutique_theme/__init__.py
from . import models
```

## Step 5: Apply the Changes

Finally, to see your new feature in action:

1.  **Restart the Odoo server.**
2.  **Upgrade the `boutique_theme` addon.**
    *   Go to the "Apps" menu in Odoo.
    *   Remove the "Apps" filter and search for "boutique_theme".
    *   Open the addon and click "Upgrade".

You should now see the "Generate Description with AI" button on your product forms. Click it to see the AI generate a description for you!
