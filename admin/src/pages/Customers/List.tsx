import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { api } from '../../lib/api';
import type { Customer, CustomerOrder } from '../../types';
import { LoadingState, LoadingButton } from '../../components/LoadingSpinner';

interface ListResponse {
  data?: Customer[];
  next_cursor?: string;
}

interface OrdersResponse {
  data: CustomerOrder[];
  next_cursor?: string;
}

export function CustomersListPage() {
  const [cursor, setCursor] = useState<string | undefined>();
  const [limit] = useState(50);
  const [expandedCustomer, setExpandedCustomer] = useState<string | null>(null);
  const [ordersCursor, setOrdersCursor] = useState<Record<string, string | undefined>>({});

  const { data, isLoading, error } = useQuery<ListResponse>({
    queryKey: ['customers', cursor, limit],
    queryFn: () => {
      const url = cursor ? `/customers?cursor=${cursor}&limit=${limit}` : `/customers?limit=${limit}`;
      return api.get<ListResponse>(url).then(r => r.data);
    },
  });

  // Handle both response formats: direct array or wrapped in data property
  const customers: Customer[] = Array.isArray(data) ? data : (data && 'data' in data && Array.isArray(data.data) ? data.data : []);

  // Query for customer orders when a customer is expanded
  const { data: ordersData, isLoading: ordersLoading } = useQuery<OrdersResponse>({
    queryKey: ['customer-orders', expandedCustomer, ordersCursor[expandedCustomer || '']],
    queryFn: () => {
      if (!expandedCustomer) return Promise.resolve({ data: [] });
      const cursor = ordersCursor[expandedCustomer || ''];
      const url = cursor ? `/customers/${expandedCustomer}/orders?cursor=${cursor}&limit=20` : `/customers/${expandedCustomer}/orders?limit=20`;
      return api.get<OrdersResponse>(url).then(r => r.data);
    },
    enabled: !!expandedCustomer,
  });

  const queryClient = useQueryClient();

  const handleCustomerClick = (customerId: string) => {
    if (expandedCustomer === customerId) {
      setExpandedCustomer(null);
    } else {
      setExpandedCustomer(customerId);
      setOrdersCursor(prev => ({ ...prev, [customerId]: undefined }));
    }
  };

  const loadMoreOrders = () => {
    if (ordersData?.next_cursor && expandedCustomer) {
      setOrdersCursor(prev => ({ ...prev, [expandedCustomer]: ordersData.next_cursor }));
    }
  };

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.delete(`/customers/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['customers'] });
    },
  });

  const handleDelete = (id: string) => {
    if (confirm('Are you sure you want to delete this customer?')) {
      deleteMutation.mutate(id);
    }
  };

  const loadMore = () => {
    if (data && typeof data === 'object' && data !== null && 'next_cursor' in data && data.next_cursor) {
      setCursor(data.next_cursor);
    }
  };

  return (
    <LoadingState isLoading={isLoading && !data} error={error}>
      <div className="p-6">
        <div className="flex justify-between items-center mb-8">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Customers</h1>
            <p className="text-gray-600 mt-1">Manage your customer base and view order history</p>
          </div>
          <div className="flex items-center gap-2 text-sm text-gray-500">
            <span className="px-3 py-1 bg-blue-100 text-blue-800 rounded-full font-medium">
              {customers?.length || 0} Total
            </span>
          </div>
        </div>

      {customers && customers.length > 0 ? (
        <>
          <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            {customers.map((customer: Customer) => (
              <div key={customer.id} className="bg-white rounded-xl border border-gray-200 shadow-sm hover:shadow-md transition-all duration-200 overflow-hidden">
                <div 
                  className="p-6 cursor-pointer hover:bg-gray-50 transition-colors"
                  onClick={() => handleCustomerClick(customer.id)}
                >
                  <div className="flex justify-between items-start mb-4">
                    <div className="flex items-center gap-3">
                      <div className="w-12 h-12 bg-gradient-to-br from-blue-500 to-purple-600 rounded-full flex items-center justify-center text-white font-semibold text-lg">
                        {customer.full_name.charAt(0).toUpperCase()}
                      </div>
                      <div>
                        <h3 className="font-semibold text-lg text-gray-900">
                          {customer.full_name || 
                           (customer.first_name && customer.last_name
                            ? `${customer.first_name} ${customer.last_name}`
                            : customer.email)}
                        </h3>
                        <p className="text-gray-600 text-sm">{customer.email}</p>
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      {customer.is_active ? (
                        <span className="px-2 py-1 text-xs bg-green-100 text-green-800 rounded-full font-medium">
                          Active
                        </span>
                      ) : (
                        <span className="px-2 py-1 text-xs bg-red-100 text-red-800 rounded-full font-medium">
                          Inactive
                        </span>
                      )}
                    </div>
                  </div>
                  
                  <div className="flex items-center justify-between">
                    <div className="flex gap-6 text-sm">
                      <div>
                        <span className="text-gray-500">Orders</span>
                        <p className="font-semibold text-gray-900">{customer.order_count}</p>
                      </div>
                      <div>
                        <span className="text-gray-500">Member Since</span>
                        <p className="font-semibold text-gray-900">
                          {new Date(customer.created_at).toLocaleDateString('en-US', { 
                            month: 'short', 
                            day: 'numeric', 
                            year: 'numeric' 
                          })}
                        </p>
                      </div>
                    </div>
                    <svg 
                      className={`w-5 h-5 text-gray-400 transition-transform duration-200 ${
                        expandedCustomer === customer.id ? 'rotate-180' : ''
                      }`}
                      fill="none" 
                      stroke="currentColor" 
                      viewBox="0 0 24 24"
                    >
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                    </svg>
                  </div>
                </div>

                {/* Orders Section - Expandable */}
                {expandedCustomer === customer.id && (
                  <div className="border-t border-gray-200 p-6 bg-gray-50">
                    <div className="flex justify-between items-center mb-6">
                      <h4 className="font-semibold text-gray-900 text-lg">
                        Order History ({customer.order_count})
                      </h4>
                      <Link
                        to={`/orders?customer=${customer.id}`}
                        className="text-sm text-blue-600 hover:text-blue-800 font-medium"
                        onClick={(e) => e.stopPropagation()}
                      >
                        View All Orders →
                      </Link>
                    </div>

                    {ordersLoading ? (
                      <div className="p-8 text-center text-gray-500">
                        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto mb-2"></div>
                        Loading orders...
                      </div>
                    ) : ordersData?.data && ordersData.data.length > 0 ? (
                      <>
                        <div className="space-y-3">
                          {ordersData.data.map((order: CustomerOrder) => (
                            <div 
                              key={order.id} 
                              className="bg-white p-4 rounded-lg border border-gray-200 hover:border-gray-300 transition-colors"
                            >
                              <div className="flex justify-between items-start">
                                <div className="flex-1">
                                  <div className="flex items-center gap-3 mb-2">
                                    <span className="font-semibold text-gray-900">#{order.order_number}</span>
                                    <span className={`px-2 py-1 text-xs rounded-full font-medium ${
                                      order.status === 'delivered' ? 'bg-green-100 text-green-800' :
                                      order.status === 'shipped' ? 'bg-blue-100 text-blue-800' :
                                      order.status === 'cancelled' ? 'bg-red-100 text-red-800' :
                                      'bg-yellow-100 text-yellow-800'
                                    }`}>
                                      {order.status}
                                    </span>
                                  </div>
                                  <div className="flex gap-4 text-sm text-gray-600">
                                    <span>{order.item_count} item{order.item_count !== 1 ? 's' : ''}</span>
                                    <span>{new Date(order.placed_at).toLocaleDateString()}</span>
                                  </div>
                                </div>
                                <div className="text-right">
                                  <div className="font-semibold text-gray-900">
                                    ${(order.total_cents / 100).toFixed(2)}
                                  </div>
                                  {order.refund_cents > 0 && (
                                    <div className="text-sm text-red-600">
                                      Refunded: ${(order.refund_cents / 100).toFixed(2)}
                                    </div>
                                  )}
                                </div>
                              </div>
                            </div>
                          ))}
                        </div>

                        {ordersData.next_cursor && (
                          <div className="mt-6 text-center">
                            <button
                              onClick={loadMoreOrders}
                              disabled={ordersLoading}
                              className="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 transition disabled:opacity-50 font-medium"
                            >
                              {ordersLoading ? 'Loading...' : 'Load More Orders'}
                            </button>
                          </div>
                        )}
                      </>
                    ) : (
                      <div className="text-center py-8 text-gray-500">
                        <svg className="w-12 h-12 text-gray-300 mx-auto mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
                        </svg>
                        <p>No orders found for this customer.</p>
                      </div>
                    )}
                  </div>
                )}
              </div>
            ))}
          </div>

          {data && typeof data === 'object' && data !== null && 'next_cursor' in data && data.next_cursor && (
            <div className="mt-8 text-center">
              <button
                onClick={loadMore}
                disabled={isLoading}
                className="bg-blue-600 text-white px-6 py-3 rounded-lg hover:bg-blue-700 transition disabled:opacity-50 font-medium"
              >
                {isLoading ? 'Loading...' : 'Load More Customers'}
              </button>
            </div>
          )}
        </>
      ) : (
        <div className="text-center py-16">
          <svg className="w-16 h-16 text-gray-300 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
          </svg>
          <h3 className="text-lg font-medium text-gray-900 mb-2">No customers found</h3>
          <p className="text-gray-600">Get started by adding your first customer.</p>
        </div>
      )}
      </div>
    </LoadingState>
  );
}
