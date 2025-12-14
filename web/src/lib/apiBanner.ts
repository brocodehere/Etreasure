export interface Banner {
  id: string;
  title: string;
  image_url: string;
  link_url?: string;
  is_active: boolean;
  sort_order: number;
  starts_at?: string;
  ends_at?: string;
  created_at: string;
  updated_at: string;
}

export interface BannerResponse {
  items: Banner[];
}

export async function fetchBanners(): Promise<BannerResponse> {
  const apiUrl = import.meta.env.PUBLIC_API_URL || 'https://etreasure-1.onrender.com';
  
  try {
    const response = await fetch(`${apiUrl}/api/public/banners`);
    if (!response.ok) {
      throw new Error(`Failed to fetch banners: ${response.statusText}`);
    }
    return await response.json();
  } catch (error) {
    console.error('Error fetching banners:', error);
    // Return empty array on error to fallback to static images
    return { items: [] };
  }
}
