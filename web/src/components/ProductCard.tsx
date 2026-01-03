import React, { useState } from 'react';
import type { Product } from '../lib/api';
import { addToCart, toggleWishlist } from '../lib/api';
import { showSuccess, showError } from '../lib/toast.js';
import StockNotificationModal from './StockNotificationModal';

interface ProductCardProps {
  product: Product;
}

const ProductCard: React.FC<ProductCardProps> = ({ product }) => {
  const [isWishlisted, setIsWishlisted] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [showNotificationModal, setShowNotificationModal] = useState(false);

  // Check if product is in stock
  const isInStock = product.stock_quantity >= 1;
  
  // Check if product is new arrival (created within last 5 days)
  const isNewArrival = () => {
    if (!product.created_at) return false;
    const createdDate = new Date(product.created_at);
    const currentDate = new Date();
    const diffTime = Math.abs(currentDate.getTime() - createdDate.getTime());
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
    return diffDays <= 5;
  };

  const handleAddToCart = async () => {
    setIsLoading(true);
    try {
      await addToCart(product.id);
      showSuccess('Product added to cart successfully!');
    } catch (error) {
      showError('Failed to add product to cart. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  const handleToggleWishlist = async () => {
    try {
      await toggleWishlist(product.id);
      setIsWishlisted(!isWishlisted);
      // Show success notification
      showSuccess(isWishlisted ? 'Product removed from wishlist!' : 'Product added to wishlist!');
    } catch (error) {
      showError('Failed to update wishlist. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="bg-white rounded-2xl shadow-lg overflow-hidden group hover:shadow-2xl transition-all duration-500 transform hover:-translate-y-1 border border-gray-100">
      {/* Product Image */}
      <div className="relative h-72 overflow-hidden bg-gray-50">
        <a 
          href={`/product/${product.slug || product.id}`}
          className="block w-full h-full relative z-10"
        >
          <img
            src={product.image_url || '/images/placeholder-product.jpg'}
            alt={product.title}
            className="w-full h-full object-cover group-hover:scale-110 transition-transform duration-700"
            loading="lazy"
          />
        </a>
        
        {/* Overlay on hover */}
        <div className="absolute inset-0 bg-gradient-to-t from-black/20 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300 pointer-events-none"></div>
        
        {/* Wishlist Button */}
        <button
          onClick={handleToggleWishlist}
          disabled={isLoading}
          className="absolute top-4 right-4 bg-white/95 backdrop-blur-sm p-3 rounded-full hover:bg-white hover:scale-110 transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed shadow-md z-20"
        >
          <svg
            className={`w-5 h-5 transition-colors duration-200 ${isWishlisted ? 'fill-red-500 text-red-500' : 'text-gray-600'}`}
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth="2" 
              d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z"
            />
          </svg>
        </button>

        {/* Quick View Badge - Only show if new arrival */}
        {isNewArrival() && (
          <div className="absolute top-4 left-4 z-20">
            <span className="bg-gradient-to-r from-maroon to-burgundy text-white text-xs px-3 py-1 rounded-full font-medium shadow-md">
              New Arrival
            </span>
          </div>
        )}
      </div>

      {/* Product Info */}
      <div className="p-6">
        <h3 className="font-playfair text-xl font-bold text-gray-800 mb-3 line-clamp-2">
          <a 
            href={`/product/${product.slug || product.id}`} 
            className="hover:text-maroon transition-colors duration-200 group-hover:text-maroon"
          >
            {product.title}
          </a>
        </h3>
        
        {product.short_description && (
          <p className="text-gray-600 text-sm mb-4 line-clamp-2 leading-relaxed">
            {product.short_description}
          </p>
        )}
        
        <div className="flex items-center justify-between mb-4">
          <div className="flex flex-col">
            <span className="text-2xl font-bold text-maroon">
              ₹{product.price_cents ? (product.price_cents / 100).toLocaleString('en-IN', { minimumFractionDigits: 2, maximumFractionDigits: 2 }) : '0'}
            </span>
            {product.original_price_cents && product.original_price_cents > product.price_cents && (
              <span className="text-sm text-gray-500 line-through">
                ₹{(product.original_price_cents / 100).toLocaleString('en-IN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
              </span>
            )}
          </div>
          
          <div className="flex items-center space-x-2">
            <span className={`text-xs font-medium ${isInStock ? 'text-green-600' : 'text-red-600'}`}>
              {isInStock ? 'In Stock' : 'Out of Stock'}
            </span>
            <div className={`w-2 h-2 rounded-full ${isInStock ? 'bg-green-500' : 'bg-red-500'}`}></div>
          </div>
        </div>
        
        <button
          onClick={isInStock ? handleAddToCart : () => setShowNotificationModal(true)}
          disabled={isLoading}
          className={`w-full px-4 py-3 rounded-xl font-semibold transition-all duration-300 disabled:cursor-not-allowed focus:outline-none focus:ring-2 focus:ring-maroon focus:ring-offset-2 transform hover:scale-[1.02] flex items-center justify-center space-x-2 ${
            isInStock 
              ? 'bg-gradient-to-r from-maroon to-burgundy text-white hover:from-gold hover:to-gold hover:text-maroon' 
              : 'bg-gradient-to-r from-blue-500 to-blue-600 text-white hover:from-blue-600 hover:to-blue-700'
          }`}
            aria-label={isInStock ? `Add ${product.title} to cart` : `Get notified when ${product.title} is back in stock`}
          >
          {isLoading ? (
            <>
              <svg className="animate-spin h-5 w-5" fill="none" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              {isInStock ? 'Adding...' : 'Creating...'}
            </>
          ) : isInStock ? (
            <>
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M3 3h2l.4 2M7 13h10l4-8H5.4M7 13L5.4 5M7 13l-2.293 2.293c-.63.63-.184 1.707.707 1.707H17m0 0a2 2 0 100 4 2 2 0 000-4zm-8 2a2 2 0 11-4 0 2 2 0 014 0z"></path>
              </svg>
              Add to Cart
            </>
          ) : (
            <>
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"></path>
              </svg>
              Notify Me
            </>
          )}
          </button>
      </div>

      {/* Stock Notification Modal */}
      <StockNotificationModal
        isOpen={showNotificationModal}
        onClose={() => setShowNotificationModal(false)}
        product={{
          id: Number(product.id),
          title: product.title,
          slug: product.slug || String(product.id)
        }}
      />
    </div>
  );
};

export default ProductCard;
