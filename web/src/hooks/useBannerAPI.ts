import { useState, useEffect } from 'react';

interface Banner {
  id: string;
  image: string;
  link?: string;
  alt?: string;
}

interface BannerAPIResponse {
  items: Array<{
    id: string;
    title: string;
    image_url: string;
    link_url?: string | null;
    is_active: boolean;
    sort_order: number;
    starts_at?: string | null;
    ends_at?: string | null;
    created_at: string;
    updated_at: string;
  }>;
}

const useBannerAPI = (apiUrl: string) => {
  const [banners, setBanners] = useState<Banner[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchBanners = async () => {
    try {
      setLoading(true);
      setError(null);
      
      // Add timeout to prevent hanging
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), 10000); // 10 second timeout
      
      const response = await fetch(apiUrl, {
        method: 'GET',
        headers: {
          'Accept': 'application/json',
          'Content-Type': 'application/json'
        },
        mode: 'cors',
        signal: controller.signal
      });
      
      clearTimeout(timeoutId);
      
      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(`HTTP error! status: ${response.status}, body: ${errorText}`);
      }

      const data: BannerAPIResponse = await response.json();
      
      if (!data.items || !Array.isArray(data.items)) {
        throw new Error('Invalid API response format - items is not an array');
      }

      // Transform API response to our component format
      const transformedBanners = data.items
        .filter(banner => banner.is_active && banner.image_url)
        .sort((a, b) => a.sort_order - b.sort_order)
        .map(banner => ({
          id: banner.id,
          image: banner.image_url,
          link: banner.link_url || undefined,
          alt: banner.title || `Banner ${banner.id}`
        }));

      setBanners(transformedBanners);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error occurred';
      setError(errorMessage);
      console.error('Error fetching banners:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (apiUrl) {
      fetchBanners();
    }
  }, [apiUrl]);

  const refetch = () => {
    if (apiUrl) {
      fetchBanners();
    }
  };

  return {
    banners,
    loading,
    error,
    refetch,
  };
};

export default useBannerAPI;
