// Updated wishlist functions to use API
async function updateWishlistCount() {
    try {
        const response = await fetch('/api/wishlist');
        const wishlistData = await response.json();
        
        if (response.ok) {
            const wishlistCount = document.querySelector('[data-wishlist-count]');
            if (wishlistCount) {
                wishlistCount.textContent = wishlistData.count || 0;
            }
        }
    } catch (error) {
        console.error('Error updating wishlist count:', error);
    }
}

async function removeFromWishlist(productId) {
    try {
        const response = await fetch(`/api/wishlist/remove/${productId}`, {
            method: 'DELETE'
        });
        
        if (!response.ok) {
            throw new Error('Failed to remove from wishlist');
        }
        
        updateWishlistCount();
        renderWishlist();
    } catch (error) {
        console.error('Error removing from wishlist:', error);
        alert('Failed to remove from wishlist');
    }
}

async function addToCartFromWishlist(productId, title, price, imageUrl) {
    try {
        const response = await fetch('/api/cart/add', {
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
        console.error('Error adding to cart:', error);
        alert('Failed to add to cart');
    }
}

async function updateCartCount() {
    try {
        const response = await fetch('/api/cart');
        const cartData = await response.json();
        
        if (response.ok) {
            const cartCount = document.querySelector('[data-cart-count]');
            if (cartCount) {
                cartCount.textContent = cartData.count || 0;
            }
        }
    } catch (error) {
        console.error('Error updating cart count:', error);
    }
}

// Make functions available globally
window.removeFromWishlist = removeFromWishlist;
window.addToCartFromWishlist = addToCartFromWishlist;
