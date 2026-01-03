import { useState, useEffect } from 'react';

const API_URL = 'https://etreasure-1.onrender.com';

export default function PasswordResetForm() {
  const [email, setEmail] = useState('');
  const [otp, setOtp] = useState(['', '', '', '', '', '']);
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  // Get email from URL parameter on component mount
  useEffect(() => {
    const urlParams = new URLSearchParams(window.location.search);
    const emailParam = urlParams.get('email');
    if (emailParam) {
      setEmail(emailParam);
    }
  }, []);

  const handleOtpChange = (index: number, value: string) => {
    if (value.length > 1) return;
    
    const newOtp = [...otp];
    newOtp[index] = value;
    setOtp(newOtp);

    // Auto-focus next input
    if (value && index < 5) {
      const nextInput = document.getElementById(`otp-${index + 1}`) as HTMLInputElement;
      nextInput?.focus();
    }
  };

  const handleKeyDown = (index: number, e: React.KeyboardEvent) => {
    if (e.key === 'Backspace' && !otp[index] && index > 0) {
      const prevInput = document.getElementById(`otp-${index - 1}`) as HTMLInputElement;
      prevInput?.focus();
    }
  };

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSuccess(null);

    // Validate inputs
    if (!email) {
      setError('Email is required');
      return;
    }

    const otpValue = otp.join('');
    if (otpValue.length !== 6) {
      setError('Please enter all 6 digits of the verification code');
      return;
    }

    if (newPassword.length < 8) {
      setError('Password must be at least 8 characters long');
      return;
    }

    if (newPassword !== confirmPassword) {
      setError('Passwords do not match');
      return;
    }

    try {
      setSubmitting(true);
      const res = await fetch(`${API_URL}/api/auth/reset-password`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ 
          email, 
          otp: otpValue, 
          newPassword 
        }),
      });

      const data = await res.json();
      if (!res.ok) {
        setError(data?.error || 'Failed to reset password');
        return;
      }

      setSuccess('Password reset successfully! Redirecting to login...');
      
      // Redirect to login after 2 seconds
      setTimeout(() => {
        window.location.href = '/login?message=Password reset successfully! Please log in with your new password.';
      }, 2000);
    } catch (err) {
      setError('Failed to reset password. Please try again.');
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
            readOnly
            className="appearance-none relative block w-full pl-10 pr-3 py-3 border border-gold/30 bg-cream/30 placeholder-dark/50 text-dark rounded-lg focus:outline-none focus:ring-2 focus:ring-maroon/50 focus:border-maroon transition-all duration-300 bg-gray-100"
            placeholder="Enter your email address"
          />
        </div>
      </div>

      {/* OTP Input */}
      <div>
        <label className="block text-sm font-medium text-dark/70 mb-2">
          Verification Code
        </label>
        <div className="flex justify-center gap-2">
          {otp.map((digit, index) => (
            <input
              key={index}
              id={`otp-${index}`}
              type="text"
              maxLength={1}
              className="w-12 h-12 text-center border rounded-lg focus:outline-none focus:ring-2 focus:ring-maroon/60 text-lg font-semibold"
              value={digit}
              onChange={(e) => handleOtpChange(index, e.target.value)}
              onKeyDown={(e) => handleKeyDown(index, e)}
              autoFocus={index === 0}
            />
          ))}
        </div>
        <p className="text-xs text-dark/50 mt-2 text-center">
          Enter the 6-digit code sent to your email
        </p>
      </div>

      {/* New Password Field */}
      <div>
        <label htmlFor="newPassword" className="block text-sm font-medium text-dark/70 mb-2">
          New Password
        </label>
        <div className="relative">
          <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
            <svg className="h-5 w-5 text-dark/40" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"></path>
            </svg>
          </div>
          <input
            id="newPassword"
            name="newPassword"
            type="password"
            autoComplete="new-password"
            required
            value={newPassword}
            onChange={(e) => setNewPassword(e.target.value)}
            className="appearance-none relative block w-full pl-10 pr-3 py-3 border border-gold/30 bg-cream/30 placeholder-dark/50 text-dark rounded-lg focus:outline-none focus:ring-2 focus:ring-maroon/50 focus:border-maroon transition-all duration-300"
            placeholder="Enter your new password"
          />
        </div>
        <p className="text-xs text-dark/50 mt-2">
          Must be at least 8 characters long
        </p>
      </div>

      {/* Confirm Password Field */}
      <div>
        <label htmlFor="confirmPassword" className="block text-sm font-medium text-dark/70 mb-2">
          Confirm New Password
        </label>
        <div className="relative">
          <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
            <svg className="h-5 w-5 text-dark/40" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"></path>
            </svg>
          </div>
          <input
            id="confirmPassword"
            name="confirmPassword"
            type="password"
            autoComplete="new-password"
            required
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            className="appearance-none relative block w-full pl-10 pr-3 py-3 border border-gold/30 bg-cream/30 placeholder-dark/50 text-dark rounded-lg focus:outline-none focus:ring-2 focus:ring-maroon/50 focus:border-maroon transition-all duration-300"
            placeholder="Confirm your new password"
          />
        </div>
      </div>

      {/* Submit Button */}
      <div>
        <button
          type="submit"
          disabled={submitting || !email || otp.join('').length !== 6 || !newPassword || !confirmPassword}
          className="group relative w-full flex justify-center py-3 px-4 border border-transparent text-base font-semibold rounded-lg text-white bg-gradient-to-r from-maroon to-maroon/90 hover:from-maroon/90 hover:to-maroon focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-maroon/50 transition-all duration-300 transform hover:scale-[1.02] hover:shadow-lg disabled:opacity-60 disabled:cursor-not-allowed"
        >
          <span className="absolute left-0 inset-y-0 flex items-center pl-3">
            <svg className="w-5 h-5 text-gold/80 group-hover:text-gold transition-colors" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"></path>
            </svg>
          </span>
          {submitting ? 'Resetting...' : 'Reset Password'}
        </button>
      </div>

      {/* Back to Forgot Password */}
      <div className="text-center">
        <a href="/forgot-password" className="text-sm text-maroon hover:text-maroon/80 transition-colors">
          Didn't receive code? Request again
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
              <h3 className="text-sm font-medium text-green-800">Success</h3>
              <div className="mt-2 text-sm text-green-700">
                <p>{success}</p>
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
