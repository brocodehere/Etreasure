const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

async function postJSON(path, body) {
  const res = await fetch(`${API_URL}${path}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    const text = await res.text().catch(() => '');
    throw new Error(text || `Request failed with ${res.status}`);
  }
  return res.json();
}

export const api = {
  products: {
    async search(query) {
      if (!query) return [];
      try {
        const data = await postJSON('/api/products/search', { query });
        return data.items ?? data;
      } catch (err) {
        console.error('Product search failed', err);
        return [];
      }
    },
  },
};

export default api;
