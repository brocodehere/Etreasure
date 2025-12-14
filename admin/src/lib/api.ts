import { getAccessToken } from './auth';

const BASE_URL = `${import.meta.env.VITE_API_URL || 'https://etreasure-1.onrender.com'}/api/admin`;

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  };
  if (options.headers) {
    Object.assign(headers, options.headers as Record<string, string>);
  }

  const token = getAccessToken();
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const res = await fetch(`${BASE_URL}${path}`, {
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
    // no content
    return undefined as unknown as T;
  }

  return res.json() as Promise<T>;
}

export const api = {
  post: <T>(path: string, body: unknown) =>
    request<T>(path, { method: 'POST', body: JSON.stringify(body) }),
  put: <T>(path: string, body: unknown) =>
    request<T>(path, { method: 'PUT', body: JSON.stringify(body) }),
  delete: (path: string) => request<void>(path, { method: 'DELETE' }),
  get: <T>(path: string) => request<T>(path, { method: 'GET' }),
};
