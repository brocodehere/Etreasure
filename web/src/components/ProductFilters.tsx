import React, { useState, useEffect } from 'react';

interface ProductFiltersProps {}

const ProductFilters: React.FC<ProductFiltersProps> = () => {
  const [searchTerm, setSearchTerm] = useState('');
  const [sortBy, setSortBy] = useState<'newest' | 'price_asc' | 'price_desc' | 'name_asc' | 'name_desc'>('newest');
  const [priceRange, setPriceRange] = useState({ min: '', max: '' });
  const [inStockOnly, setInStockOnly] = useState(false);

  // Local debounce function for SSR compatibility
  const debounce = <T extends (...args: any[]) => any>(
    func: T,
    wait: number
  ): ((...args: Parameters<T>) => void) => {
    let timeout: ReturnType<typeof setTimeout>;
    return (...args: Parameters<T>) => {
      clearTimeout(timeout);
      timeout = setTimeout(() => func(...args), wait);
    };
  };

  // Debounced dispatch for search typing
  const debouncedDispatch = debounce((term: string) => {
    window.dispatchEvent(new CustomEvent('searchChange', {
      detail: { search: term, sortBy, priceRange, inStockOnly }
    }));
  }, 250);

  useEffect(() => {
    debouncedDispatch(searchTerm);
  }, [searchTerm, debouncedDispatch]);

  const handleSortChange = (value: string) => {
    setSortBy(value as any);
    window.dispatchEvent(new CustomEvent('searchChange', { 
      detail: { search: searchTerm, sortBy: value, priceRange, inStockOnly } 
    }));
  };

  const handlePriceChange = (type: 'min' | 'max', value: string) => {
    const newRange = { ...priceRange, [type]: value };
    setPriceRange(newRange);
    window.dispatchEvent(new CustomEvent('searchChange', { 
      detail: { search: searchTerm, sortBy, priceRange: newRange, inStockOnly } 
    }));
  };

  const handleStockToggle = (checked: boolean) => {
    setInStockOnly(checked);
    window.dispatchEvent(new CustomEvent('searchChange', {
      detail: { search: searchTerm, sortBy, priceRange, inStockOnly: checked }
    }));
  };

  const handleClearFilters = () => {
    setSearchTerm('');
    setSortBy('newest');
    setPriceRange({ min: '', max: '' });
    setInStockOnly(false);
    window.dispatchEvent(new CustomEvent('searchChange', { 
      detail: { search: '', sortBy: 'newest', priceRange: { min: '', max: '' }, inStockOnly: false }
    }));
  };

  return (
    <div className="bg-white rounded-xl shadow-lg border border-gray-100 p-6 mb-8">
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 items-end">
        {/* Search Input */}
        <div className="sm:col-span-2">
          <label htmlFor="search" className="block text-sm font-medium text-gray-700 mb-2">
            Search Products
          </label>
          <div className="relative">
            <input
              type="text"
              id="search"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              placeholder="Search for products..."
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-maroon focus:border-transparent"
            />
            {searchTerm && (
              <button
                onClick={() => setSearchTerm('')}
                className="absolute right-2 top-2.5 text-gray-400 hover:text-gray-600"
                aria-label="Clear search"
              >
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M6 18L18 6M6 6l6 6" />
                </svg>
              </button>
            )}
          </div>
        </div>

        {/* Sort Dropdown */}
        <div>
          <label htmlFor="sort" className="block text-sm font-medium text-gray-700 mb-2">
            Sort By
          </label>
          <select
            id="sort"
            value={sortBy}
            onChange={(e) => handleSortChange(e.target.value)}
            className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-maroon focus:border-transparent"
          >
            <option value="newest">Newest First</option>
            <option value="price_asc">Price: Low to High</option>
            <option value="price_desc">Price: High to Low</option>
            <option value="name_asc">Name: A to Z</option>
            <option value="name_desc">Name: Z to A</option>
          </select>
        </div>

        {/* Stock */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">Availability</label>
          <label className="flex items-center gap-2 px-4 py-2 border border-gray-300 rounded-lg">
            <input
              type="checkbox"
              checked={inStockOnly}
              onChange={(e) => handleStockToggle(e.target.checked)}
            />
            <span className="text-sm text-gray-700">In stock only</span>
          </label>
        </div>

        {/* Price Range */}
        <div className="sm:col-span-2">
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Price Range
          </label>
          <div className="flex gap-2">
            <input
              type="number"
              placeholder="Min"
              value={priceRange.min}
              onChange={(e) => handlePriceChange('min', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-maroon focus:border-transparent"
            />
            <span className="self-center text-gray-500">-</span>
            <input
              type="number"
              placeholder="Max"
              value={priceRange.max}
              onChange={(e) => handlePriceChange('max', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-maroon focus:border-transparent"
            />
          </div>
        </div>

        {/* Clear Filters Button */}
        <div className="sm:col-span-2 flex justify-end">
          <button
            onClick={handleClearFilters}
            className="px-4 py-2 bg-gray-100 hover:bg-gray-200 text-gray-700 rounded-lg transition-colors duration-200 font-medium"
          >
            Clear
          </button>
        </div>
      </div>
    </div>
  );
};

export default ProductFilters;
