import React, { useEffect, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { api } from '../lib/api';
import type { Order } from '../types';

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
  const [stats, setStats] = useState<DashboardStats>({
    totalProducts: 0,
    totalCategories: 0,
    totalBanners: 0,
    activeBanners: 0,
    totalOffers: 0,
    activeOffers: 0,
    recentOrders: 0,
    totalRevenue: 0,
  });

  const [recentActivity, setRecentActivity] = useState<RecentActivity[]>([]);

  // Fetch just arrived orders
  const { data: justArrivedOrders } = useQuery<{data: Order[]}>({
    queryKey: ['just-arrived-orders'],
    queryFn: () => api.get('/orders?limit=10').then((r: any) => r.data),
  });

  useEffect(() => {
    // Fetch dashboard statistics
    const fetchStats = async () => {
      try {
        // Fetch real data from all endpoints
        const [productsRes, categoriesRes, bannersRes, offersRes, ordersRes] = await Promise.all([
          api.get<{items: any[]}>('/products?limit=100').catch(() => ({ items: [] })),
          api.get<{items: any[]}>('/categories?limit=100').catch(() => ({ items: [] })),
          api.get<{data: any[]}>('/banners?limit=100').catch(() => ({ data: [] })),
          api.get<{data: any[]}>('/offers?limit=100').catch(() => ({ data: [] })),
          api.get<{data: Order[]}>('/orders?limit=100').catch(() => ({ data: [] })),
        ]);

        const products = productsRes.items?.length || 0;
        const categories = categoriesRes.items?.length || 0;
        const banners = bannersRes.data?.length || 0;
        const offers = offersRes.data?.length || 0;
        const orders = ordersRes.data?.length || 0;
        const activeOffers = offersRes.data?.filter((o: any) => o.is_active).length || 0;
        const activeBanners = bannersRes.data?.filter((b: any) => b.is_active).length || 0;
        const justArrived = ordersRes.data?.filter((o: Order) => o.status === 'just_arrived').length || 0;
        const totalRevenue = ordersRes.data?.reduce((sum: number, o: Order) => sum + o.total_price, 0) || 0;

        setStats({
          totalProducts: products,
          totalCategories: categories,
          totalBanners: banners,
          activeBanners: activeBanners,
          totalOffers: offers,
          activeOffers: activeOffers,
          recentOrders: orders,
          totalRevenue: totalRevenue,
        });
      } catch (error) {
        console.error('Failed to fetch dashboard stats:', error);
        // Set default values on error
        setStats({
          totalProducts: 0,
          totalCategories: 0,
          totalBanners: 0,
          activeBanners: 0,
          totalOffers: 0,
          activeOffers: 0,
          recentOrders: 0,
          totalRevenue: 0,
        });
      }
    };

    fetchStats();
  }, []);

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
              <span className="text-2xl font-bold text-green-600">${stats.totalRevenue.toLocaleString()}</span>
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
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
        <div className="flex justify-between items-center mb-4">
          <h3 className="text-lg font-semibold text-gray-900">Just Arrived Orders</h3>
          <Link
            to="/orders"
            className="text-sm text-indigo-600 hover:text-indigo-800 font-medium"
          >
            View All Orders →
          </Link>
        </div>
        {justArrivedOrders?.data && justArrivedOrders.data.filter((order: Order) => order.status === 'just_arrived').length > 0 ? (
          <div className="space-y-3">
            {justArrivedOrders.data
              .filter((order: Order) => order.status === 'just_arrived')
              .slice(0, 5)
              .map((order: Order) => (
                <div key={order.id} className="flex items-center justify-between p-4 bg-indigo-50 border border-indigo-200 rounded-lg">
                  <div className="flex-1">
                    <div className="flex items-center gap-3">
                      <span className="text-sm font-semibold text-gray-900">{order.order_number}</span>
                      <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-indigo-100 text-indigo-800">
                        Just Arrived
                      </span>
                    </div>
                    <div className="mt-1 flex items-center gap-4 text-sm text-gray-600">
                      <span>{order.customer_name}</span>
                      <span>{order.currency} {order.total_price.toFixed(2)}</span>
                      <span>{new Date(order.created_at).toLocaleDateString()}</span>
                    </div>
                  </div>
                  <Link
                    to={`/orders/${order.id}`}
                    className="px-3 py-1 bg-indigo-600 text-white text-sm font-medium rounded hover:bg-indigo-700 transition-colors"
                  >
                    View Details
                  </Link>
                </div>
              ))}
          </div>
        ) : (
          <div className="text-center py-8 text-gray-500">
            <svg className="w-12 h-12 text-gray-400 mx-auto mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
            </svg>
            <p className="text-sm">No orders marked as "Just Arrived"</p>
            <p className="text-xs text-gray-400 mt-1">Mark orders as "Just Arrived" to see them here</p>
          </div>
        )}
      </div>

      {/* Recent Activity */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
        <h3 className="text-lg font-semibold text-gray-900 mb-4">System Overview</h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="text-center p-4 bg-gray-50 rounded-lg">
            <div className="text-2xl font-bold text-blue-600">{stats.totalProducts}</div>
            <div className="text-sm text-gray-600 mt-1">Total Products</div>
          </div>
          <div className="text-center p-4 bg-gray-50 rounded-lg">
            <div className="text-2xl font-bold text-green-600">{stats.activeOffers}</div>
            <div className="text-sm text-gray-600 mt-1">Active Offers</div>
          </div>
          <div className="text-center p-4 bg-gray-50 rounded-lg">
            <div className="text-2xl font-bold text-purple-600">{stats.totalBanners}</div>
            <div className="text-sm text-gray-600 mt-1">Total Banners</div>
          </div>
        </div>
      </div>
    </div>
  );
};
