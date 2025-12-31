// Updated wishlist functions to use API
const PUBLIC_API_URL = 'https://etreasure-1.onrender.com';

async function updateWishlistCount() {
    try {
        const response = await fetch(`${PUBLIC_API_URL}/api/wishlist`);
        const wishlistData = await response.json();
        
        if (response.ok) {
            const wishlistCount = document.querySelector('[data-wishlist-count]');
            if (wishlistCount) {
                wishlistCount.textContent = wishlistData.count || 0;
            }
        }
    } catch (error) {
    }
}

async function removeFromWishlist(productId) {
    try {
        const response = await fetch(`https://etreasure-1.onrender.com/api/wishlist/remove/${productId}`, {
            method: 'DELETE'
        });
        
        if (!response.ok) {
            throw new Error('Failed to remove from wishlist');
        }
        
        updateWishlistCount();
        renderWishlist();
    } catch (error) {
        alert('Failed to remove from wishlist');
    }
}

async function addToCartFromWishlist(productId, title, price, imageUrl) {
    try {
        const response = await fetch(`${PUBLIC_API_URL}/api/cart/add`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ product_id: productId, quantity: 1 })
        });
        
        if (!response.ok) {
            throw new Error('Failed to add to cart');
        }
        
        // Update cart count
        updateCartCount();
        
        // Show notification
        alert(`${title} added to cart!`);
    } catch (error) {
        alert('Failed to add to cart');
    }
}

async function updateCartCount() {
    try {
        const response = await fetch(`${PUBLIC_API_URL}/api/cart`);
        const cartData = await response.json();
        
        if (response.ok) {
            const cartCount = document.querySelector('[data-cart-count]');
            if (cartCount) {
                cartCount.textContent = cartData.count || 0;
            }
        }
    } catch (error) {
    }
}

// Make functions available globally
window.removeFromWishlist = removeFromWishlist;
window.addToCartFromWishlist = addToCartFromWishlist;
