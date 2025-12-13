import React, { useState } from 'react';

interface Facet {
  id: number;
  name: string;
  productCount: number;
}

interface PriceRange {
  min: number;
  max: number;
  avg: number;
}

interface SearchFiltersProps {
  categories: Facet[];
  priceRange: PriceRange;
  currentCategory?: number;
  currentMinPrice?: number;
  currentMaxPrice?: number;
  onFilterChange: (filters: {
    category?: number;
    minPrice?: number;
    maxPrice?: number;
  }) => void;
}

export default function SearchFilters({
  categories,
  priceRange,
  currentCategory,
  currentMinPrice,
  currentMaxPrice,
  onFilterChange,
}: SearchFiltersProps) {
  const [expandedCategory, setExpandedCategory] = useState(true);
  const [expandedPrice, setExpandedPrice] = useState(true);
  const [minPrice, setMinPrice] = useState(currentMinPrice || priceRange.min);
  const [maxPrice, setMaxPrice] = useState(currentMaxPrice || priceRange.max);

  const handleCategoryChange = (categoryId: number) => {
    onFilterChange({
      category: currentCategory === categoryId ? undefined : categoryId,
      minPrice,
      maxPrice,
    });
  };

  const handlePriceChange = () => {
    onFilterChange({
      category: currentCategory,
      minPrice,
      maxPrice,
    });
  };

  const handleReset = () => {
    setMinPrice(priceRange.min);
    setMaxPrice(priceRange.max);
    onFilterChange({});
  };

  return (
    <div className="bg-white rounded-lg border border-gold/20 p-6 h-fit">
      <div className="flex justify-between items-center mb-6">
        <h3 className="font-semibold text-lg text-dark">Filters</h3>
        <button
          onClick={handleReset}
          className="text-sm text-gold hover:text-gold/80 transition-colors"
        >
          Reset
        </button>
      </div>

      {/* Categories */}
      {categories.length > 0 && (
        <div className="mb-6 pb-6 border-b border-gold/10">
          <button
            onClick={() => setExpandedCategory(!expandedCategory)}
            className="flex items-center justify-between w-full font-medium text-dark mb-3 hover:text-gold transition-colors"
          >
            <span>Categories</span>
            <svg
              className={`w-4 h-4 transition-transform ${expandedCategory ? 'rotate-180' : ''}`}
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M19 14l-7 7m0 0l-7-7m7 7V3"
              />
            </svg>
          </button>

          {expandedCategory && (
            <div className="space-y-2">
              {categories.map((category) => (
                <label key={category.id} className="flex items-center gap-2 cursor-pointer group">
                  <input
                    type="checkbox"
                    checked={currentCategory === category.id}
                    onChange={() => handleCategoryChange(category.id)}
                    className="w-4 h-4 rounded border-gold/30 text-gold focus:ring-gold"
                  />
                  <span className="text-sm text-dark/80 group-hover:text-dark transition-colors flex-1">
                    {category.name}
                  </span>
                  <span className="text-xs text-dark/40">({category.productCount})</span>
                </label>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Price Range */}
      <div className="mb-6">
        <button
          onClick={() => setExpandedPrice(!expandedPrice)}
          className="flex items-center justify-between w-full font-medium text-dark mb-3 hover:text-gold transition-colors"
        >
          <span>Price Range</span>
          <svg
            className={`w-4 h-4 transition-transform ${expandedPrice ? 'rotate-180' : ''}`}
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M19 14l-7 7m0 0l-7-7m7 7V3"
            />
          </svg>
        </button>

        {expandedPrice && (
          <div className="space-y-4">
            <div>
              <label className="block text-xs text-dark/60 mb-1">Min Price: ₹{minPrice}</label>
              <input
                type="range"
                min={priceRange.min}
                max={priceRange.max}
                value={minPrice}
                onChange={(e) => setMinPrice(parseInt(e.target.value))}
                onMouseUp={handlePriceChange}
                onTouchEnd={handlePriceChange}
                className="w-full h-2 bg-gold/20 rounded-lg appearance-none cursor-pointer accent-gold"
              />
            </div>

            <div>
              <label className="block text-xs text-dark/60 mb-1">Max Price: ₹{maxPrice}</label>
              <input
                type="range"
                min={priceRange.min}
                max={priceRange.max}
                value={maxPrice}
                onChange={(e) => setMaxPrice(parseInt(e.target.value))}
                onMouseUp={handlePriceChange}
                onTouchEnd={handlePriceChange}
                className="w-full h-2 bg-gold/20 rounded-lg appearance-none cursor-pointer accent-gold"
              />
            </div>

            <div className="text-sm text-dark/60 bg-maroon/5 rounded p-2">
              ₹{minPrice} - ₹{maxPrice}
            </div>
          </div>
        )}
      </div>

      {/* Average Price Info */}
      {priceRange.avg > 0 && (
        <div className="text-xs text-dark/50 bg-gold/5 rounded p-2">
          Average Price: ₹{priceRange.avg}
        </div>
      )}
    </div>
  );
}
