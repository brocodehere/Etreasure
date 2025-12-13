import { api } from './api';
import { getAccessToken } from './auth';

export interface MediaUploadResponse {
  key: string;
  url: string;
}

// Raw request function for FormData uploads
async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers: Record<string, string> = {};
  if (options.headers) {
    Object.assign(headers, options.headers as Record<string, string>);
  }

  // Don't set Content-Type for FormData - let browser set it with boundary
  if (!(options.body instanceof FormData)) {
    headers['Content-Type'] = 'application/json';
  }

  const token = getAccessToken();
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const res = await fetch(`http://localhost:8080/api/admin${path}`, {
    ...options,
    headers,
  });

  if (!res.ok) {
    let message = 'Request failed';
    try {
      const data = await res.json();
      if (data && typeof (data as any).error === 'string') message = (data as any).error;
    } catch {
      // ignore
    }
    throw new Error(message);
  }

  if (res.status === 204) {
    return undefined as unknown as T;
  }

  return res.json() as Promise<T>;
}

export async function uploadImage(file: File, type: 'product' | 'banner' | 'category'): Promise<MediaUploadResponse> {
  const formData = new FormData();
  formData.append('file', file);
  formData.append('type', type);
  
  const response = await request<MediaUploadResponse>('/media/upload', {
    method: 'POST',
    body: formData,
  });
  
  return response;
}

export function getPublicImageUrl(path: string): string {
  // If it's already a full URL, return as is
  if (path.startsWith('http')) {
    return path;
  }
  
  // If it's a localhost proxy URL (new format), return as-is
  if (path.startsWith('http://localhost:8080/api/public/media/')) {
    return path;
  }
  
  // If it's a local path starting with /uploads/, prepend localhost
  if (path.startsWith('/uploads/')) {
    return `http://localhost:8080${path}`;
  }
  
  // Otherwise, assume it's an R2 path and use the public base URL
  const publicBaseUrl = import.meta.env.VITE_R2_PUBLIC_BASE_URL || 'https://static.ethnictreasures.co.in';
  return `${publicBaseUrl}/${path}`;
}
