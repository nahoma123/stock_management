odoo.define('shopping_portal.portal', function (require) {
    'use strict';

    var publicWidget = require('web.public.widget');
    var ajax = require('web.ajax');

    publicWidget.registry.ShoppingPortal = publicWidget.Widget.extend({
        selector: '.shopping-portal',
        events: {
            'click .add-to-cart': '_onAddToCart',
            'click .remove-item': '_onRemoveItem',
            'change .quantity-input': '_onQuantityChange',
            'click .checkout-btn': '_onCheckout',
        },

        _onAddToCart: function (ev) {
            var self = this;
            var productId = $(ev.currentTarget).data('product-id');
            
            ajax.jsonRpc('/shop/cart/add', 'call', {
                product_id: productId,
                quantity: 1,
            }).then(function (result) {
                if (result.success) {
                    self._updateCart();
                }
            });
        },

        _onRemoveItem: function (ev) {
            var self = this;
            var lineId = $(ev.currentTarget).data('line-id');
            
            ajax.jsonRpc('/shop/cart/remove', 'call', {
                line_id: lineId,
            }).then(function (result) {
                if (result.success) {
                    self._updateCart();
                }
            });
        },

        _onQuantityChange: function (ev) {
            var self = this;
            var lineId = $(ev.currentTarget).data('line-id');
            var quantity = $(ev.currentTarget).val();
            
            ajax.jsonRpc('/shop/cart/update', 'call', {
                line_id: lineId,
                quantity: quantity,
            }).then(function (result) {
                if (result.success) {
                    self._updateCart();
                }
            });
        },

        _onCheckout: function (ev) {
            window.location.href = '/shop/checkout';
        },

        _updateCart: function () {
            var self = this;
            ajax.jsonRpc('/shop/cart', 'call', {}).then(function (result) {
                if (result.cart_html) {
                    $('.shopping-cart-container').html(result.cart_html);
                }
            });
        },
    });
}); 