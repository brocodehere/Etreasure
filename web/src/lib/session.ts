// Session management utilities for Etreasure cart functionality
export interface SessionData {
  sessionId: string;
  expiresAt: number;
}

export interface CartSessionManager {
  getSessionId(): string;
  isValidSession(): boolean;
  refreshSession(): Promise<void>;
  clearSession(): void;
}

class ProductionCartSession implements CartSessionManager {
  private readonly SESSION_KEY = 'etreasure_session';
  private readonly SESSION_DURATION = 24 * 60 * 60 * 1000; // 24 hours

  getSessionId(): string {
    if (typeof window === 'undefined') return '';
    
    // Try to get session ID from cookie first (backend sets this)
    const sessionId = this.getCookie('session_id');
    if (sessionId) {
      return sessionId;
    }
    
    // Fallback to sessionStorage
    try {
      const sessionData = this.getSessionData();
      return sessionData?.sessionId || '';
    } catch (error) {
      console.error('Error getting session ID:', error);
      return '';
    }
  }

  private getSessionData(): SessionData | null {
    try {
      const sessionStr = sessionStorage.getItem(this.SESSION_KEY);
      if (!sessionStr) return null;
      
      const sessionData: SessionData = JSON.parse(sessionStr);
      
      // Check if session has expired
      if (Date.now() > sessionData.expiresAt) {
        this.clearSession();
        return null;
      }
      
      return sessionData;
    } catch (error) {
      console.error('Error parsing session data:', error);
      this.clearSession();
      return null;
    }
  }

  async createSession(): Promise<string> {
    // Session creation is handled by backend when cart operations are performed
    // This method is kept for compatibility but actual session creation
    // happens when first cart operation is called
    return '';
  }

  isValidSession(): boolean {
    const sessionId = this.getSessionId();
    return sessionId !== null && sessionId !== '';
  }

  async refreshSession(): Promise<void> {
    // Session refresh is handled by backend automatically
    // through cookie expiration settings
    const currentSession = this.getSessionData();
    if (currentSession && Date.now() > currentSession.expiresAt) {
      this.clearSession();
    }
  }

  clearSession(): void {
    try {
      sessionStorage.removeItem(this.SESSION_KEY);
    } catch (error) {
      console.error('Error clearing session:', error);
    }
  }

  private getCookie(name: string): string {
    if (typeof document === 'undefined') return '';
    
    const value = `; ${document.cookie}`;
    const parts = value.split(`; ${name}=`);
    
    if (parts.length === 2) {
      return parts.pop()?.split(';').shift() || '';
    }
    
    return '';
  }
}

// Export singleton instance
export const cartSession = new ProductionCartSession();

// Enhanced API wrapper with session management
export async function apiRequestWithSession(url: string, options: RequestInit = {}): Promise<Response> {
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
    ...options.headers,
  };

  // The backend handles session ID through cookies automatically
  // No need to manually add X-Session-ID header or authorization

  // Use dynamic URL based on environment
  const baseUrl = import.meta.env.DEV ? 'https://etreasure-1.onrender.com' : 'https://etreasure-1.onrender.com';
  
  const response = await fetch(`${baseUrl}${url}`, {
    ...options,
    headers,
    credentials: 'include', // Important for session cookies
  });

  // Handle session-related errors
  if (response.status === 401) {
    // Session might be expired, clear local session data
    cartSession.clearSession();
    
    // For cart operations, we can retry once as backend will create new session
    if (url.startsWith('/api/cart')) {
      const baseUrl = import.meta.env.DEV ? 'https://etreasure-1.onrender.com' : 'https://etreasure-1.onrender.com';
      const retryResponse = await fetch(`${baseUrl}${url}`, {
        ...options,
        headers,
        credentials: 'include',
      });
      return retryResponse;
    }
  }

  return response;
}
