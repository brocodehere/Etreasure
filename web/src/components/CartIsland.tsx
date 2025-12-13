import React, { useState } from 'react';

interface CartItem {
  id: number;
  name: string;
  price: string;
  image: string;
  quantity: number;
  size: string;
}

interface CartIslandProps {
  isOpen: boolean;
  onClose: () => void;
}

const CartIsland: React.FC<CartIslandProps> = ({ isOpen, onClose }) => {
  const [cartItems, setCartItems] = useState<CartItem[]>([
    {
      id: 1,
      name: "Handcrafted Leather Hand Bag",
      price: "3500",
      image: "/images/products/hand-bag.avif",
      quantity: 1,
      size: "Standard"
    },
    {
      id: 2,
      name: "Canvas Tote Bag",
      price: "1200",
      image: "/images/products/tote-bag.avif",
      quantity: 2,
      size: "One Size"
    }
  ]);

  const updateQuantity = (id: number, change: number) => {
    setCartItems(items =>
      items.map(item =>
        item.id === id
          ? { ...item, quantity: Math.max(1, Math.min(10, item.quantity + change)) }
          : item
      )
    );
  };

  const removeItem = (id: number) => {
    setCartItems(items => items.filter(item => item.id !== id));
  };

  const calculateSubtotal = () => {
    return cartItems.reduce((total, item) => {
      const price = parseInt(item.price);
      return total + (price * item.quantity);
    }, 0);
  };

  const calculateShipping = () => {
    return cartItems.length > 0 ? (calculateSubtotal() > 5000 ? 0 : 250) : 0;
  };

  const calculateTotal = () => {
    return calculateSubtotal() + calculateShipping();
  };

  const getTotalItems = () => {
    return cartItems.reduce((sum, item) => sum + item.quantity, 0);
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 overflow-hidden">
      <div 
        className="absolute inset-0 bg-black/50 transition-opacity duration-300"
        onClick={onClose}
      />

      <div className="absolute right-0 top-0 h-full w-full max-w-md bg-white shadow-2xl transform transition-transform duration-300">
        <div className="flex flex-col h-full">
          <div className="flex items-center justify-between p-6 border-b border-gold/20">
            <h2 className="font-playfair text-2xl font-bold text-maroon">
              Shopping Cart ({cartItems.length})
            </h2>
            <button
              onClick={onClose}
              className="text-dark/60 hover:text-maroon transition-colors duration-300"
            >
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M6 18L18 6M6 6l12 12"></path>
              </svg>
            </button>
          </div>

          <div className="flex-1 overflow-y-auto p-6">
            {cartItems.length === 0 ? (
              <div className="text-center py-12">
                <div className="w-20 h-20 bg-cream rounded-full flex items-center justify-center mx-auto mb-4">
                  <svg className="w-10 h-10 text-maroon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M16 11V7a4 4 0 00-8 0v4M5 9h14l1 12H4L5 9z"></path>
                  </svg>
                </div>
                <p className="text-dark/80 mb-4">Your cart is empty</p>
                <button onClick={onClose} className="btn-primary">
                  Continue Shopping
                </button>
              </div>
            ) : (
              <div className="space-y-6">
                {cartItems.map((item) => (
                  <div key={item.id} className="flex space-x-4 bg-cream p-4 rounded-lg">
                    <img
                      src={item.image}
                      alt={item.name}
                      className="w-20 h-20 object-cover rounded-lg"
                    />
                    <div className="flex-1">
                      <h3 className="font-semibold text-dark mb-1">{item.name}</h3>
                      <p className="text-sm text-dark/60 mb-2">Size: {item.size}</p>
                      <div className="flex items-center justify-between">
                        <div className="flex items-center space-x-2">
                          <button
                            onClick={() => updateQuantity(item.id, -1)}
                            className="w-8 h-8 bg-white border border-gold/30 rounded-full flex items-center justify-center hover:bg-gold hover:text-white transition-colors duration-300"
                          >
                            -
                          </button>
                          <span className="font-semibold w-8 text-center">{item.quantity}</span>
                          <button
                            onClick={() => updateQuantity(item.id, 1)}
                            className="w-8 h-8 bg-white border border-gold/30 rounded-full flex items-center justify-center hover:bg-gold hover:text-white transition-colors duration-300"
                          >
                            +
                          </button>
                        </div>
                        <span className="font-semibold text-maroon">₹{item.price}</span>
                      </div>
                    </div>
                    <button
                      onClick={() => removeItem(item.id)}
                      className="text-dark/60 hover:text-red-500 transition-colors duration-300"
                    >
                      <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path>
                      </svg>
                    </button>
                  </div>
                ))}
              </div>
            )}
          </div>

          {cartItems.length > 0 && (
            <div className="border-t border-gold/20 p-6 space-y-4">
              <div className="bg-gradient-to-r from-cream to-white p-6 rounded-lg border border-gold/20">
                <h3 className="font-semibold text-maroon mb-4">Order Summary</h3>
                <div className="space-y-3">
                  <div className="flex justify-between text-dark/80">
                    <span>Subtotal ({getTotalItems()} items)</span>
                    <span>₹{calculateSubtotal().toLocaleString()}</span>
                  </div>
                  <div className="flex justify-between text-dark/80">
                    <span>Shipping</span>
                    <span>{calculateShipping() === 0 ? 'FREE' : `₹${calculateShipping()}`}</span>
                  </div>
                  {calculateShipping() === 0 && (
                    <p className="text-sm text-green-600">You have qualified for free shipping!</p>
                  )}
                  <div className="border-t border-gold/20 pt-3">
                    <div className="flex justify-between font-semibold text-lg text-maroon">
                      <span>Total</span>
                      <span>₹{calculateTotal().toLocaleString()}</span>
                    </div>
                  </div>
                </div>
              </div>

              <div className="space-y-3">
                <button className="w-full btn-primary py-3">
                  Proceed to Checkout
                </button>
                <button onClick={onClose} className="w-full btn-secondary py-3">
                  Continue Shopping
                </button>
              </div>

              <div className="text-center text-sm text-dark/60">
                <p>Secure checkout powered by Razorpay</p>
                <div className="flex justify-center space-x-2 mt-2">
                  <div className="w-8 h-5 bg-gray-200 rounded flex items-center justify-center text-xs font-bold">VISA</div>
                  <div className="w-8 h-5 bg-gray-200 rounded flex items-center justify-center text-xs font-bold">MC</div>
                  <div className="w-8 h-5 bg-gray-200 rounded flex items-center justify-center text-xs font-bold">UPI</div>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default CartIsland;
