// API helper functions for shop functionality
export interface Category {
  id?: string;
  uuid_id?: string;
  name: string;
  slug: string;
  product_count?: number;
  image_key?: string;
  image_url?: string;
}

export interface Product {
  id: string;
  title: string;
  slug: string;
  price_cents: number;
  original_price_cents?: number;
  currency?: string;
  description?: string;
  short_description?: string;
  image_key?: string;
  image_url?: string;
  category_id?: string;
  stock_quantity: number;
  created_at: string;
}

export interface ProductsResponse {
  items: Product[];
  total: number;
  page: number;
  limit: number;
}

export interface CategoriesResponse {
  items: Category[];
}

export interface Banner {
  id: string;
  title: string;
  desktop_image_key?: string;
  desktop_image_url?: string;
  laptop_image_key?: string;
  laptop_image_url?: string;
  mobile_image_key?: string;
  mobile_image_url?: string;
  is_active: boolean;
  sort_order: number;
  starts_at?: string;
  ends_at?: string;
  created_at: string;
  updated_at: string;
}

export interface BannersResponse {
  items: Banner[];
}

export const API_BASE_URL = import.meta.env?.PUBLIC_API_URL as string || 
  (import.meta.env.DEV ? 'http://localhost:8080' : 'http://localhost:8080');
const R2_BASE_URL = import.meta.env?.PUBLIC_R2_BASE_URL as string || 'https://pub-1a3924a6c6994107be6fe9f3ed794c0a.r2.dev';

// Import session management
import { cartSession, apiRequestWithSession } from './session';

// Authentication types
export interface User {
  id: number;
  email: string;
  name: string;
  roles: string[];
}

export interface AuthResponse {
  accessToken: string;
  refreshToken: string;
  user: User;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface SignupRequest {
  email: string;
  password: string;
  fullName: string;
  rememberMe?: boolean;
}

// Get auth token from localStorage
function getAuthToken(): string | null {
  if (typeof localStorage !== 'undefined') {
    return localStorage.getItem('accessToken');
  }
  return null;
}

// Get current user from localStorage
export function getCurrentUser(): User | null {
  if (typeof localStorage !== 'undefined') {
    const userStr = localStorage.getItem('user');
    return userStr ? JSON.parse(userStr) : null;
  }
  return null;
}

// Check if user is authenticated
export function isAuthenticated(): boolean {
  return !!getAuthToken() && !!getCurrentUser();
}

// Clear auth data
export function clearAuth() {
  if (typeof localStorage !== 'undefined') {
    localStorage.removeItem('accessToken');
    localStorage.removeItem('refreshToken');
    localStorage.removeItem('user');
  }
}

// Refresh access token using refresh token
async function refreshAccessToken(): Promise<boolean> {
  try {
    const refreshToken = localStorage.getItem('refreshToken');
    if (!refreshToken) {
      return false;
    }

    const response = await fetch(`${API_BASE_URL}/api/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refreshToken }),
    });

    if (response.ok) {
      const data = await response.json();
      localStorage.setItem('accessToken', data.accessToken);
      if (data.refreshToken) {
        localStorage.setItem('refreshToken', data.refreshToken);
      }
      return true;
    }
  } catch (error) {
    // Token refresh failed
  }
  
  return false;
}

// API helper with authentication and automatic token refresh
async function apiRequest(url: string, options: RequestInit = {}): Promise<Response> {
  let token = getAuthToken();
  
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const fullUrl = `${API_BASE_URL}${url}`;

  try {
    let response = await fetch(fullUrl, {
      method: options.method,
      headers,
      body: options.body,
      credentials: 'include', // Include cookies for session management
    });

    // Handle 401 unauthorized - try to refresh token
    if (response.status === 401 && token) {
      const refreshed = await refreshAccessToken();
      if (refreshed) {
        // Retry request with new token
        token = getAuthToken();
        const newHeaders = {
          'Content-Type': 'application/json',
          ...(token && { Authorization: `Bearer ${token}` }),
          ...options.headers,
        };

        response = await fetch(`${API_BASE_URL}${url}`, {
          ...options,
          headers: newHeaders,
          credentials: 'include', // Include cookies for session management
        });
      } else {
        // Refresh failed, clear auth and redirect to login
        clearAuth();
        if (typeof window !== 'undefined') {
          window.location.href = '/login';
        }
      }
    } else if (response.status === 401) {
      // No token or refresh failed, clear auth and redirect
      clearAuth();
      if (typeof window !== 'undefined') {
        window.location.href = '/login';
      }
    }

    return response;
  } catch (error) {
    throw error;
  }
}

// Session-only API wrapper for cart operations (no authentication)
async function apiRequestSessionOnly(url: string, options: RequestInit = {}): Promise<Response> {
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
    ...options.headers,
  };

  const response = await fetch(`${API_BASE_URL}${url}`, {
    ...options,
    headers,
    credentials: 'include', // Include cookies for session management only
  });

  return response;
}

// Authentication API functions
export async function login(credentials: LoginRequest): Promise<AuthResponse> {
  const response = await fetch(`${API_BASE_URL}/api/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(credentials),
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Login failed');
  }

  return response.json();
}

export async function signup(userData: SignupRequest): Promise<AuthResponse> {
  const response = await fetch(`${API_BASE_URL}/api/auth/signup`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(userData),
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Signup failed');
  }

  return response.json();
}

export async function logout(): Promise<void> {
  try {
    await apiRequest('/api/auth/logout', { method: 'POST' });
  } finally {
    clearAuth();
  }
}

export async function getCurrentUserProfile(): Promise<User> {
  const response = await apiRequest('/api/auth/me');
  
  if (!response.ok) {
    throw new Error('Failed to get user profile');
  }

  return response.json();
}

// Helper to ensure absolute R2 URLs
export function getImageUrl(imageKey?: string, imageUrl?: string): string {
  if (imageUrl && imageUrl.startsWith('http')) {
    return imageUrl;
  }
  if (imageKey) {
    return `${R2_BASE_URL}/${imageKey}`;
  }
  // Fallback placeholder
  return `${R2_BASE_URL}/product-placeholder.webp`;
}

// API functions
export async function fetchCategories(): Promise<CategoriesResponse> {
  try {
    const response = await fetch(`${API_BASE_URL}/api/public/categories`);
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    const data = await response.json();
    
    // Ensure image URLs are absolute
    if (data.items) {
      data.items = data.items.map((category: Category) => ({
        ...category,
        image_url: getImageUrl(category.image_key, category.image_url)
      }));
    }
    
    return data;
  } catch (error) {
    return { items: [] };
  }
}

export async function fetchBanners(): Promise<BannersResponse> {
  try {
    const response = await fetch(`${API_BASE_URL}/api/public/banners`);
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    const data = await response.json();
    
    // Ensure image URLs are absolute
    if (data.items) {
      data.items = data.items.map((banner: Banner) => ({
        ...banner,
        desktop_image_url: getImageUrl(banner.desktop_image_key, banner.desktop_image_url),
        laptop_image_url: getImageUrl(banner.laptop_image_key, banner.laptop_image_url),
        mobile_image_url: getImageUrl(banner.mobile_image_key, banner.mobile_image_url)
      }));
    }
    
    return data;
  } catch (error) {
    return { items: [] };
  }
}

export async function fetchProducts(params: {
  category?: string;
  search?: string;
  sort?: 'price_asc' | 'price_desc' | 'newest' | 'name_asc' | 'name_desc';
  min_price?: string;
  max_price?: string;
  in_stock?: string;
  page?: number;
  limit?: number;
}): Promise<ProductsResponse> {
  try {
    const searchParams = new URLSearchParams();
    
    if (params.category) searchParams.set('category', params.category);
    if (params.search) searchParams.set('search', params.search);
    if (params.sort) searchParams.set('sort', params.sort);
    if (params.min_price) searchParams.set('min_price', params.min_price);
    if (params.max_price) searchParams.set('max_price', params.max_price);
    if (params.in_stock) searchParams.set('in_stock', params.in_stock);
    if (params.page) searchParams.set('page', params.page.toString());
    if (params.limit) searchParams.set('limit', params.limit.toString());
    
    const response = await fetch(`${API_BASE_URL}/api/products?${searchParams}`);
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    const data = await response.json();
    
    // Ensure image URLs are absolute
    if (data.items) {
      data.items = data.items.map((product: Product) => ({
        ...product,
        image_url: getImageUrl(product.image_key, product.image_url)
      }));
    }
    
    return data;
  } catch (error) {
    return { items: [], total: 0, page: 1, limit: 12 };
  }
}

function dispatchShopEvent(eventName: string) {
  if (typeof window === 'undefined') return;
  window.dispatchEvent(new Event(eventName));
}

// Cart and Wishlist functions with session management (no authentication required)
export const addToCart = async (productId: string, quantity = 1, price?: number) => {
  try {
    const requestBody: any = { product_id: productId, quantity };
    if (price !== undefined && price > 0) {
      requestBody.price = price;
    }
    
    const response = await apiRequestSessionOnly('/api/cart/add', {
      method: 'POST',
      body: JSON.stringify(requestBody),
    });

    const data = await response.json();
    
    dispatchShopEvent('cart-updated');
    return data;
  } catch (error) {
    throw error;
  }
};

export const getCart = async () => {
  try {
    const response = await apiRequestSessionOnly('/api/cart');
    
    if (!response.ok) {
      throw new Error(`Failed to get cart: ${response.status}`);
    }
    
    const data = await response.json();
    return data;
  } catch (error) {
    console.error('Get cart error:', error);
    // Return empty cart for better UX instead of throwing
    return { items: [], total: 0 };
  }
};

export const removeFromCart = async (itemId: string) => {
  try {
    const response = await apiRequestSessionOnly(`/api/cart/${itemId}`, {
      method: 'DELETE',
    });

    const data = await response.json();
    dispatchShopEvent('cart-updated');
    return data;
  } catch (error) {
    console.error('Remove from cart error:', error);
    throw error;
  }
};

export const clearCart = async () => {
  try {
    const response = await apiRequestSessionOnly('/api/cart/clear', {
      method: 'POST',
    });

    const data = await response.json();
    dispatchShopEvent('cart-updated');
    return data;
  } catch (error) {
    console.error('Clear cart error:', error);
    throw error;
  }
};

export const toggleWishlist = async (productId: string, price?: number) => {
  try {
    const requestBody: any = { product_id: productId };
    if (price !== undefined && price > 0) {
      requestBody.price = price;
    }
    
    const response = await apiRequestSessionOnly('/api/wishlist/toggle', {
      method: 'POST',
      body: JSON.stringify(requestBody),
    });

    const data = await response.json();
    dispatchShopEvent('wishlist-updated');
    return data;
  } catch (error) {
    console.error('Toggle wishlist error:', error);
    throw error;
  }
};

export const getWishlist = async () => {
  try {
    const response = await apiRequestSessionOnly('/api/wishlist');
    const data = await response.json();
    return data;
  } catch (error) {
    console.error('Get wishlist error:', error);
    throw error;
  }
};

export const removeFromWishlist = async (productId: string) => {
  try {
    const response = await apiRequestSessionOnly(`/api/wishlist/${productId}`, {
      method: 'DELETE',
    });

    const data = await response.json();
    dispatchShopEvent('wishlist-updated');
    return data;
  } catch (error) {
    console.error('Remove from wishlist error:', error);
    throw error;
  }
};

// Content management types
export interface ContentPage {
  id: string;
  title: string;
  slug: string;
  content: string;
  type: string;
  is_active: boolean;
  meta_title: string;
  meta_description: string;
  created_at: string;
  updated_at: string;
}

// Offers management types
export interface Offer {
  id: string;
  title: string;
  description: string;
  discount_type: string;
  discount_value: number;
  applies_to: string;
  applies_to_ids: string;
  min_order_amount: number;
  usage_limit: number;
  usage_count: number;
  is_active: boolean;
  starts_at: string;
  ends_at: string;
  created_at: string;
  updated_at: string;
}

export interface OffersResponse {
  items: Offer[];
  total: number;
  page: number;
  limit: number;
}

// Fetch content page by slug
export async function fetchContent(slug: string): Promise<ContentPage | null> {
  try {
    const response = await fetch(`${API_BASE_URL}/api/public/content/pages/${slug}`);
    if (!response.ok) {
      if (response.status === 404) {
        return null;
      }
      throw new Error(`Failed to fetch content: ${response.statusText}`);
    }
    return await response.json();
  } catch (error) {
    return null;
  }
}

// Fetch offers from database
export async function fetchOffers(): Promise<OffersResponse> {
  try {
    const response = await fetch(`${API_BASE_URL}/api/public/offers`);
    if (!response.ok) {
            return { items: [], total: 0, page: 1, limit: 50 };
    }
    return await response.json();
  } catch (error) {
    // Return empty data instead of throwing error to prevent build failures
    return { items: [], total: 0, page: 1, limit: 50 };
  }
}

// Debounced search function
export function debounce<T extends (...args: any[]) => any>(
  func: T,
  wait: number
): (...args: Parameters<T>) => void {
  let timeout: ReturnType<typeof setTimeout>;
  return (...args: Parameters<T>) => {
    clearTimeout(timeout);
    timeout = setTimeout(() => func(...args), wait);
  };
}
