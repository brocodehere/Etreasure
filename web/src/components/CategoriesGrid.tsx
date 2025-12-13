import React, { useState, useEffect } from 'react';
import { fetchCategories, type Category } from '../lib/api';


interface CategoriesGridProps {
  initialCategories?: Category[];
}

const CategoriesGrid: React.FC<CategoriesGridProps> = ({ initialCategories = [] }) => {
  const [categories, setCategories] = useState<Category[]>(initialCategories);
  const [loading, setLoading] = useState(!initialCategories.length);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    // Only fetch if we don't have initial categories
    if (initialCategories.length > 0) {
      return;
    }

    const loadCategories = async () => {
      try {
        setLoading(true);
        const data = await fetchCategories();
        setCategories(data.items);
      } catch (err) {
        console.error('Failed to fetch categories:', err);
        setError('Failed to load categories');
      } finally {
        setLoading(false);
      }
    };

    loadCategories();
  }, [initialCategories]);

  if (loading) {
    return (
      <div className="flex items-center justify-center py-20">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-maroon"></div>
        <span className="ml-3 text-dark/70">Loading categories...</span>
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

  if (categories.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-20">
        <div className="text-dark/70 text-center">
          <svg className="w-16 h-16 mx-auto mb-4 text-gold/40" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
          </svg>
          <p className="text-lg font-medium mb-2">No categories available</p>
          <p className="text-sm">Check back later for new categories</p>
        </div>
      </div>
    );
  }

  // Get image URL for category
  const getImageUrl = (category: Category) => {
    // If category has a media image URL from the API, use it
    if (category.image_url) {
      return category.image_url;
    }
    
    // Fallback to static images
    return getFallbackImage(category.name);
  };

  // Fallback image if no image is provided
  const getFallbackImage = (categoryName: string) => {
    const imageMap: { [key: string]: string } = {
      'hand bags': '/images/categories/hand.webp',
      'laptop bags': '/images/categories/enhanced_laptop__bag.png',
      'tote bags': '/images/categories/tote.webp',
      'crochet throws': '/images/categories/enhanced_cohort.png',
      'embroidered clutches': '/images/categories/clutches.png',
      'baby throw': '/images/categories/baby.webp',
      'everyday bags': '/images/categories/daily.webp',
    };
    
    const normalizedName = categoryName.toLowerCase();
    return imageMap[normalizedName] || '/images/categories/daily.webp';
  };

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-8">
      {categories.map((category, index) => (
        <a 
          href={`/shop?category=${category.slug}`}
          key={category.uuid_id || category.id || category.slug}
          className="group relative overflow-hidden rounded-2xl shadow-xl hover:shadow-2xl transition-all duration-700 transform hover:scale-105 animate-fade-up"
          style={{ animationDelay: `${index * 0.1}s` }}
        >
          {/* Category Card Container */}
          <div className="relative h-72 overflow-hidden bg-white">
            {/* Image with Enhanced Effects */}
            <div className="relative w-full h-full">
              <picture>
                <source srcSet={getImageUrl(category)} type="image/webp" />
                <source srcSet={getImageUrl(category).replace('.webp', '.png')} type="image/png" />
                <img 
                  src={getImageUrl(category)}
                  alt={category.name}
                  className="w-full h-full object-cover transition-all duration-700 group-hover:scale-110 group-hover:brightness-110"
                  loading="lazy"
                />
              </picture>
              
              {/* Overlay Gradient */}
              <div className="absolute inset-0 bg-gradient-to-t from-black/60 via-transparent to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-500"></div>
              
              {/* Category Name Badge */}
              <div className="absolute top-4 left-4 bg-gradient-to-r from-gold to-gold/80 text-maroon px-4 py-2 rounded-full font-semibold text-sm md:text-base shadow-lg backdrop-blur-sm transform group-hover:scale-110 transition-transform duration-300">
                {category.name}
              </div>
              
                            
              {/* Decorative Corner Accent */}
              <div className="absolute top-2 right-2 w-8 h-8 border-t-2 border-r-2 border-gold/50 transform rotate-45 group-hover:scale-150 transition-transform duration-500"></div>
            </div>
            
            {/* Card Border Effect */}
            <div className="absolute inset-0 rounded-2xl border-2 border-transparent group-hover:border-gold/50 transition-all duration-500 pointer-events-none"></div>
          </div>
        </a>
      ))}
    </div>
  );
};

export default CategoriesGrid;
