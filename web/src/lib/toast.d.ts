// Toast notification types
declare global {
  interface Window {
    toastManager: {
      show(message: string, type?: 'success' | 'error' | 'warning' | 'info', duration?: number): string;
      success(message: string, duration?: number): string;
      error(message: string, duration?: number): string;
      warning(message: string, duration?: number): string;
      info(message: string, duration?: number): string;
      remove(toastId: string): void;
      clear(): void;
    };
    showToast: (message: string, type?: 'success' | 'error' | 'warning' | 'info', duration?: number) => string;
    showSuccess: (message: string, duration?: number) => string;
    showError: (message: string, duration?: number) => string;
    showWarning: (message: string, duration?: number) => string;
    showInfo: (message: string, duration?: number) => string;
  }
}

export {};
