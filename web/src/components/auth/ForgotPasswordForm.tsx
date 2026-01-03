import { useState } from 'react';

const API_URL = 'https://etreasure-1.onrender.com';

export default function ForgotPasswordForm() {
  const [email, setEmail] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSuccess(null);

    if (!email) {
      setError('Email is required');
      return;
    }

    try {
      setSubmitting(true);
      const res = await fetch(`${API_URL}/api/auth/forgot-password`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email }),
      });

      const data = await res.json();
      
      if (res.status === 404) {
        setError(data?.error || 'Email not found in our system');
        return;
      }
      
      if (!res.ok) {
        setError(data?.error || 'Failed to send reset code');
        return;
      }

      // Redirect to reset password page with email
      window.location.href = `/reset-password?email=${encodeURIComponent(email)}`;
    } catch (err) {
      setError('Failed to send reset code. Please try again.');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <form onSubmit={onSubmit} className="space-y-6">
      {/* Email Field */}
      <div>
        <label htmlFor="email" className="block text-sm font-medium text-dark/70 mb-2">
          Email Address
        </label>
        <div className="relative">
          <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
            <svg className="h-5 w-5 text-dark/40" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M16 12a4 4 0 10-8 0 4 4 0 008 0zm0 0v1.5a2.5 2.5 0 005 0V12a9 9 0 10-9 9m4.5-1.206a8.959 8.959 0 01-4.5 1.207"></path>
            </svg>
          </div>
          <input
            id="email"
            name="email"
            type="email"
            autoComplete="email"
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="appearance-none relative block w-full pl-10 pr-3 py-3 border border-gold/30 bg-cream/30 placeholder-dark/50 text-dark rounded-lg focus:outline-none focus:ring-2 focus:ring-maroon/50 focus:border-maroon transition-all duration-300"
            placeholder="Enter your email address"
          />
        </div>
      </div>

      {/* Submit Button */}
      <div>
        <button
          type="submit"
          disabled={submitting || !email}
          className="group relative w-full flex justify-center py-3 px-4 border border-transparent text-base font-semibold rounded-lg text-white bg-gradient-to-r from-maroon to-maroon/90 hover:from-maroon/90 hover:to-maroon focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-maroon/50 transition-all duration-300 transform hover:scale-[1.02] hover:shadow-lg disabled:opacity-60 disabled:cursor-not-allowed"
        >
          <span className="absolute left-0 inset-y-0 flex items-center pl-3">
            <svg className="w-5 h-5 text-gold/80 group-hover:text-gold transition-colors" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"></path>
            </svg>
          </span>
          {submitting ? 'Sending...' : 'Send Reset Code'}
        </button>
      </div>

      {/* Back to Login */}
      <div className="text-center">
        <a href="/login" className="text-sm text-maroon hover:text-maroon/80 transition-colors">
          Back to Login
        </a>
      </div>

      {/* Success Message */}
      {success && (
        <div className="rounded-lg bg-green-50 border border-green-200 p-4">
          <div className="flex">
            <div className="flex-shrink-0">
              <svg className="h-5 w-5 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
              </svg>
            </div>
            <div className="ml-3">
              <h3 className="text-sm font-medium text-green-800">Reset Code Sent</h3>
              <div className="mt-2 text-sm text-green-700">
                <p>{success}</p>
                <p className="mt-2">
                  <a href="/reset-password" className="font-medium text-green-800 hover:text-green-700">
                    Go to Reset Password Page
                  </a>
                </p>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Error Message */}
      {error && (
        <div className="rounded-lg bg-red-50 border border-red-200 p-4">
          <div className="flex">
            <div className="flex-shrink-0">
              <svg className="h-5 w-5 text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
              </svg>
            </div>
            <div className="ml-3">
              <h3 className="text-sm font-medium text-red-800">Error</h3>
              <div className="mt-2 text-sm text-red-700">
                <p>{error}</p>
              </div>
            </div>
          </div>
        </div>
      )}
    </form>
  );
}
