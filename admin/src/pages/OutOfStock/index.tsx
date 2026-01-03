import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../../lib/api';
import { LoadingState, LoadingSpinner } from '../../components/LoadingSpinner';

interface OutOfStockProduct {
  id: string;
  title: string;
  slug: string;
  currency: string;
  image_key?: string;
  image_url?: string;
  category_id?: string;
  total_stock: number;
  variants: ProductVariant[];
}

interface ProductVariant {
  id: number;
  product_id: string;
  sku: string;
  title: string;
  price_cents: number;
  stock_quantity: number;
  currency: string;
}

export const OutOfStockPage: React.FC = () => {
  const [products, setProducts] = useState<OutOfStockProduct[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [updatingStock, setUpdatingStock] = useState<string | null>(null);
  const [stockInputs, setStockInputs] = useState<Record<string, string>>({});

  useEffect(() => {
    fetchOutOfStockProducts();
  }, []);

  const fetchOutOfStockProducts = async () => {
    try {
      setLoading(true);
      const response = await api.get<{ products: OutOfStockProduct[] }>('/products/out-of-stock');
      const productsData = response.products || [];
      setProducts(productsData);
      
      // Initialize stock inputs with current values
      const initialInputs: Record<string, string> = {};
      productsData.forEach(product => {
        product.variants.forEach(variant => {
          initialInputs[`${product.id}-${variant.id}`] = variant.stock_quantity.toString();
        });
      });
      setStockInputs(initialInputs);
    } catch (err: any) {
      setError(err.message || 'Failed to fetch out of stock products');
    } finally {
      setLoading(false);
    }
  };

  const updateStock = async (productId: string, variantId: number, newStock: number) => {
    try {
      setUpdatingStock(`${productId}-${variantId}`);
      await api.patch(`/products/${productId}/variants/${variantId}/stock`, {
        stock_quantity: newStock
      });
      
      // Refresh the products list
      await fetchOutOfStockProducts();
    } catch (err: any) {
      alert(err.message || 'Failed to update stock');
    } finally {
      setUpdatingStock(null);
    }
  };

  const handleStockInputChange = (productId: string, variantId: number, value: string) => {
    const key = `${productId}-${variantId}`;
    setStockInputs(prev => ({
      ...prev,
      [key]: value
    }));
  };

  const handleStockChange = (productId: string, variantId: number, value: string) => {
    const stockValue = parseInt(value);
    if (!isNaN(stockValue) && stockValue >= 0) {
      updateStock(productId, variantId, stockValue);
    }
  };

  const handleUpdateClick = (productId: string, variantId: number) => {
    const key = `${productId}-${variantId}`;
    const value = stockInputs[key];
    if (value !== undefined) {
      handleStockChange(productId, variantId, value);
    }
  };

  if (loading) {
    return <LoadingSpinner />;
  }

  if (error) {
    return (
      <div className="p-6">
        <div className="bg-red-50 border border-red-200 rounded-lg p-4">
          <h3 className="text-red-800 font-medium">Error</h3>
          <p className="text-red-600 mt-1">{error}</p>
          <button
            onClick={fetchOutOfStockProducts}
            className="mt-3 bg-red-600 text-white px-4 py-2 rounded hover:bg-red-700"
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <header>
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Out of Stock Products</h1>
            <p className="text-gray-600 mt-1">
              Manage inventory for products that are currently out of stock
            </p>
          </div>
          <Link
            to="/products"
            className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors"
          >
            View All Products
          </Link>
        </div>
      </header>

      {products.length === 0 ? (
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-12 text-center">
          <div className="inline-flex items-center justify-center w-16 h-16 bg-green-100 rounded-full mb-4">
            <svg className="w-8 h-8 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <h3 className="text-xl font-semibold text-gray-900 mb-2">All products are in stock!</h3>
          <p className="text-gray-600">Great job managing your inventory.</p>
        </div>
      ) : (
        <div className="bg-white rounded-lg shadow-sm border border-gray-200">
          <div className="px-6 py-4 border-b border-gray-200">
            <h2 className="text-lg font-semibold text-gray-900">
              {products.length} {products.length === 1 ? 'Product' : 'Products'} Out of Stock
            </h2>
          </div>
          
          <div className="divide-y divide-gray-200">
            {products.map((product) => (
              <div key={product.id} className="p-6">
                <div className="flex items-start space-x-4">
                  {/* Product Image */}
                  <div className="flex-shrink-0">
                    {product.image_url ? (
                      <img
                        src={product.image_url}
                        alt={product.title}
                        className="w-20 h-20 object-cover rounded-lg border border-gray-200"
                      />
                    ) : (
                      <div className="w-20 h-20 bg-gray-100 rounded-lg border border-gray-200 flex items-center justify-center">
                        <svg className="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                        </svg>
                      </div>
                    )}
                  </div>

                  {/* Product Info */}
                  <div className="flex-1">
                    <div className="flex items-start justify-between">
                      <div>
                        <h3 className="text-lg font-medium text-gray-900">
                          <Link
                            to={`/products/${product.id}`}
                            className="hover:text-blue-600 transition-colors"
                          >
                            {product.title}
                          </Link>
                        </h3>
                        <p className="text-sm text-gray-600 mt-1">SKU: {product.variants[0]?.sku}</p>
                        <p className="text-lg font-semibold text-gray-900 mt-2">
                          {product.variants.length > 0 ? (() => {
                            const prices = product.variants.map(v => v.price_cents / 100);
                            const minPrice = Math.min(...prices);
                            const maxPrice = Math.max(...prices);
                            
                            if (minPrice === maxPrice) {
                              return `₹${minPrice.toLocaleString('en-IN')}`;
                            } else {
                              return `₹${minPrice.toLocaleString('en-IN')} - ₹${maxPrice.toLocaleString('en-IN')}`;
                            }
                          })() : 'N/A'}
                        </p>
                      </div>
                      
                      <div className="text-right">
                        <span className="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-red-100 text-red-800">
                          Out of Stock
                        </span>
                      </div>
                    </div>

                    {/* Variants with Stock Management */}
                    <div className="mt-4 space-y-3">
                      {product.variants.map((variant) => (
                        <div key={variant.id} className="flex items-center justify-between bg-gray-50 rounded-lg p-3">
                          <div className="flex-1">
                            <p className="text-sm font-medium text-gray-900">{variant.title}</p>
                            <p className="text-sm text-gray-600">Current Stock: {variant.stock_quantity}</p>
                          </div>
                          
                          <div className="flex items-center space-x-2">
                            <input
                              type="number"
                              min="0"
                              value={stockInputs[`${product.id}-${variant.id}`] ?? variant.stock_quantity.toString()}
                              onChange={(e) => handleStockInputChange(product.id, variant.id, e.target.value)}
                              className="w-20 px-2 py-1 border border-gray-300 rounded text-sm focus:ring-blue-500 focus:border-blue-500"
                              disabled={updatingStock === `${product.id}-${variant.id}`}
                              onKeyDown={(e) => {
                                if (e.key === 'Enter') {
                                  handleStockChange(product.id, variant.id, (e.target as HTMLInputElement).value);
                                }
                              }}
                            />
                            <button
                              onClick={() => handleUpdateClick(product.id, variant.id)}
                              disabled={updatingStock === `${product.id}-${variant.id}`}
                              className="px-3 py-1 bg-blue-600 text-white text-sm rounded hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
                            >
                              {updatingStock === `${product.id}-${variant.id}` ? (
                                <svg className="animate-spin h-4 w-4" fill="none" viewBox="0 0 24 24">
                                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                                </svg>
                              ) : (
                                'Update'
                              )}
                            </button>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};
