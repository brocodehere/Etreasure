import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { api } from '../lib/api';
import { saveTokens } from '../lib/auth';

interface LoginResponse {
  accessToken: string;
  refreshToken: string;
  user: {
    id: number;
    email: string;
    name: string;
    roles: string[];
  };
}

export const LoginPage: React.FC = () => {
  const navigate = useNavigate();
  const [email, setEmail] = useState('admin@example.com');
  const [password, setPassword] = useState('ChangeMe123!');
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [showForgotPassword, setShowForgotPassword] = useState(false);
  const [forgotEmail, setForgotEmail] = useState('');
  const [otp, setOtp] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [step, setStep] = useState<'email' | 'otp' | 'reset'>('email');
  const [message, setMessage] = useState<string | null>(null);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError(null);
    try {
      const res = await api.post<LoginResponse>('/auth/login', { email, password });
      saveTokens({ accessToken: res.accessToken, refreshToken: res.refreshToken });
      navigate('/', { replace: true });
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Login failed';
      setError(message);
    } finally {
      setLoading(false);
    }
  }

  async function handleForgotPassword(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError(null);
    setMessage(null);

    try {
      if (step === 'email') {
        // Send OTP
        await api.post('/auth/forgot-password', { email: forgotEmail });
        setStep('otp');
        setMessage('OTP sent to your email. Please check your inbox.');
      } else if (step === 'otp') {
        // Verify OTP
        await api.post('/auth/verify-otp', { email: forgotEmail, otp });
        setStep('reset');
        setMessage('OTP verified. Please set your new password.');
      } else if (step === 'reset') {
        // Reset password
        if (newPassword !== confirmPassword) {
          setError('Passwords do not match');
          return;
        }
        await api.post('/auth/reset-password', { email: forgotEmail, newPassword });
        setMessage('Password reset successfully. Please login with your new password.');
        setTimeout(() => {
          setShowForgotPassword(false);
          resetForgotPasswordForm();
        }, 2000);
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Operation failed';
      setError(message);
    } finally {
      setLoading(false);
    }
  }

  function resetForgotPasswordForm() {
    setForgotEmail('');
    setOtp('');
    setNewPassword('');
    setConfirmPassword('');
    setStep('email');
    setMessage(null);
    setError(null);
  }

  if (showForgotPassword) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-cream px-4">
        <div className="w-full max-w-md bg-white shadow-card rounded-lg p-8 border border-gold/30">
          <h1 className="text-2xl font-playfair text-maroon mb-2">Reset Password</h1>
          <p className="text-sm text-dark/70 mb-6">
            {step === 'email' && 'Enter your email to receive a password reset OTP.'}
            {step === 'otp' && 'Enter the OTP sent to your email.'}
            {step === 'reset' && 'Set your new password.'}
          </p>
          
          <form onSubmit={handleForgotPassword} className="space-y-4">
            {step === 'email' && (
              <div>
                <label htmlFor="forgot-email" className="block text-sm font-medium text-dark mb-1">
                  Email
                </label>
                <input
                  id="forgot-email"
                  type="email"
                  className="w-full rounded-md border border-gold/40 bg-cream/60 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gold focus:border-gold"
                  value={forgotEmail}
                  onChange={(e) => setForgotEmail(e.target.value)}
                  required
                />
              </div>
            )}

            {step === 'otp' && (
              <div>
                <label htmlFor="otp" className="block text-sm font-medium text-dark mb-1">
                  One-Time Password (OTP)
                </label>
                <input
                  id="otp"
                  type="text"
                  className="w-full rounded-md border border-gold/40 bg-cream/60 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gold focus:border-gold"
                  value={otp}
                  onChange={(e) => setOtp(e.target.value)}
                  placeholder="Enter 6-digit OTP"
                  maxLength={6}
                  required
                />
              </div>
            )}

            {step === 'reset' && (
              <>
                <div>
                  <label htmlFor="new-password" className="block text-sm font-medium text-dark mb-1">
                    New Password
                  </label>
                  <input
                    id="new-password"
                    type="password"
                    className="w-full rounded-md border border-gold/40 bg-cream/60 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gold focus:border-gold"
                    value={newPassword}
                    onChange={(e) => setNewPassword(e.target.value)}
                    required
                    minLength={8}
                  />
                </div>
                <div>
                  <label htmlFor="confirm-password" className="block text-sm font-medium text-dark mb-1">
                    Confirm New Password
                  </label>
                  <input
                    id="confirm-password"
                    type="password"
                    className="w-full rounded-md border border-gold/40 bg-cream/60 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gold focus:border-gold"
                    value={confirmPassword}
                    onChange={(e) => setConfirmPassword(e.target.value)}
                    required
                    minLength={8}
                  />
                </div>
              </>
            )}

            {error && (
              <p className="text-sm text-red-600" role="alert">
                {error}
              </p>
            )}

            {message && (
              <p className="text-sm text-green-600" role="status">
                {message}
              </p>
            )}

            <div className="flex gap-3">
              <button
                type="submit"
                disabled={loading}
                className="flex-1 inline-flex items-center justify-center rounded-md bg-maroon text-cream text-sm font-medium px-3 py-2 hover:bg-maroon/90 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-gold disabled:opacity-60"
              >
                {loading ? 'Processing…' : 
                 step === 'email' ? 'Send OTP' :
                 step === 'otp' ? 'Verify OTP' : 'Reset Password'}
              </button>
              <button
                type="button"
                onClick={() => {
                  setShowForgotPassword(false);
                  resetForgotPasswordForm();
                }}
                className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800"
              >
                Cancel
              </button>
            </div>
          </form>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-cream px-4">
      <div className="w-full max-w-md bg-white shadow-card rounded-lg p-8 border border-gold/30">
        <h1 className="text-2xl font-playfair text-maroon mb-2">Admin Sign In</h1>
        <p className="text-sm text-dark/70 mb-6">
          Sign in to manage products, content, and orders.
        </p>
        <form onSubmit={handleSubmit} className="space-y-4" aria-label="Admin login form">
          <div>
            <label htmlFor="email" className="block text-sm font-medium text-dark mb-1">
              Email
            </label>
            <input
              id="email"
              type="email"
              autoComplete="email"
              className="w-full rounded-md border border-gold/40 bg-cream/60 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gold focus:border-gold"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
            />
          </div>
          <div>
            <label htmlFor="password" className="block text-sm font-medium text-dark mb-1">
              Password
            </label>
            <input
              id="password"
              type="password"
              autoComplete="current-password"
              className="w-full rounded-md border border-gold/40 bg-cream/60 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-gold focus:border-gold"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
            />
          </div>
          {error && (
            <p className="text-sm text-red-600" role="alert">
              {error}
            </p>
          )}
          <button
            type="submit"
            disabled={loading}
            className="w-full inline-flex items-center justify-center rounded-md bg-maroon text-cream text-sm font-medium px-3 py-2 mt-2 hover:bg-maroon/90 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-gold disabled:opacity-60"
          >
            {loading ? 'Signing in…' : 'Sign in'}
          </button>
          <div className="text-center">
            <button
              type="button"
              onClick={() => setShowForgotPassword(true)}
              className="text-sm text-gold hover:text-maroon transition-colors"
            >
              Forgot your password?
            </button>
          </div>
          <p className="text-[11px] text-dark/60 mt-2">
            Default admin: <span className="font-mono">admin@example.com / ChangeMe123!</span> (change in production).
          </p>
        </form>
      </div>
    </div>
  );
};
