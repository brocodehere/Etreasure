// Toast notification system
class ToastManager {
  constructor() {
    this.container = null;
    // Only initialize in browser environment
    if (typeof window !== 'undefined' && typeof document !== 'undefined') {
      this.init();
    }
  }

  init() {
    // Create toast container if it doesn't exist
    if (!this.container && typeof document !== 'undefined') {
      this.container = document.createElement('div');
      this.container.id = 'toast-container';
      this.container.className = 'fixed top-4 right-4 z-50 space-y-2';
      document.body.appendChild(this.container);
    }
  }

  show(message, type = 'info', duration = 3000) {
    // Only run in browser environment
    if (typeof document === 'undefined') return;
    
    const toast = document.createElement('div');
    const toastId = Date.now();
    
    // Toast styles based on type
    const typeStyles = {
      success: 'bg-green-500 text-white',
      error: 'bg-red-500 text-white',
      warning: 'bg-yellow-500 text-black',
      info: 'bg-blue-500 text-white'
    };

    const icons = {
      success: '<svg class="w-5 h-5" fill="currentColor" viewBox="oldt 0 20 20"><path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"></path></svg>',
      error: '<svg class="w-5 h-5" fill="currentColor" viewBox ​​viewBox="0 0 20 20"><path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd"></path></svg>',
      warning: '<svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd"></path></svg>',
      info: '<svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd"></path></svg>'
    };

    toast.id = `toast-${toastId}`;
    toast.className = `${typeStyles[type]} px-4 py-3 rounded-lg shadow-lg flex items-center space-x-2 min-w-[250px] max-w-md transform transition-all duration-300 translate-x-full`;
    
    toast.innerHTML = `
      <div class="flex-shrink-0">
        ${icons[type]}
      </div>
      <div class="flex-1">
        <p class="text-sm font-medium">${message}</p>
      </div>
      <button class="flex-shrink-0 ml-2 hover:opacity-75" onclick="toastManager.remove('${toastId}')">
        <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
          <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd"></path>
        </svg>
      </button>
    `;

    this.container.appendChild(toast);

    // Animate in
    setTimeout(() => {
      toast.classList.remove('translate-x-full');
      toast.classList.add('translate-x-0');
    }, 100);

    // Auto remove after duration
    if (duration > 0) {
      setTimeout(() => {
        this.remove(toastId);
      }, duration);
    }

    return toastId;
  }

  remove(toastId) {
    // Only run in browser environment
    if (typeof document === 'undefined') return;
    
    const toast = document.getElementById(`toast-${toastId}`);
    if (toast) {
      toast.classList.add('translate-x-full');
      setTimeout(() => {
        if (toast.parentNode) {
          toast.parentNode.removeChild(toast);
        }
      }, 300);
    }
  }

  success(message, duration = 3000) {
    return this.show(message, 'success', duration);
  }

  error(message, duration = 5000) {
    return this.show(message, 'error', duration);
  }

  warning(message, duration = 4000) {
    return this.show(message, 'warning', duration);
  }

  info(message, duration = 3000) {
    return this.show(message, 'info', duration);
  }

  clear() {
    if (this.container) {
      this.container.innerHTML = '';
    }
  }
}

// Global toast manager instance - only create in browser environment
let toastManager = null;

// Initialize toast manager only in browser
function initToastManager() {
  if (!toastManager && typeof window !== 'undefined' && typeof document !== 'undefined') {
    toastManager = new ToastManager();
  }
  return toastManager;
}

// Initialize toast manager and create functions
let manager = null;
let initPromise = null;

function getManager() {
  if (!initPromise) {
    initPromise = new Promise((resolve) => {
      if (typeof window !== 'undefined' && typeof document !== 'undefined') {
        if (document.readyState === 'loading') {
          document.addEventListener('DOMContentLoaded', () => {
            manager = initToastManager();
            resolve(manager);
          });
        } else {
          manager = initToastManager();
          resolve(manager);
        }
      } else {
        // SSR environment - create dummy manager
        manager = {
          show: () => '',
          success: () => '',
          error: () => '',
          warning: () => '',
          info: () => ''
        };
        resolve(manager);
      }
    });
  }
  return initPromise;
}

// ES6 exports for modern imports
export const showToast = async (message, type, duration) => {
  const mgr = await getManager();
  return mgr.show(message, type, duration);
};

export const showSuccess = async (message, duration) => {
  const mgr = await getManager();
  return mgr.success(message, duration);
};

export const showError = async (message, duration) => {
  const mgr = await getManager();
  return mgr.error(message, duration);
};

export const showWarning = async (message, duration) => {
  const mgr = await getManager();
  return mgr.warning(message, duration);
};

export const showInfo = async (message, duration) => {
  const mgr = await getManager();
  return mgr.info(message, duration);
};

// Legacy global exports for compatibility
if (typeof window !== 'undefined') {
  getManager().then(mgr => {
    window.toastManager = mgr;
    window.showToast = (message, type, duration) => mgr.show(message, type, duration);
    window.showSuccess = (message, duration) => mgr.success(message, duration);
    window.showError = (message, duration) => mgr.error(message, duration);
    window.showWarning = (message, duration) => mgr.warning(message, duration);
    window.showInfo = (message, duration) => mgr.info(message, duration);
  });
}
