import { useState } from 'react';

interface OTPVerificationProps {
  email: string;
  password: string;
  fullName: string;
  onSuccess: () => void;
  onBack: () => void;
}

export default function OTPVerification({ email, password, fullName, onSuccess, onBack }: OTPVerificationProps) {
  const [otp, setOtp] = useState(['', '', '', '', '', '']);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [resending, setResending] = useState(false);

  const API_URL = 'https://etreasure-1.onrender.com';

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

  const handleVerify = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    const otpValue = otp.join('');
    if (otpValue.length !== 6) {
      setError('Please enter all 6 digits');
      return;
    }

    try {
      setSubmitting(true);
      const res = await fetch(`${API_URL}/api/auth/verify-signup-otp`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, otp: otpValue }),
      });

      const data = await res.json();
      if (!res.ok) {
        setError(data?.error || 'Verification failed');
        return;
      }

      onSuccess();
    } catch (err) {
      setError('Verification failed. Please try again.');
    } finally {
      setSubmitting(false);
    }
  };

  const handleResend = async () => {
    setError(null);
    setResending(true);

    try {
      const res = await fetch(`${API_URL}/api/auth/send-signup-otp`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password, fullName }),
      });

      const data = await res.json();
      if (!res.ok) {
        setError(data?.error || 'Failed to resend OTP');
        return;
      }

      // Clear OTP inputs
      setOtp(['', '', '', '', '', '']);
      document.getElementById('otp-0')?.focus();
    } catch (err) {
      setError('Failed to resend OTP. Please try again.');
    } finally {
      setResending(false);
    }
  };

  return (
    <div className="max-w-md mx-auto bg-white p-6 rounded-xl shadow-md border border-gold/30">
      <h1 className="text-2xl font-playfair text-maroon mb-2 text-center">Verify Your Email</h1>
      <p className="text-gray-600 text-center mb-6">
        We've sent a 6-digit code to <strong>{email}</strong>
      </p>

      {error && <p className="text-sm text-red-600 text-center mb-4">{error}</p>}

      <form onSubmit={handleVerify} className="space-y-6">
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

        <button
          type="submit"
          disabled={submitting || otp.join('').length !== 6}
          className="w-full btn-primary rounded-lg text-center disabled:opacity-60 disabled:cursor-not-allowed"
        >
          {submitting ? 'Verifying…' : 'Verify Email'}
        </button>

        <div className="text-center space-y-2">
          <button
            type="button"
            onClick={handleResend}
            disabled={resending}
            className="text-sm text-maroon hover:underline disabled:opacity-60"
          >
            {resending ? 'Resending…' : "Didn't receive the code? Resend"}
          </button>
          <br />
          <button
            type="button"
            onClick={onBack}
            className="text-sm text-gray-600 hover:underline"
          >
            Back to signup
          </button>
        </div>
      </form>
    </div>
  );
}
