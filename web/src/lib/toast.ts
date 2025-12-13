// Toast notification functions for TypeScript
// These functions are made available globally by the toast.js script

// Declare global window interface to include toast functions
declare global {
  interface Window {
    showSuccess: (message: string, duration?: number) => string;
    showError: (message: string, duration?: number) => string;
    showWarning: (message: string, duration?: number) => string;
    showInfo: (message: string, duration?: number) => string;
    showToast: (message: string, type?: 'success' | 'error' | 'warning' | 'info', duration?: number) => string;
  }
}

export const showSuccess = (message: string, duration?: number): string => {
  if (typeof window !== 'undefined' && window.showSuccess) {
    return window.showSuccess(message, duration);
  }
  return '';
};

export const showError = (message: string, duration?: number): string => {
  if (typeof window !== 'undefined' && window.showError) {
    return window.showError(message, duration);
  }
  return '';
};

export const showWarning = (message: string, duration?: number): string => {
  if (typeof window !== 'undefined' && window.showWarning) {
    return window.showWarning(message, duration);
  }
  return '';
};

export const showInfo = (message: string, duration?: number): string => {
  if (typeof window !== 'undefined' && window.showInfo) {
    return window.showInfo(message, duration);
  }
  return '';
};

export const showToast = (message: string, type?: 'success' | 'error' | 'warning' | 'info', duration?: number): string => {
  if (typeof window !== 'undefined' && window.showToast) {
    return window.showToast(message, type, duration);
  }
  return '';
};
