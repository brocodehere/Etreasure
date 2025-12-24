import React from 'react';
import { useQuery, useMutation } from '@apollo/client';
import { Link } from 'react-router-dom';
import type { Order } from '../types';
import { LoadingState, LoadingSpinner } from '../components/LoadingSpinner';
import { GET_DASHBOARD_STATS, GET_JUST_ARRIVED_ORDERS } from '../graphql/queries';
import { api } from '../lib/api';
import { SHIPPING_STATUS_OPTIONS, SHIPPING_STATUS } from '../constants/shippingStatus';

interface DashboardStats {
  totalProducts: number;
  totalCategories: number;
  totalBanners: number;
  activeBanners: number;
  totalOffers: number;
  activeOffers: number;
  recentOrders: number;
  totalRevenue: number;
}

interface RecentActivity {
  id: string;
  type: 'product' | 'category' | 'banner' | 'offer';
  action: 'created' | 'updated' | 'deleted';
  title: string;
  timestamp: string;
}

export const DashboardPage: React.FC = () => {
  // Use GraphQL to fetch all dashboard data in a single request
  const { data: dashboardData, loading: dashboardLoading, error: dashboardError } = useQuery(GET_DASHBOARD_STATS);
  
  // Use GraphQL to fetch just arrived orders
  const { data: ordersData, loading: ordersLoading, error: ordersError, refetch: refetchOrders } = useQuery(GET_JUST_ARRIVED_ORDERS);

  // Function to update shipping status
  const updateShippingStatus = async (orderId: string, newStatus: string) => {
    try {
      await api.put(`/orders/${orderId}`, { shipping_status: newStatus });
      // Refetch the orders to get updated data
      refetchOrders();
    } catch (error) {
          }
  };

  // Extract stats from GraphQL response
  const stats = React.useMemo(() => {
    return dashboardData?.dashboardStats || {
      totalProducts: 0,
      totalCategories: 0,
      totalBanners: 0,
      activeBanners: 0,
      totalOffers: 0,
      activeOffers: 0,
      recentOrders: 0,
      totalRevenue: 0,
    };
  }, [dashboardData]);

  // Get just arrived orders from GraphQL response
  const justArrivedOrders = React.useMemo(() => {
    return ordersData?.justArrivedOrders || [];
  }, [ordersData]);

  // Check if any queries are loading
  const isLoading = dashboardLoading || ordersLoading;
  const error = dashboardError || ordersError;

  const StatCard = ({ title, value, icon, color, link }: {
    title: string;
    value: string | number;
    icon: React.ReactNode;
    color: string;
    link: string;
  }) => (
    <Link to={link} className="block">
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 hover:shadow-md transition-shadow duration-200">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-sm font-medium text-gray-600">{title}</p>
            <p className="text-2xl font-bold text-gray-900 mt-1">{value}</p>
          </div>
          <div className={`p-3 rounded-full ${color}`}>
            {icon}
          </div>
        </div>
      </div>
    </Link>
  );

  return (
    <LoadingState isLoading={isLoading} error={error}>
      <div className="space-y-6">
      <header>
        <h1 className="text-3xl font-bold text-gray-900">Dashboard</h1>
        <p className="text-gray-600 mt-1">
          Welcome back! Here's an overview of your store performance.
        </p>
      </header>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <StatCard
          title="Total Products"
          value={stats.totalProducts}
          icon={
            <svg className="w-6 h-6 text-orange-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
            </svg>
          }
          color="bg-orange-100"
          link="/products"
        />
        <StatCard
          title="Categories"
          value={stats.totalCategories}
          icon={
            <svg className="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
            </svg>
          }
          color="bg-blue-100"
          link="/categories"
        />
        <StatCard
          title="Active Banners"
          value={stats.activeBanners}
          icon={
            <svg className="w-6 h-6 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
          }
          color="bg-purple-100"
          link="/banners"
        />
        <StatCard
          title="Active Offers"
          value={stats.activeOffers}
          icon={
            <svg className="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1" />
            </svg>
          }
          color="bg-green-100"
          link="/offers"
        />
      </div>

      {/* Revenue and Orders */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-4">Recent Performance</h3>
          <div className="space-y-4">
            <div className="flex justify-between items-center">
              <span className="text-gray-600">Recent Orders</span>
              <span className="text-2xl font-bold text-gray-900">{stats.recentOrders}</span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-gray-600">Total Revenue</span>
              <span className="text-2xl font-bold text-green-600">
                ₹ {stats.totalRevenue.toLocaleString("en-IN")}
              </span>
            </div>
          </div>
        </div>

        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-4">Quick Actions</h3>
          <div className="grid grid-cols-2 gap-3">
            <Link
              to="/products/new"
              className="flex items-center justify-center px-4 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
            >
              <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
              </svg>
              Add Product
            </Link>
            <Link
              to="/offers/new"
              className="flex items-center justify-center px-4 py-3 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors"
            >
              <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2" />
              </svg>
              Create Offer
            </Link>
            <Link
              to="/banners/new"
              className="flex items-center justify-center px-4 py-3 bg-purple-600 text-white rounded-lg hover:bg-purple-700 transition-colors"
            >
              <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
              </svg>
              Add Banner
            </Link>
            <Link
              to="/categories/new"
              className="flex items-center justify-center px-4 py-3 bg-yellow-600 text-white rounded-lg hover:bg-yellow-700 transition-colors"
            >
              <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
              </svg>
              Add Category
            </Link>
          </div>
        </div>
      </div>

      {/* Just Arrived Orders */}
        <div className="bg-white rounded-xl shadow-lg border-0 overflow-hidden transform hover:scale-[1.01] transition-transform duration-300">
          <div className="bg-gradient-to-r from-indigo-600 via-purple-600 to-pink-600 px-6 py-5 relative overflow-hidden">
            <div className="absolute inset-0 bg-white/10 backdrop-blur-sm"></div>
            <div className="relative z-10">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div className="p-3 bg-gray-800/50 rounded-xl backdrop-blur-sm border border-white/30 shadow-lg">
                    <svg className="w-7 h-7 text-white animate-pulse" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
                    </svg>
                  </div>
                  <div>
                    <h3 className="text-xl font-bold text-gray-800">Just Arrived Orders</h3>
                    <p className="text-gray-600 text-sm mt-1">New orders awaiting your attention</p>
                  </div>
                </div>
                <Link
                  to="/orders"
                  className="text-sm text-white/90 hover:text-white font-semibold flex items-center gap-2 transition-all duration-200 bg-white/20 px-4 py-2 rounded-lg backdrop-blur-sm hover:bg-white/30 border border-white/30"
                >
                  View All Orders
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                  </svg>
                </Link>
              </div>
            </div>
            <div className="absolute -bottom-2 -right-2 w-20 h-20 bg-white/5 rounded-full blur-xl"></div>
            <div className="absolute -top-2 -left-2 w-16 h-16 bg-white/5 rounded-full blur-xl"></div>
          </div>
          <div className="p-6 bg-gradient-to-b from-gray-50 to-white">
            {justArrivedOrders && justArrivedOrders.length > 0 ? (
              <div className="space-y-4">
                {justArrivedOrders.map((order: Order, index: number) => (
                  <div 
                    key={order.id} 
                    className="group relative bg-gradient-to-r from-indigo-50 via-purple-50 to-pink-50 border border-indigo-200 rounded-xl p-5 hover:shadow-xl transition-all duration-300 hover:-translate-y-1 hover:scale-[1.02] animate-fadeIn"
                    style={{ animationDelay: `${index * 100}ms` }}
                  >
                    <div className="absolute top-3 right-3">
                      <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-bold bg-gradient-to-r from-indigo-500 to-purple-600 text-white shadow-lg animate-pulse">
                        {order.shipping_status || 'just arrived'}
                      </span>
                    </div>
                    <div className="flex-1">
                      <div className="mb-3">
                        {order.line_items ? (
                          <div className="space-y-2">
                            {JSON.parse(order.line_items as unknown as string).map((item: any, itemIndex: number) => (
                              <div key={itemIndex} className="flex items-center justify-between bg-white/70 backdrop-blur-sm rounded-lg px-4 py-3 border border-white/30 shadow-sm hover:shadow-md transition-shadow">
                                <div className="flex items-center gap-3">
                                  <div className="w-10 h-10 bg-gray-700 rounded-xl flex items-center justify-center shadow-lg">
                                    <svg className="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 11V7a4 4 0 00-8 0v4M5 9h14l1 12H4L5 9z" />
                                    </svg>
                                  </div>
                                  <div>
                                    <span className="font-bold text-gray-900 text-sm">{item.title}</span>
                                    <div className="flex items-center gap-2 mt-1">
                                      <span className="text-xs bg-indigo-100 text-indigo-700 px-2 py-1 rounded-full font-semibold border border-indigo-200">
                                        Qty: {item.quantity}
                                      </span>
                                      {item.price && (
                                        <span className="text-xs bg-green-100 text-green-700 px-2 py-1 rounded-full font-semibold border border-green-200">
                                          {order.currency} {item.price.toFixed(2)}
                                        </span>
                                      )}
                                    </div>
                                  </div>
                                </div>
                              </div>
                            ))}
                          </div>
                        ) : (
                          <div className="flex items-center gap-2 text-gray-500 bg-gray-50 rounded-lg p-3">
                            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
                            </svg>
                            <span className="text-sm font-medium">No items found</span>
                          </div>
                        )}
                      </div>
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-6 text-sm">
                          <div className="flex items-center gap-2 bg-white/60 px-3 py-2 rounded-lg border border-white/30">
                            <svg className="w-4 h-4 text-indigo-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                            </svg>
                            <span className="font-semibold text-gray-900">{order.customer_name}</span>
                          </div>
                          <div className="flex items-center gap-2 bg-white/60 px-3 py-2 rounded-lg border border-white/30">
                            <svg className="w-4 h-4 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              {/* <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1" /> */}
                            </svg>
                            <span className="font-bold text-green-600">{order.currency} {order.total_price.toFixed(2)}</span>
                          </div>
                          <div className="flex items-center gap-2 bg-white/60 px-3 py-2 rounded-lg border border-white/30">
                            <svg className="w-4 h-4 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                            </svg>
                            <span className="text-gray-700 font-medium">{new Date(order.created_at.replace(' IST', '+05:30')).toLocaleString()}</span>
                          </div>
                        </div>
                      </div>
                    </div>
                    <div className="flex items-center justify-between mt-4 pt-4 border-t border-indigo-200">
                      <div className="w-48">
                        <select
                          value={order.shipping_status || SHIPPING_STATUS.JUST_ARRIVED}
                          onChange={(e) => updateShippingStatus(order.id, e.target.value)}
                          className="w-full px-3 py-2 text-sm border border-indigo-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent bg-white/90 backdrop-blur-sm font-medium shadow-sm"
                        >
                          {SHIPPING_STATUS_OPTIONS.map((status) => (
                            <option key={status.value} value={status.value}>
                              {status.label}
                            </option>
                          ))}
                        </select>
                      </div>
                      <div className="flex gap-2">
                        <Link
                          to={`/orders/${order.id}`}
                          className="px-4 py-2 bg-gray-800 text-white text-sm font-bold rounded-lg hover:bg-gray-900 transition-all duration-300 shadow-lg hover:shadow-xl flex items-center gap-2 transform hover:scale-105"
                        >
                          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                          </svg>
                          View Details
                        </Link>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-center py-16 bg-gradient-to-b from-gray-50 to-white rounded-xl">
                <div className="w-20 h-20 bg-gradient-to-br from-gray-100 to-gray-200 rounded-full flex items-center justify-center mx-auto mb-6 shadow-lg">
                  <svg className="w-10 h-10 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
                  </svg>
                </div>
                <h4 className="text-xl font-bold text-gray-900 mb-3">No orders marked as 'Just Arrived'</h4>
                <p className="text-gray-600 font-medium">Orders with 'just arrived' shipping status will appear here</p>
              </div>
            )}
          </div>
        </div>
      </div>

        </LoadingState>
  );
};
