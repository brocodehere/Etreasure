import React, { useState, useEffect, useCallback } from 'react';
import type { Product } from '../lib/api';
import { fetchProducts } from '../lib/api';
import ProductCard from './ProductCard';
import Pagination from './Pagination';

interface ProductGridProps {
  initialProducts?: Product[];
  initialTotal?: number;
}

const ProductGrid: React.FC<ProductGridProps> = ({ 
  initialProducts = [], 
  initialTotal = 0 
}) => {
  const [products, setProducts] = useState<Product[]>(initialProducts);
  const [total, setTotal] = useState(initialTotal);
  const [currentPage, setCurrentPage] = useState(1);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  
  // Filter state
  const [filters, setFilters] = useState({
    category: '',
    categoryId: '',
    search: '',
    sortBy: 'newest' as const,
    priceRange: { min: '', max: '' },
    inStockOnly: false
  });

  const syncUrlWithFilters = useCallback((nextFilters: typeof filters, page: number = 1) => {
    const url = new URL(window.location.href);
    if (nextFilters.category) url.searchParams.set('category', nextFilters.category);
    else url.searchParams.delete('category');
    if (nextFilters.search) url.searchParams.set('search', nextFilters.search);
    else url.searchParams.delete('search');
    if (nextFilters.sortBy && nextFilters.sortBy !== 'newest') url.searchParams.set('sort', nextFilters.sortBy);
    else url.searchParams.delete('sort');
    if (nextFilters.priceRange.min) url.searchParams.set('min_price', nextFilters.priceRange.min);
    else url.searchParams.delete('min_price');
    if (nextFilters.priceRange.max) url.searchParams.set('max_price', nextFilters.priceRange.max);
    else url.searchParams.delete('max_price');
    if (nextFilters.inStockOnly) url.searchParams.set('in_stock', '1');
    else url.searchParams.delete('in_stock');
    url.searchParams.set('page', String(page));
    window.history.pushState({}, '', url.toString());
  }, []);

  // Load products with filters
  const loadProducts = useCallback(async (page: number = 1) => {
    setIsLoading(true);
    setError(null);
    
    try {
      const params: any = {
        page,
        limit: 12
      };
      
      if (filters.category) params.category = filters.category;
      if (filters.search) params.search = filters.search;
      if (filters.sortBy !== 'newest') params.sort = filters.sortBy;
      if (filters.priceRange.min) params.min_price = filters.priceRange.min;
      if (filters.priceRange.max) params.max_price = filters.priceRange.max;
      if (filters.inStockOnly) params.in_stock = '1';
      const data = await fetchProducts(params);
      setProducts(data.items);
      setTotal(data.total);
      setCurrentPage(data.page);
    } catch (err) {
      setError('Failed to load products. Please try again.');
    } finally {
      setIsLoading(false);
    }
  }, [filters]);

  // Parse URL parameters on mount and handle navigation
  useEffect(() => {
    const parseUrlParams = () => {
      const params = new URLSearchParams(window.location.search);
      const category = params.get('category') || '';
      const search = params.get('search') || '';
      const sort = params.get('sort') as any || 'newest';
      const minPrice = params.get('min_price') || '';
      const maxPrice = params.get('max_price') || '';
      const inStockOnly = params.get('in_stock') === '1';
      
      return {
        category,
        search,
        sortBy: sort,
        priceRange: { min: minPrice, max: maxPrice },
        inStockOnly
      };
    };

    // Set initial filters from URL
    const urlFilters = parseUrlParams();
        
    // Get initial data from window if available (from SSR)
    const initialData = (window as any).__INITIAL_DATA__;
        
    // If we have initial category from SSR, use it
    if (initialData?.initialCategory) {
      setFilters({ 
        ...urlFilters, 
        category: initialData.initialCategory,
        categoryId: '' // Will be set by categoryChange event
      });
    } else {
      setFilters({ ...urlFilters, categoryId: '' });
    }

    // Listen for filter changes
    const handleCategoryChange = (e: CustomEvent) => {
      setFilters(prev => {
        const next = {
          ...prev,
          category: e.detail.category,
          categoryId: e.detail.categoryId || ''
        };
        syncUrlWithFilters(next, 1);
        return next;
      });
    };
    
    const handleSearchChange = (e: CustomEvent) => {
      setFilters(prev => {
        const next = {
          ...prev,
          search: e.detail.search,
          sortBy: e.detail.sortBy,
          priceRange: e.detail.priceRange,
          inStockOnly: !!e.detail.inStockOnly,
          categoryId: prev.categoryId
        };
        syncUrlWithFilters(next, 1);
        return next;
      });
    };
    
    // Handle browser back/forward
    const handlePopState = () => {
      const urlFilters = parseUrlParams();
      setFilters({ ...urlFilters, categoryId: '' });
    };
    
        window.addEventListener('categoryChange', handleCategoryChange as EventListener);
    window.addEventListener('searchChange', handleSearchChange as EventListener);
    window.addEventListener('popstate', handlePopState);
    
    return () => {
            window.removeEventListener('categoryChange', handleCategoryChange as EventListener);
      window.removeEventListener('searchChange', handleSearchChange as EventListener);
      window.removeEventListener('popstate', handlePopState);
    };
  }, []);

  // Reload products when filters change
  useEffect(() => {
    loadProducts(1);
  }, [loadProducts]);

  // Handle pagination with URL updates
  const handlePageChange = (page: number) => {
    // Update URL with page parameter
    const url = new URL(window.location.href);
    url.searchParams.set('page', page.toString());
    window.history.pushState({}, '', url.toString());
    
    loadProducts(page);
  };

  // Loading skeleton
  if (isLoading && products.length === 0) {
    return (
      <div>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
          {[...Array(12)].map((_, i) => (
            <div key={i} className="bg-white rounded-2xl shadow-lg overflow-hidden border border-gray-100">
              <div className="h-64 bg-gradient-to-br from-gray-100 to-gray-200 animate-pulse"></div>
              <div className="p-6">
                <div className="h-6 bg-gradient-to-r from-gray-200 to-gray-300 rounded-lg mb-3 animate-pulse"></div>
                <div className="h-4 bg-gradient-to-r from-gray-200 to-gray-300 rounded mb-4 animate-pulse"></div>
                <div className="flex items-center justify-between">
                  <div className="h-8 w-24 bg-gradient-to-r from-gray-200 to-gray-300 rounded-lg animate-pulse"></div>
                  <div className="h-10 w-10 bg-gradient-to-r from-gray-200 to-gray-300 rounded-full animate-pulse"></div>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    );
  }

  // Error state
  if (error && products.length === 0) {
    return (
      <div className="text-center py-16">
        <div className="inline-flex items-center justify-center w-16 h-16 bg-red-100 rounded-full mb-6">
          <svg className="w-8 h-8 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.932-3L13.932 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.932 3z"></path>
          </svg>
        </div>
        <h3 className="text-2xl font-playfair font-bold text-gray-800 mb-4">Oops! Something went wrong</h3>
        <p className="text-gray-600 mb-8 max-w-md mx-auto">{error}</p>
        <button
          onClick={() => loadProducts(1)}
          className="bg-gradient-to-r from-maroon to-burgundy text-white px-8 py-3 rounded-full hover:shadow-lg transform hover:scale-105 transition-all duration-300"
        >
          Try Again
        </button>
      </div>
    );
  }

  // Empty state
  if (products.length === 0 && !isLoading) {
    return (
      <div className="text-center py-16">
        <div className="inline-flex items-center justify-center w-20 h-20 bg-gray-100 rounded-full mb-6">
          <svg className="w-10 h-10 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
          </svg>
        </div>
        <h3 className="text-2xl font-playfair font-bold text-gray-800 mb-4">No products found</h3>
        <p className="text-gray-600 mb-8 max-w-md mx-auto">Try adjusting your filters or browse all categories to discover more items</p>
        <a
          href="/shop"
          className="inline-flex items-center bg-gradient-to-r from-maroon to-burgundy text-white px-8 py-3 rounded-full hover:shadow-lg transform hover:scale-105 transition-all duration-300"
        >
          <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M4 6h16M4 12h16M4 18h16"></path>
          </svg>
          Browse All Products
        </a>
      </div>
    );
  }

  return (
    <div>
      {/* Results Header */}
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center mb-8 pb-6 border-b border-gray-200">
        <div>
          <h2 className="text-2xl font-playfair font-bold text-gray-800 mb-2">
            {total} Products Found
          </h2>
          <p className="text-gray-600">Discover our handcrafted collection</p>
        </div>
        {isLoading && (
          <div className="flex items-center text-maroon mt-4 sm:mt-0">
            <svg className="animate-spin h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24">
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            Loading...
          </div>
        )}
      </div>

      {/* Product Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
        {products.map((product) => (
          <ProductCard key={product.id} product={product} />
        ))}
      </div>

      {/* Pagination */}
      {total > 12 && (
        <div className="mt-8">
          <Pagination
            currentPage={currentPage}
            totalItems={total}
            itemsPerPage={12}
            onPageChange={handlePageChange}
            isLoading={isLoading}
          />
        </div>
      )}

      {/* Loading overlay for pagination */}
      {isLoading && products.length > 0 && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white p-6 rounded-lg">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-maroon"></div>
            <p className="mt-2 text-gray-600">Loading...</p>
          </div>
        </div>
      )}
    </div>
  );
};

export default ProductGrid;
