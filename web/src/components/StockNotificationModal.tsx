import { useState } from 'react';

interface StockNotificationModalProps {
  isOpen: boolean;
  onClose: () => void;
  product: {
    id: number;
    title: string;
    slug: string;
  };
}

export default function StockNotificationModal({ isOpen, onClose, product }: StockNotificationModalProps) {
  const [notificationType, setNotificationType] = useState<'email' | 'mobile'>('email');
  const [email, setEmail] = useState('');
  const [mobileNumber, setMobileNumber] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSuccess(null);

    // Validate inputs
    if (notificationType === 'email') {
      if (!email) {
        setError('Email is required');
        return;
      }
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
      if (!emailRegex.test(email)) {
        setError('Please enter a valid email address');
        return;
      }
    } else {
      if (!mobileNumber) {
        setError('Mobile number is required');
        return;
      }
      // Basic mobile number validation (10 digits for India)
      const mobileRegex = /^[6-9]\d{9}$/;
      if (!mobileRegex.test(mobileNumber.replace(/\D/g, ''))) {
        setError('Please enter a valid 10-digit mobile number');
        return;
      }
    }

    try {
      setSubmitting(true);
      const API_URL = 'https://etreasure-1.onrender.com';
      
      const res = await fetch(`${API_URL}/api/stock-notifications`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          productId: product.id,
          productSlug: product.slug,
          notificationType,
          [notificationType === 'email' ? 'email' : 'mobileNumber']: 
            notificationType === 'email' ? email : mobileNumber
        }),
      });

      const data = await res.json();
      if (!res.ok) {
        setError(data?.error || 'Failed to create notification request');
        return;
      }

      setSuccess('We will notify you when this product is back in stock!');
      
      // Close modal after 2 seconds
      setTimeout(() => {
        onClose();
        // Reset form
        setEmail('');
        setMobileNumber('');
        setSuccess(null);
      }, 2000);
    } catch (err) {
      setError('Failed to create notification request. Please try again.');
    } finally {
      setSubmitting(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 z-50 flex items-center justify-center p-4">
      <div className="bg-white rounded-2xl shadow-2xl max-w-md w-full max-h-[90vh] overflow-y-auto">
        {/* Header */}
        <div className="p-6 border-b border-gray-200">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="text-xl font-bold text-gray-900">Stock Notification</h3>
              <p className="text-sm text-gray-600 mt-1">
                Get notified when "{product.title}" is back in stock
              </p>
            </div>
            <button
              onClick={onClose}
              className="text-gray-400 hover:text-gray-600 transition-colors"
            >
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M6 18L18 6M6 6l12 12"></path>
              </svg>
            </button>
          </div>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} className="p-6">
          {/* Notification Type Selection */}
          <div className="mb-6">
            <label className="block text-sm font-medium text-gray-700 mb-3">
              How would you like to be notified?
            </label>
            <div className="grid grid-cols-2 gap-3">
              <button
                type="button"
                onClick={() => setNotificationType('email')}
                className={`p-3 rounded-lg border-2 transition-all ${
                  notificationType === 'email'
                    ? 'border-maroon bg-maroon/5 text-maroon'
                    : 'border-gray-200 text-gray-600 hover:border-gray-300'
                }`}
              >
                <svg className="w-5 h-5 mx-auto mb-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"></path>
                </svg>
                <span className="text-sm font-medium">Email</span>
              </button>
              <button
                type="button"
                onClick={() => setNotificationType('mobile')}
                className={`p-3 rounded-lg border-2 transition-all ${
                  notificationType === 'mobile'
                    ? 'border-maroon bg-maroon/5 text-maroon'
                    : 'border-gray-200 text-gray-600 hover:border-gray-300'
                }`}
              >
                <svg className="w-5 h-5 mx-auto mb-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z"></path>
                </svg>
                <span className="text-sm font-medium">Mobile</span>
              </button>
            </div>
          </div>

          {/* Email Input */}
          {notificationType === 'email' && (
            <div className="mb-6">
              <label htmlFor="email" className="block text-sm font-medium text-gray-700 mb-2">
                Email Address
              </label>
              <input
                type="email"
                id="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-maroon focus:border-maroon"
                placeholder="Enter your email address"
                required
              />
            </div>
          )}

          {/* Mobile Input */}
          {notificationType === 'mobile' && (
            <div className="mb-6">
              <label htmlFor="mobile" className="block text-sm font-medium text-gray-700 mb-2">
                Mobile Number
              </label>
              <div className="relative">
                <span className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-500 text-sm">
                  +91
                </span>
                <input
                  type="tel"
                  id="mobile"
                  value={mobileNumber}
                  onChange={(e) => setMobileNumber(e.target.value.replace(/\D/g, '').slice(0, 10))}
                  className="w-full pl-12 pr-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-maroon focus:border-maroon"
                  placeholder="9876543210"
                  maxLength={10}
                  required
                />
              </div>
              <p className="text-xs text-gray-500 mt-1">We'll send SMS when the product is back in stock</p>
            </div>
          )}

          {/* Success Message */}
          {success && (
            <div className="mb-6 p-4 bg-green-50 border border-green-200 rounded-lg">
              <div className="flex items-center">
                <svg className="w-5 h-5 text-green-500 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                </svg>
                <p className="text-sm text-green-800">{success}</p>
              </div>
            </div>
          )}

          {/* Error Message */}
          {error && (
            <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg">
              <div className="flex items-center">
                <svg className="w-5 h-5 text-red-500 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                </svg>
                <p className="text-sm text-red-800">{error}</p>
              </div>
            </div>
          )}

          {/* Actions */}
          <div className="flex space-x-3">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 px-4 py-3 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={submitting}
              className="flex-1 px-4 py-3 bg-gradient-to-r from-maroon to-burgundy text-white rounded-lg hover:from-gold hover:to-gold hover:text-maroon transition-all disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {submitting ? 'Creating...' : 'Create Notification'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
