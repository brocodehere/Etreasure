import { useState } from 'react';
import PasswordStrengthIndicator from './PasswordStrengthIndicator';

// Temporarily hardcoded for debugging
const API_URL = 'https://etreasure-1.onrender.com';

export default function SignupForm() {
  const [fullName, setFullName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [rememberMe, setRememberMe] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!email || !password) {
      setError('Email and password are required');
      return;
    }

    if (password !== confirmPassword) {
      setError('Passwords do not match');
      return;
    }

    try {
      setSubmitting(true);
      const res = await fetch(`${API_URL}/api/auth/signup`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password, fullName, rememberMe }),
        credentials: 'include',
      });
      const data = await res.json();
      if (!res.ok) {
        setError(data?.error || 'Signup failed');
        return;
      }

      if (data.accessToken) {
        // Store access token in sessionStorage (short-lived)
        sessionStorage.setItem('accessToken', data.accessToken);
      }

      window.location.href = '/account/dashboard';
    } catch (err) {
      console.error(err);
      setError('Signup failed. Please try again.');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <form onSubmit={onSubmit} className="max-w-md mx-auto space-y-4 bg-white p-6 rounded-xl shadow-md border border-gold/30">
      <h1 className="text-2xl font-playfair text-maroon mb-2 text-center">Create your account</h1>
      {error && <p className="text-sm text-red-600 text-center">{error}</p>}
      <div className="space-y-1">
        <label className="block text-sm font-medium text-gray-700">Full name</label>
        <input
          type="text"
          className="w-full border rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-maroon/60"
          value={fullName}
          onChange={(e) => setFullName(e.target.value)}
          placeholder="Your name"
        />
      </div>
      <div className="space-y-1">
        <label className="block text-sm font-medium text-gray-700">Email</label>
        <input
          type="email"
          required
          className="w-full border rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-maroon/60"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          placeholder="you@example.com"
        />
      </div>
      <div className="space-y-1">
        <label className="block text-sm font-medium text-gray-700">Password</label>
        <div className="relative">
          <input
            type={showPassword ? 'text' : 'password'}
            required
            minLength={8}
            className="w-full border rounded-lg px-3 py-2 pr-10 focus:outline-none focus:ring-2 focus:ring-maroon/60"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="At least 8 characters"
          />
          <button
            type="button"
            onClick={() => setShowPassword((v) => !v)}
            className="absolute right-0 top-1/2 -translate-y-1/2 px-3 flex items-center text-gray-500 hover:text-maroon focus:outline-none"
            aria-label={showPassword ? 'Hide password' : 'Show password'}
          >
            {showPassword ? '🙈' : '👁️'}
          </button>
        </div>
        <PasswordStrengthIndicator password={password} />
      </div>
      <div className="space-y-1">
        <label className="block text-sm font-medium text-gray-700">Confirm password</label>
        <div className="relative">
          <input
            type={showConfirmPassword ? 'text' : 'password'}
            required
            minLength={8}
            className="w-full border rounded-lg px-3 py-2 pr-10 focus:outline-none focus:ring-2 focus:ring-maroon/60"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            placeholder="Re-enter your password"
          />
          <button
            type="button"
            onClick={() => setShowConfirmPassword((v) => !v)}
            className="absolute right-0 top-1/2 -translate-y-1/2 px-3 flex items-center text-gray-500 hover:text-maroon focus:outline-none"
            aria-label={showConfirmPassword ? 'Hide confirm password' : 'Show confirm password'}
          >
            {showConfirmPassword ? '🙈' : '👁️'}
          </button>
        </div>
      </div>
      <label className="flex items-center gap-2 text-sm text-gray-700">
        <input
          type="checkbox"
          checked={rememberMe}
          onChange={(e) => setRememberMe(e.target.checked)}
          className="rounded border-gray-400 text-maroon focus:ring-maroon/70"
        />
        <span>Remember me (keep me signed in longer)</span>
      </label>
      <button
        type="submit"
        disabled={submitting}
        className="w-full btn-primary rounded-lg text-center disabled:opacity-60 disabled:cursor-not-allowed"
      >
        {submitting ? 'Creating account…' : 'Sign up'}
      </button>
      <p className="text-xs text-center text-gray-600">
        Already have an account? <a href="/account/login" className="text-maroon underline">Log in</a>
      </p>
    </form>
  );
}
