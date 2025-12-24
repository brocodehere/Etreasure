import React, { useEffect, useState } from 'react';
import { NavLink, Outlet, useNavigate } from 'react-router-dom';
import { clearTokens, getAccessToken } from '../lib/auth';
import { api } from '../lib/api';
import { useQuery } from '@tanstack/react-query';

// User interface for TypeScript
interface User {
  id: number;
  email: string;
  name: string;
  roles: string[];
}

const navItems = [
  { to: '/', label: 'Dashboard' },
  { to: '/products', label: 'Products' },
  { to: '/categories', label: 'Categories' },
  { to: '/banners', label: 'Banners' },
  { to: '/offers', label: 'Offers' },
  { to: '/media', label: 'Media' },
  { to: '/content', label: 'Content' },
  { to: '/orders', label: 'Orders' },
  { to: '/customers', label: 'Customers' },
  { to: '/settings', label: 'Settings' },
  { to: '/users', label: 'Users & Roles' },
  { to: '/audit-log', label: 'Audit Log' },
];

export const AdminLayout: React.FC = () => {
  const navigate = useNavigate();

  // Fetch current user data
  const { data: user, isLoading, error } = useQuery<User>({
    queryKey: ['currentUser'],
    queryFn: async () => {
      try {
        const response = await api.get<User>('/me');
        return response;
      } catch (err) {
        throw err;
      }
    },
    staleTime: 5 * 60 * 1000, // 5 minutes
  });

  const handleLogout = () => {
    clearTokens();
    navigate('/login');
  };

  // Get primary role (first role in the array)
  const primaryRole = user?.roles?.[0] || 'Admin';
  const displayName = user?.name || user?.email || 'Admin';

  return (
    <div className="app-shell">
      <aside className="app-sidebar">
        <div className="px-6 py-4 border-b border-gold/40">
          <div className="text-lg font-playfair tracking-wide">Etreasure Admin</div>
        </div>
        <nav className="flex-1 overflow-y-auto px-2 py-4 space-y-1">
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              className={({ isActive }) =>
                `block px-3 py-2 rounded-md text-sm font-medium transition-colors ${
                  isActive
                    ? 'app-link-active'
                    : 'text-cream/80 hover:bg-maroon/60 hover:text-cream'
                }`
              }
              aria-label={item.label}
            >
              {item.label}
            </NavLink>
          ))}
        </nav>
        {/* Logout Button */}
        <div className="px-2 py-2 border-t border-gold/40">
          <button
            onClick={handleLogout}
            className="w-full block px-3 py-2 rounded-md text-sm font-medium transition-colors text-red-400 hover:bg-red-900/20 hover:text-red-300"
          >
            Logout
          </button>
        </div>
      </aside>
      <div className="flex flex-col flex-1">
        <header className="app-header">
          <div className="flex items-center gap-2">
            <span className="text-sm uppercase tracking-wide text-maroon">
              {isLoading ? 'Loading...' : primaryRole}
            </span>
          </div>
          <div className="text-sm text-dark/70">
            {isLoading ? 'Loading...' : `Signed in as ${displayName}`}
          </div>
        </header>
        <main className="app-main">
          <Outlet />
        </main>
      </div>
    </div>
  );
};
