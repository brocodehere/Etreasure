import React from 'react';
import { useParams, Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { api } from '../../lib/api';
import type { Order } from '../../types';
import { LoadingState, LoadingSpinner } from '../../components/LoadingSpinner';

export function OrderDetailPage() {
  const { id } = useParams<{ id: string }>();

  const { data: order, isLoading, error } = useQuery<Order>({
    queryKey: ['order', id],
    queryFn: () => api.get<Order>(`/orders/${id}`).then(r => r),
    enabled: !!id,
  });

  const formatAddress = (address: Partial<Order>) => {
    const parts = [
      address.shipping_name,
      address.shipping_address_line1,
      address.shipping_address_line2,
      address.shipping_city,
      address.shipping_state,
      address.shipping_country,
      address.shipping_pin_code
    ].filter(Boolean);
    return parts.join(', ') || 'No address available';
  };

  const formatBillingAddress = (address: Partial<Order>) => {
    const parts = [
      address.billing_name,
      address.billing_address_line1,
      address.billing_address_line2,
      address.billing_city,
      address.billing_state,
      address.billing_country,
      address.billing_pin_code
    ].filter(Boolean);
    return parts.join(', ') || 'No address available';
  };

  if (isLoading) {
    return <LoadingSpinner />;
  }

  if (error || !order) {
    return (
      <div className="p-6">
        <div className="bg-red-50 border border-red-200 rounded-lg p-4">
          <h3 className="text-red-800 font-semibold">Error loading order</h3>
          <p className="text-red-600 mt-1">Unable to load order details. Please try again.</p>
        </div>
      </div>
    );
  }

  const lineItems = order.line_items 
  ? typeof order.line_items === 'string' 
    ? JSON.parse(order.line_items as unknown as string) 
    : order.line_items 
  : [];

  return (
    <LoadingState isLoading={isLoading} error={error}>
      <div className="space-y-6">
        {/* Header */}
        <div className="bg-white rounded-xl shadow-lg border-0 overflow-hidden">
          <div className="bg-gradient-to-r from-indigo-600 via-purple-600 to-pink-600 px-6 py-5">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="p-3 bg-gray-800/50 rounded-xl backdrop-blur-sm border border-white/30 shadow-lg">
                  <svg className="w-7 h-7 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
                  </svg>
                </div>
                <div>
                  <h1 className="text-2xl font-bold text-white">Order #{order.order_number}</h1>
                  <p className="text-white/80 text-sm mt-1">Order Details</p>
                </div>
              </div>
              <div className="flex gap-2">
                <Link
                  to={`/orders/${order.id}/edit`}
                  className="px-4 py-2 bg-white/20 text-white font-semibold rounded-lg hover:bg-white/30 transition-all duration-200 backdrop-blur-sm border border-white/30 flex items-center gap-2"
                >
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                  </svg>
                  Edit Order
                </Link>
                <Link
                  to="/orders"
                  className="px-4 py-2 bg-gray-800/50 text-white font-semibold rounded-lg hover:bg-gray-900/50 transition-all duration-200 backdrop-blur-sm border border-white/30"
                >
                  Back to Orders
                </Link>
              </div>
            </div>
          </div>
        </div>

        {/* Order Status and Basic Info */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <div className="bg-white rounded-xl shadow-lg border-0 p-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Order Status</h3>
            <div className="space-y-3">
              <div className="flex justify-between items-center">
                <span className="text-gray-600">Status</span>
                <span className={`px-3 py-1 rounded-full text-xs font-bold ${
                  order.status === 'delivered' ? 'bg-green-100 text-green-700' :
                  order.status === 'shipped' ? 'bg-blue-100 text-blue-700' :
                  order.status === 'processing' ? 'bg-yellow-100 text-yellow-700' :
                  order.status === 'cancelled' ? 'bg-red-100 text-red-700' :
                  'bg-gray-100 text-gray-700'
                }`}>
                  {order.status}
                </span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-gray-600">Shipping Status</span>
                <span className="px-3 py-1 rounded-full text-xs font-bold bg-indigo-100 text-indigo-700">
                  {order.shipping_status}
                </span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-gray-600">Payment Method</span>
                <span className="text-gray-900 font-medium">{order.payment_method || 'N/A'}</span>
              </div>
            </div>
          </div>

          <div className="bg-white rounded-xl shadow-lg border-0 p-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Order Information</h3>
            <div className="space-y-3">
              <div className="flex justify-between items-center">
                <span className="text-gray-600">Order Date</span>
                <span className="text-gray-900 font-medium">
                  {new Date(order.created_at.replace(' IST', '+05:30')).toLocaleDateString()}
                </span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-gray-600">Total Amount</span>
                <span className="text-xl font-bold text-green-600">
                  {order.currency} {order.total_price ? order.total_price.toFixed(2) : '0.00'}
                </span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-gray-600">Currency</span>
                <span className="text-gray-900 font-medium">{order.currency}</span>
              </div>
            </div>
          </div>

          <div className="bg-white rounded-xl shadow-lg border-0 p-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Customer Information</h3>
            <div className="space-y-3">
              <div>
                <span className="text-gray-600 text-sm">Name</span>
                <p className="text-gray-900 font-medium">{order.customer_name}</p>
              </div>
              <div>
                <span className="text-gray-600 text-sm">Email</span>
                <p className="text-gray-900 font-medium">{order.customer_email}</p>
              </div>
              <div>
                <span className="text-gray-600 text-sm">Phone</span>
                <p className="text-gray-900 font-medium">{order.customer_phone}</p>
              </div>
            </div>
          </div>
        </div>

        {/* Order Items */}
        <div className="bg-white rounded-xl shadow-lg border-0 overflow-hidden">
          <div className="px-6 py-4 border-b border-gray-200">
            <h3 className="text-lg font-semibold text-gray-900">Order Items</h3>
          </div>
          <div className="p-6">
            {lineItems.length > 0 ? (
              <div className="space-y-4">
                {lineItems.map((item: any, index: number) => (
                  <div key={index} className="flex items-center justify-between p-4 bg-gray-50 rounded-lg border border-gray-200">
                    <div className="flex items-center gap-4">
                      <div className="w-12 h-12 bg-gray-700 rounded-lg flex items-center justify-center">
                        <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 11V7a4 4 0 00-8 0v4M5 9h14l1 12H4L5 9z" />
                        </svg>
                      </div>
                      <div>
                        <h4 className="font-semibold text-gray-900">{item.title}</h4>
                        <p className="text-sm text-gray-600">SKU: {item.sku}</p>
                      </div>
                    </div>
                    <div className="text-right">
                      <p className="text-sm text-gray-600">Qty: {item.quantity}</p>
                      <p className="font-semibold text-gray-900">
                        {order.currency} {item.price ? item.price.toFixed(2) : '0.00'}
                      </p>
                      <p className="text-sm text-gray-600">
                        Total: {order.currency} {item.total ? item.total.toFixed(2) : '0.00'}
                      </p>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-center py-8">
                <p className="text-gray-500">No items found in this order</p>
              </div>
            )}
          </div>
        </div>

        {/* Shipping and Billing Addresses */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <div className="bg-white rounded-xl shadow-lg border-0 p-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Shipping Address</h3>
            <div className="space-y-2">
              <p className="text-gray-900 font-medium">{order.shipping_name || order.customer_name}</p>
              {order.shipping_phone && <p className="text-gray-600">{order.shipping_phone}</p>}
              {order.shipping_email && <p className="text-gray-600">{order.shipping_email}</p>}
              <p className="text-gray-600">{formatAddress(order)}</p>
            </div>
          </div>

          <div className="bg-white rounded-xl shadow-lg border-0 p-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Billing Address</h3>
            <div className="space-y-2">
              <p className="text-gray-900 font-medium">{order.billing_name || order.customer_name}</p>
              {order.billing_phone && <p className="text-gray-600">{order.billing_phone}</p>}
              {order.billing_email && <p className="text-gray-600">{order.billing_email}</p>}
              <p className="text-gray-600">{formatBillingAddress(order)}</p>
            </div>
          </div>
        </div>

        {/* Price Breakdown */}
        <div className="bg-white rounded-xl shadow-lg border-0 p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-4">Price Breakdown</h3>
          <div className="space-y-3">
            <div className="flex justify-between items-center">
              <span className="text-gray-600">Subtotal</span>
              <span className="text-gray-900 font-medium">
                {order.currency} {order.subtotal ? order.subtotal.toFixed(2) : '0.00'}
              </span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-gray-600">Tax Amount</span>
              <span className="text-gray-900 font-medium">
                {order.currency} {order.tax_amount ? order.tax_amount.toFixed(2) : '0.00'}
              </span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-gray-600">Shipping Amount</span>
              <span className="text-gray-900 font-medium">
                {order.currency} {order.shipping_amount ? order.shipping_amount.toFixed(2) : '0.00'}
              </span>
            </div>
            {(order.discount_amount && order.discount_amount > 0) && (
              <div className="flex justify-between items-center">
                <span className="text-gray-600">Discount Amount</span>
                <span className="text-green-600 font-medium">
                  -{order.currency} {order.discount_amount.toFixed(2)}
                </span>
              </div>
            )}
            <div className="border-t border-gray-200 pt-3">
              <div className="flex justify-between items-center">
                <span className="text-lg font-semibold text-gray-900">Total Amount</span>
                <span className="text-xl font-bold text-green-600">
                  {order.currency} {order.total_price ? order.total_price.toFixed(2) : '0.00'}
                </span>
              </div>
            </div>
          </div>
        </div>

        {/* Tracking Information */}
        {(order.tracking_number || order.tracking_provider || order.estimated_delivery) && (
          <div className="bg-white rounded-xl shadow-lg border-0 p-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Tracking Information</h3>
            <div className="space-y-3">
              {order.tracking_number && (
                <div className="flex justify-between items-center">
                  <span className="text-gray-600">Tracking Number</span>
                  <span className="text-gray-900 font-medium">{order.tracking_number}</span>
                </div>
              )}
              {order.tracking_provider && (
                <div className="flex justify-between items-center">
                  <span className="text-gray-600">Tracking Provider</span>
                  <span className="text-gray-900 font-medium">{order.tracking_provider}</span>
                </div>
              )}
              {order.estimated_delivery && (
                <div className="flex justify-between items-center">
                  <span className="text-gray-600">Estimated Delivery</span>
                  <span className="text-gray-900 font-medium">
                    {new Date(order.estimated_delivery).toLocaleDateString()}
                  </span>
                </div>
              )}
            </div>
          </div>
        )}

        {/* Notes */}
        {order.notes && (
          <div className="bg-white rounded-xl shadow-lg border-0 p-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Order Notes</h3>
            <p className="text-gray-600 bg-gray-50 p-4 rounded-lg border border-gray-200">
              {order.notes}
            </p>
          </div>
        )}
      </div>
    </LoadingState>
  );
}
