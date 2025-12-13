import { useState } from 'react';

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

export default function LoginForm() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [rememberMe, setRememberMe] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [info, setInfo] = useState<string | null>(null);

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setInfo(null);

    if (!email || !password) {
      setError('Email and password are required');
      return;
    }

    try {
      setSubmitting(true);
      const res = await fetch(`${API_URL}/api/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password, rememberMe }),
        credentials: 'include',
      });
      const data = await res.json();
      if (!res.ok) {
        setError(data?.error || 'Login failed');
        return;
      }

      if (data.accessToken) {
        sessionStorage.setItem('accessToken', data.accessToken);
      }

      window.location.href = '/account/dashboard';
    } catch (err) {
      console.error(err);
      setError('Login failed. Please try again.');
    } finally {
      setSubmitting(false);
    }
  };

  const onForgotPassword = async () => {
    setError(null);
    setInfo(null);

    if (!email) {
      setError('Enter your email first to reset password');
      return;
    }

    try {
      const res = await fetch(`${API_URL}/api/auth/forgot-password`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email }),
        credentials: 'include',
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) {
        setError(data?.error || 'Could not start password reset');
        return;
      }
      setInfo('If this email exists, we have sent a reset code. Please check your inbox.');
    } catch (err) {
      console.error(err);
      setError('Could not start password reset. Please try again.');
    }
  };

  return (
    <form onSubmit={onSubmit} className="max-w-md mx-auto space-y-4 bg-white p-6 rounded-xl shadow-md border border-gold/30">
      <h1 className="text-2xl font-playfair text-maroon mb-2 text-center">Welcome back</h1>
      {error && <p className="text-sm text-red-600 text-center">{error}</p>}
      {info && !error && <p className="text-sm text-emerald-700 text-center">{info}</p>}
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
        <div className="flex items-center justify-between">
          <label className="block text-sm font-medium text-gray-700">Password</label>
          <button
            type="button"
            onClick={onForgotPassword}
            className="text-xs text-maroon hover:underline"
          >
            Forgot password?
          </button>
        </div>
        <input
          type="password"
          required
          className="w-full border rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-maroon/60"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          placeholder="Your password"
        />
      </div>
      <label className="flex items-center gap-2 text-sm text-gray-700">
        <input
          type="checkbox"
          checked={rememberMe}
          onChange={(e) => setRememberMe(e.target.checked)}
          className="rounded border-gray-400 text-maroon focus:ring-maroon/70"
        />
        <span>Remember me</span>
      </label>
      <button
        type="submit"
        disabled={submitting}
        className="w-full btn-primary rounded-lg text-center disabled:opacity-60 disabled:cursor-not-allowed"
      >
        {submitting ? 'Signing in…' : 'Log in'}
      </button>
      <p className="text-xs text-center text-gray-600">
        New here? <a href="/account/signup" className="text-maroon underline">Create an account</a>
      </p>
    </form>
  );
}
