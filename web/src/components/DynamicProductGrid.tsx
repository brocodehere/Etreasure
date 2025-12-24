import React, { useState, useEffect } from 'react';
import { fetchCategories, fetchProducts, API_BASE_URL } from '../lib/api';

interface Product {
  id: string;
  slug: string;
  title: string;
  description?: string;
  category_id?: string;
  price_cents?: number;
  currency?: string;
  image_key?: string;
  image_url?: string;
}

const DynamicProductGrid: React.FC = () => {
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const loadProducts = async () => {
      try {
        setLoading(true);
        const data = await fetchProducts({ page: 1, limit: 20 });
        setProducts(data.items || []);
      } catch (err) {
      } finally {
        setLoading(false);
      }
    };

    loadProducts();
  }, []);

  if (loading) {
    return (
      <div className="flex items-center justify-center py-20">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-maroon"></div>
        <span className="ml-3 text-dark/70">Loading products...</span>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center py-20">
        <div className="text-red-600 mb-4">{error}</div>
        <button 
          onClick={() => window.location.reload()}
          className="px-4 py-2 bg-maroon text-white rounded hover:bg-maroon/90"
        >
          Retry
        </button>
      </div>
    );
  }

  if (products.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-20">
        <div className="text-dark/70 text-center">
          <svg className="w-16 h-16 mx-auto mb-4 text-gold/40" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
          </svg>
          <p className="text-lg font-medium mb-2">No products available</p>
          <p className="text-sm">Check back later for new products</p>
        </div>
      </div>
    );
  }

  // Format price
  const formatPrice = (cents?: number, currency?: string) => {
    if (!cents) return 'Price not available';
    const price = cents / 100;
    return new Intl.NumberFormat('en-IN', {
      style: 'currency',
      currency: currency || 'INR'
    }).format(price);
  };

  // Get image URL for product
  const getImageUrl = (product: Product) => {
    // If product has a media image URL from the API, use it
    if (product.image_url) {
      return product.image_url;
    }
    
    // Fallback to placeholder
    return 'https://pub-1a3924a6c6994107be6fe9f3ed794c0a.r2.dev/product-placeholder.webp';
  };

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
      {products.map((product) => (
        <div key={product.id} className="bg-white rounded-lg shadow-md overflow-hidden group hover:shadow-lg transition-shadow duration-300">
          <div className="relative overflow-hidden">
            <img
              src={getImageUrl(product)}
              alt={product.title}
              className="w-full h-64 object-cover group-hover:scale-105 transition-transform duration-300"
              onError={(e) => {
                // Fallback to placeholder if image fails to load
                e.currentTarget.src = 'https://pub-1a3924a6c6994107be6fe9f3ed794c0a.r2.dev/product-placeholder.webp';
              }}
            />
            <div className="absolute top-2 right-2 bg-gold text-white px-2 py-1 rounded text-xs font-medium">
              New
            </div>
          </div>
          <div className="p-4">
            <h3 className="font-semibold text-lg text-dark mb-2 line-clamp-1">{product.title}</h3>
            {product.description && (
              <p className="text-sm text-dark/60 mb-3 line-clamp-2">{product.description}</p>
            )}
            <div className="flex items-center justify-between">
              <div className="text-lg font-bold text-maroon">
                {formatPrice(product.price_cents, product.currency)}
              </div>
              <button
                onClick={() => window.location.href = `/product/${product.slug}`}
                className="px-3 py-1 bg-maroon text-white text-sm rounded hover:bg-maroon/90 transition-colors"
              >
                View Details
              </button>
            </div>
          </div>
        </div>
      ))}
    </div>
  );
};

export default DynamicProductGrid;
