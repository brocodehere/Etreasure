import React from 'react';

interface LoadingSpinnerProps {
  size?: 'sm' | 'md' | 'lg';
  className?: string;
  text?: string;
}

export function LoadingSpinner({ size = 'md', className = '', text }: LoadingSpinnerProps) {
  const sizeClasses = {
    sm: 'w-4 h-4',
    md: 'w-6 h-6', 
    lg: 'w-8 h-8'
  };

  return (
    <div className={`flex items-center justify-center ${className}`}>
      <div className="inline-flex items-center gap-2">
        <div 
          className={`
            ${sizeClasses[size]} 
            border-2 border-gray-200 border-t-gold 
            rounded-full animate-spin
          `}
        />
        {text && (
          <span className="text-gray-600 text-sm">{text}</span>
        )}
      </div>
    </div>
  );
}

interface LoadingStateProps {
  isLoading: boolean;
  error?: Error | null;
  children: React.ReactNode;
  fallback?: React.ReactNode;
  errorFallback?: React.ReactNode;
}

export function LoadingState({ 
  isLoading, 
  error, 
  children, 
  fallback, 
  errorFallback 
}: LoadingStateProps) {
  if (isLoading) {
    return fallback || (
      <div className="flex items-center justify-center p-8">
        <LoadingSpinner size="lg" text="Loading..." />
      </div>
    );
  }

  if (error) {
    return errorFallback || (
      <div className="flex items-center justify-center p-8">
        <div className="text-red-600 text-center">
          <div className="text-lg font-semibold mb-2">Error</div>
          <div className="text-sm">{error.message}</div>
        </div>
      </div>
    );
  }

  return <>{children}</>;
}

interface LoadingButtonProps {
  isLoading: boolean;
  children: React.ReactNode;
  className?: string;
  disabled?: boolean;
  loadingText?: string;
  onClick?: () => void;
  type?: 'button' | 'submit' | 'reset';
}

export function LoadingButton({ 
  isLoading, 
  children, 
  className = '', 
  disabled = false,
  loadingText = 'Loading...',
  onClick,
  type = 'button'
}: LoadingButtonProps) {
  return (
    <button
      type={type}
      onClick={onClick}
      className={`
        inline-flex items-center justify-center gap-2
        px-4 py-2 rounded-lg font-medium
        transition-colors duration-200
        disabled:opacity-50 disabled:cursor-not-allowed
        ${className}
      `}
      disabled={disabled || isLoading}
    >
      {isLoading && <LoadingSpinner size="sm" />}
      {isLoading ? loadingText : children}
    </button>
  );
}
