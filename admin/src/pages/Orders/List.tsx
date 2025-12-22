import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { api } from '../../lib/api';
import type { Order } from '../../types';
import { LoadingState, LoadingButton } from '../../components/LoadingSpinner';
import { SHIPPING_STATUS_OPTIONS, SHIPPING_STATUS } from '../../constants/shippingStatus';

interface ListResponse {
  data: Order[];
  next_cursor?: string;
}

export function OrdersListPage() {
  const [cursor, setCursor] = useState<string | undefined>();
  const [limit] = useState(50);
  const [shippingStatusFilter, setShippingStatusFilter] = useState<string>('');

  const { data, isLoading, error } = useQuery<Order[]>({
    queryKey: ['orders', cursor, limit, shippingStatusFilter],
    queryFn: () => {
      const params = new URLSearchParams();
      if (cursor) params.append('cursor', cursor);
      params.append('limit', limit.toString());
      if (shippingStatusFilter) params.append('shipping_status', shippingStatusFilter);
      
      const url = `/orders?${params.toString()}`;
      return api.get<{data: Order[]}>(url).then((r: any) => r.data);
    },
  });

  const queryClient = useQueryClient();

  const updateStatusMutation = useMutation({
    mutationFn: ({ id, status }: { id: string; status: string }) => 
      api.put(`/orders/${id}`, { status }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['orders'] });
    },
  });

  const updateShippingStatusMutation = useMutation({
    mutationFn: ({ id, shipping_status }: { id: string; shipping_status: string }) => 
      api.put(`/orders/${id}`, { shipping_status }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['orders'] });
    },
  });

  const handleStatusUpdate = (id: string, newStatus: string) => {
    if (confirm(`Mark order as ${newStatus}?`)) {
      updateStatusMutation.mutate({ id, status: newStatus });
    }
  };

  const handleShippingStatusUpdate = (id: string, newShippingStatus: string) => {
    if (confirm(`Update shipping status to ${newShippingStatus}?`)) {
      updateShippingStatusMutation.mutate({ id, shipping_status: newShippingStatus });
    }
  };

  const handleDelete = (id: string) => {
    if (confirm('Are you sure you want to delete this order?')) {
      deleteMutation.mutate(id);
    }
  };

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.delete(`/orders/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['orders'] });
    },
  });

  const loadMore = () => {
    // Since we're getting the array directly, there's no next_cursor
    // This would need to be implemented in the backend if pagination is needed
  };

  return (
    <LoadingState isLoading={isLoading && !data} error={error}>
      <div className="p-6">
      
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">Orders</h1>
        <Link
          to="new"
          className="bg-gold text-white px-4 py-2 rounded hover:bg-yellow-600 transition"
        >
          New Order
        </Link>
      </div>

      {/* Filter Section */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4 mb-6">
        <div className="flex items-center gap-4">
          <div className="flex-1 max-w-xs">
            <label htmlFor="shipping-status-filter" className="block text-sm font-medium text-gray-700 mb-2">
              Filter by Shipping Status
            </label>
            <select
              id="shipping-status-filter"
              value={shippingStatusFilter}
              onChange={(e) => setShippingStatusFilter(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            >
              <option value="">All Orders</option>
              {SHIPPING_STATUS_OPTIONS.map((status) => (
                <option key={status.value} value={status.value}>
                  {status.label}
                </option>
              ))}
            </select>
          </div>
          {shippingStatusFilter && (
            <button
              onClick={() => setShippingStatusFilter('')}
              className="px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 transition-colors mt-6"
            >
              Clear Filter
            </button>
          )}
        </div>
      </div>

      {data && data.length > 0 ? (
        <>
          <div className="grid gap-4">
            {data.map((order) => (
              <div key={order.id} className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden hover:shadow-md transition-shadow">
                {/* Order Header */}
                <div className="bg-gradient-to-r from-gray-50 to-gray-100 px-6 py-4 border-b border-gray-200">
                  <div className="flex justify-between items-start">
                    <div>
                      <h3 className="text-lg font-bold text-gray-900">{order.order_number}</h3>
                      <div className="flex items-center gap-3 mt-2">
                        <span className={`inline-flex items-center px-3 py-1 rounded-full text-xs font-semibold ${statusBadgeColor(order.status)}`}>
                          {order.status.replace('_', ' ').toUpperCase()}
                        </span>
                        <span className="text-sm text-gray-500">
                          {new Date(order.created_at).toLocaleDateString()}
                        </span>
                      </div>
                    </div>
                    <div className="text-right">
                      <div className="text-2xl font-bold text-gray-900">
                        {order.currency} {order.total_price.toFixed(2)}
                      </div>
                      <div className="text-sm text-gray-500">
                        {order.line_items?.length || 0} items
                      </div>
                    </div>
                  </div>
                </div>
                
                {/* Order Details */}
                <div className="p-6">
                  <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                    {/* Customer Info */}
                    <div>
                      <h4 className="text-sm font-semibold text-gray-700 mb-3">Customer Information</h4>
                      <div className="space-y-2">
                        <div className="flex items-center gap-2">
                          <div className="w-2 h-2 bg-blue-500 rounded-full"></div>
                          <span className="text-sm font-medium text-gray-900">{order.customer_name}</span>
                        </div>
                        <div className="flex items-center gap-2">
                          <div className="w-2 h-2 bg-green-500 rounded-full"></div>
                          <span className="text-sm text-gray-600">{order.customer_email}</span>
                        </div>
                        <div className="flex items-center gap-2">
                          <div className="w-2 h-2 bg-purple-500 rounded-full"></div>
                          <span className="text-sm text-gray-600">{order.customer_phone}</span>
                        </div>
                        {order.user_id && (
                          <div className="flex items-center gap-2">
                            <div className="w-2 h-2 bg-orange-500 rounded-full"></div>
                            <span className="text-sm text-gray-600">User ID: {order.user_id}</span>
                          </div>
                        )}
                      </div>
                    </div>
                    
                    {/* Shipping Status */}
                    <div>
                      <h4 className="text-sm font-semibold text-gray-700 mb-3">Shipping Status</h4>
                      <div className="space-y-2">
                        <div className="flex items-center gap-2">
                          <div className="w-2 h-2 bg-indigo-500 rounded-full"></div>
                          <span className="text-sm font-medium text-gray-900">
                            {order.shipping_status ? order.shipping_status.replace('_', ' ').toUpperCase() : 'JUST ARRIVED'}
                          </span>
                        </div>
                        <select
                          value={order.shipping_status || SHIPPING_STATUS.JUST_ARRIVED}
                          onChange={(e) => handleShippingStatusUpdate(order.id, e.target.value)}
                          className="mt-2 px-3 py-2 text-sm border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 w-full"
                        >
                          {SHIPPING_STATUS_OPTIONS.map((status) => (
                            <option key={status.value} value={status.value}>
                              {status.label}
                            </option>
                          ))}
                        </select>
                      </div>
                    </div>
                    
                    {/* Order Summary */}
                    <div>
                      <h4 className="text-sm font-semibold text-gray-700 mb-3">Order Summary</h4>
                      <div className="space-y-2">
                        <div className="flex justify-between text-sm">
                          <span className="text-gray-600">Subtotal:</span>
                          <span className="font-medium">{order.currency} {order.subtotal.toFixed(2)}</span>
                        </div>
                        <div className="flex justify-between text-sm">
                          <span className="text-gray-600">Tax:</span>
                          <span className="font-medium">{order.currency} {order.tax_amount.toFixed(2)}</span>
                        </div>
                        <div className="flex justify-between text-sm">
                          <span className="text-gray-600">Shipping:</span>
                          <span className="font-medium">{order.currency} {order.shipping_amount.toFixed(2)}</span>
                        </div>
                        <div className="flex justify-between text-sm">
                          <span className="text-gray-600">Discount:</span>
                          <span className="font-medium text-red-600">-{order.currency} {order.discount_amount.toFixed(2)}</span>
                        </div>
                        <div className="pt-2 border-t border-gray-200 flex justify-between">
                          <span className="font-semibold text-gray-900">Total:</span>
                          <span className="font-bold text-lg text-gray-900">{order.currency} {order.total_price.toFixed(2)}</span>
                        </div>
                      </div>
                    </div>
                  </div>
                  
                  {/* Action Buttons */}
                  <div className="mt-6 pt-6 border-t border-gray-200">
                    <div className="flex flex-wrap gap-3">
                      {/* Status Update Buttons */}
                      <div className="flex gap-2">
                        {order.status !== SHIPPING_STATUS.JUST_ARRIVED && order.status !== SHIPPING_STATUS.PROCESSING && order.status !== SHIPPING_STATUS.SHIPPED && order.status !== SHIPPING_STATUS.DELIVERED && (
                          <button
                            onClick={() => handleStatusUpdate(order.id, SHIPPING_STATUS.JUST_ARRIVED)}
                            className="px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700 transition-colors"
                            disabled={updateStatusMutation.isPending}
                          >
                            Mark as Just Arrived
                          </button>
                        )}
                        {order.status !== SHIPPING_STATUS.PROCESSING && order.status !== SHIPPING_STATUS.SHIPPED && order.status !== SHIPPING_STATUS.DELIVERED && (
                          <button
                            onClick={() => handleStatusUpdate(order.id, SHIPPING_STATUS.PROCESSING)}
                            className="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors"
                            disabled={updateStatusMutation.isPending}
                          >
                            Mark as Packed
                          </button>
                        )}
                        {order.status !== SHIPPING_STATUS.SHIPPED && order.status !== SHIPPING_STATUS.DELIVERED && (
                          <button
                            onClick={() => handleStatusUpdate(order.id, SHIPPING_STATUS.SHIPPED)}
                            className="px-4 py-2 bg-purple-600 text-white text-sm font-medium rounded-lg hover:bg-purple-700 transition-colors"
                            disabled={updateStatusMutation.isPending}
                          >
                            Mark as Shipped
                          </button>
                        )}
                        {order.status !== SHIPPING_STATUS.DELIVERED && (
                          <button
                            onClick={() => handleStatusUpdate(order.id, SHIPPING_STATUS.DELIVERED)}
                            className="px-4 py-2 bg-green-600 text-white text-sm font-medium rounded-lg hover:bg-green-700 transition-colors"
                            disabled={updateStatusMutation.isPending}
                          >
                            Mark as Delivered
                          </button>
                        )}
                      </div>
                      
                      {/* Action Links */}
                      <div className="flex gap-2 ml-auto">
                        <Link
                          to={order.id}
                          className="px-4 py-2 bg-gray-100 text-gray-700 text-sm font-medium rounded-lg hover:bg-gray-200 transition-colors"
                        >
                          View Details
                        </Link>
                        <Link
                          to={`${order.id}/edit`}
                          className="px-4 py-2 bg-blue-100 text-blue-700 text-sm font-medium rounded-lg hover:bg-blue-200 transition-colors flex items-center gap-2"
                        >
                          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                          </svg>
                          Edit
                        </Link>
                        <LoadingButton
                          onClick={() => handleDelete(order.id)}
                          className="px-4 py-2 bg-red-100 text-red-700 text-sm font-medium rounded-lg hover:bg-red-200 transition-colors"
                          isLoading={deleteMutation.isPending}
                          loadingText="Deleting..."
                        >
                          Delete
                        </LoadingButton>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </>
      ) : (
        <div className="text-center py-12 text-gray-500">
          No orders yet. <Link to="new" className="text-gold hover:underline">Create the first order</Link>.
        </div>
      )}
      </div>
    </LoadingState>
  );
}

function statusBadgeColor(status: string): string {
  switch (status) {
    case 'pending': return 'bg-yellow-100 text-yellow-800';
    case 'pending_payment': return 'bg-orange-100 text-orange-800';
    case SHIPPING_STATUS.JUST_ARRIVED: return 'bg-indigo-100 text-indigo-800';
    case SHIPPING_STATUS.PROCESSING: return 'bg-blue-100 text-blue-800';
    case SHIPPING_STATUS.SHIPPED: return 'bg-purple-100 text-purple-800';
    case SHIPPING_STATUS.DELIVERED: return 'bg-green-100 text-green-800';
    case SHIPPING_STATUS.CANCELLED: return 'bg-red-100 text-red-800';
    case 'paid': return 'bg-green-100 text-green-800';
    default: return 'bg-gray-100 text-gray-800';
  }
}
