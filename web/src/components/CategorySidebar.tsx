import React, { useState, useEffect } from 'react';
import type { Category } from '../lib/api';
import { fetchCategories } from '../lib/api';

interface CategorySidebarProps {
  initialCategories?: Category[];
  initialCategory?: string;
}

const CategorySidebar: React.FC<CategorySidebarProps> = ({ initialCategories = [], initialCategory = '' }) => {
  const [categories, setCategories] = useState<Category[]>(initialCategories);
  const [selectedCategory, setSelectedCategory] = useState<string>(initialCategory);
  const [isLoading, setIsLoading] = useState(false);
  const [categorySlugToId, setCategorySlugToId] = useState<Record<string, string>>({});

  useEffect(() => {
    // Create slug-to-ID mapping from categories (both initial and fetched)
    const allCategories = [...initialCategories, ...categories];
    const slugToIdMap: Record<string, string> = {};
    
    allCategories.forEach((cat, index) => {
      if (cat.slug && (cat.id || cat.uuid_id)) {
        const categoryId = cat.id || cat.uuid_id || '';
        slugToIdMap[cat.slug] = categoryId;
      }
    });
    
    setCategorySlugToId(slugToIdMap);
  }, [categories, initialCategories]);

  useEffect(() => {
    if (initialCategories.length === 0) {
      setIsLoading(true);
      fetchCategories()
        .then(data => {
          setCategories(data.items);
        })
        .catch(() => {})
        .finally(() => setIsLoading(false));
    }
  }, [initialCategories]);

  // Handle initial category from URL on mount
  useEffect(() => {
    const urlParams = new URLSearchParams(window.location.search);
    const urlCategory = urlParams.get('category') || '';
    
    if (urlCategory && urlCategory !== selectedCategory) {
      setSelectedCategory(urlCategory);
    }
  }, [selectedCategory, initialCategory]);

  const handleCategorySelect = (slug: string) => {
    const newCategory = slug === selectedCategory ? '' : slug;
    
    // Get category ID from pre-built mapping
    const categoryId = categorySlugToId[newCategory] || '';
    
    setSelectedCategory(newCategory);
    
    // Update URL without page reload
    const url = newCategory ? `/shop?category=${newCategory}` : '/shop';
    window.history.pushState({}, '', url);
    
    // Emit custom event with category ID for API filtering
    window.dispatchEvent(new CustomEvent('categoryChange', { 
      detail: { category: newCategory, categoryId } 
    }));
  };

  if (isLoading) {
    return (
      <div className="bg-white rounded-2xl shadow-lg border border-gray-100 p-6">
        <div className="flex items-center mb-6">
          <div className="w-8 h-8 bg-gradient-to-r from-maroon to-burgundy rounded-lg mr-3"></div>
          <h2 className="text-xl font-playfair font-bold text-gray-800">Categories</h2>
        </div>
        <div className="space-y-3">
          {[...Array(5)].map((_, i) => (
            <div key={i} className="h-12 bg-gradient-to-r from-gray-100 to-gray-200 rounded-xl animate-pulse"></div>
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-2xl shadow-lg border border-gray-100 p-6">
      <div className="flex items-center mb-6">
        <div className="w-8 h-8 bg-gradient-to-r from-maroon to-burgundy rounded-lg mr-3"></div>
        <h2 className="text-xl font-playfair font-bold text-gray-800">Categories</h2>
      </div>
      
      {/* Mobile Category Select */}
      <div className="lg:hidden mb-4">
        <select 
          className="w-full p-3 border-2 border-gray-200 rounded-xl bg-white text-gray-700 focus:border-maroon focus:outline-none transition-colors"
          value={selectedCategory}
          onChange={(e) => handleCategorySelect(e.target.value)}
        >
          <option value="">All Categories</option>
          {categories.map(category => (
            <option key={category.id || category.slug} value={category.slug}>
              {category.name}
            </option>
          ))}
        </select>
      </div>

      {/* Desktop Category Pills */}
      <nav className="hidden lg:block space-y-2">
        <button
          onClick={() => handleCategorySelect('')}
          className={`w-full text-left px-4 py-3 rounded-xl transition-all duration-200 transform hover:scale-[1.02] ${
            selectedCategory === '' 
              ? 'bg-gradient-to-r from-maroon to-burgundy text-white shadow-md' 
              : 'bg-gray-50 hover:bg-gray-100 text-gray-700 border border-gray-200'
          }`}
        >
          <div className="flex items-center justify-between">
            <span className="font-medium">All Categories</span>
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M4 6h16M4 12h16M4 18h16"></path>
            </svg>
          </div>
        </button>
        
        {categories.map(category => (
          <button
            key={category.id || category.slug}
            onClick={() => handleCategorySelect(category.slug)}
            className={`w-full text-left px-4 py-3 rounded-xl transition-all duration-200 transform hover:scale-[1.02] ${
              selectedCategory === category.slug 
                ? 'bg-gradient-to-r from-maroon to-burgundy text-white shadow-md' 
                : 'bg-gray-50 hover:bg-gray-100 text-gray-700 border border-gray-200'
            }`}
          >
            <div className="flex items-center justify-between">
              <span className="font-medium">{category.name}</span>
              {category.product_count && (
                <span className={`text-sm px-2 py-1 rounded-full ${
                  selectedCategory === category.slug 
                    ? 'bg-white/20' 
                    : 'bg-gray-200 text-gray-600'
                }`}>
                  {category.product_count}
                </span>
              )}
            </div>
          </button>
        ))}
      </nav>
    </div>
  );
};

export default CategorySidebar;
