import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '../../lib/api';
import { Link } from 'react-router-dom';

type Product = {
  uuid_id: string;
  slug: string;
  title: string;
  published: boolean;
  publish_at?: string | null;
  category_id?: string | null;
};

type ProductsResponse = {
  items: Array<Product>;
  nextCursor?: string;
};

export const ProductsListPage: React.FC = () => {
  const { data, isLoading, error } = useQuery<ProductsResponse>({
    queryKey: ['products', { first: 20 }],
    queryFn: () => api.get<ProductsResponse>('/products?first=20'),
    retry: (failureCount: number, error: any) => {
      // Don't retry on 4xx errors
      if (error?.status >= 400 && error?.status < 500) return false;
      return failureCount < 3;
    },
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
  });

  return (
    <div className="space-y-6">
      <header className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-playfair text-maroon">Products</h1>
          <p className="text-sm text-dark/70 mt-1">Manage products with variants, categories, and images.</p>
        </div>
        <div className="flex items-center gap-2">
          <Link
            to="/products/new"
            className="inline-flex items-center justify-center rounded-md bg-maroon text-cream text-sm font-medium px-3 py-2 hover:bg-maroon/90"
          >
            New Product
          </Link>
        </div>
      </header>

      {isLoading && (
        <div className="flex items-center justify-center py-12">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-maroon"></div>
          <span className="ml-2 text-dark/70">Loading products...</span>
        </div>
      )}

      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4">
          <p className="text-sm text-red-600 font-medium">Error loading products</p>
          <p className="text-xs text-red-500 mt-1">{String((error as any)?.message || error)}</p>
          <button 
            onClick={() => window.location.reload()} 
            className="mt-2 text-xs bg-red-100 text-red-700 px-2 py-1 rounded hover:bg-red-200"
          >
            Retry
          </button>
        </div>
      )}

      {!isLoading && !error && (!data || !data.items || data.items.length === 0) && (
        <div className="bg-white border border-gold/30 rounded-lg p-8 shadow-card text-center">
          <div className="text-dark/60">
            <svg className="w-16 h-16 mx-auto mb-4 text-gold/40" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
            </svg>
            <p className="text-lg font-medium mb-2">No products yet</p>
            <p className="text-sm mb-4">Create your first product to get started</p>
            <Link
              to="/products/new"
              className="inline-flex items-center justify-center rounded-md bg-maroon text-cream text-sm font-medium px-3 py-2 hover:bg-maroon/90"
            >
              Create Product
            </Link>
          </div>
        </div>
      )}

      {data && data.items && data.items.length > 0 && (
        <div className="bg-white border border-gold/30 rounded-lg shadow-card overflow-x-auto">
          <table className="min-w-full text-sm">
            <thead>
              <tr className="bg-cream/60 text-dark">
                <th className="text-left px-4 py-2">ID</th>
                <th className="text-left px-4 py-2">Title</th>
                <th className="text-left px-4 py-2">Slug</th>
                <th className="text-left px-4 py-2">Published</th>
                <th className="text-left px-4 py-2">Actions</th>
              </tr>
            </thead>
            <tbody>
              {data.items.map((p) => (
                <tr key={p.uuid_id} className="border-t border-gold/20">
                  <td className="px-4 py-2">{p.uuid_id}</td>
                  <td className="px-4 py-2">{p.title}</td>
                  <td className="px-4 py-2">{p.slug}</td>
                  <td className="px-4 py-2">{p.published ? 'Yes' : 'No'}</td>
                  <td className="px-4 py-2">
                    <Link to={`/products/${p.uuid_id}`} className="text-maroon hover:underline">Edit</Link>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
};
