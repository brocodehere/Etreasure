import React from 'react';
import BannerSlider from './BannerSlider';
import useBannerAPI from '../hooks/useBannerAPI';

interface BannerContainerProps {
  apiUrl: string;
  autoSlide?: boolean;
  slideInterval?: number;
  fallbackBanners?: Array<{
    id: string;
    image: string;
    link?: string;
    alt?: string;
  }>;
}

const BannerContainer: React.FC<BannerContainerProps> = ({
  apiUrl,
  autoSlide = true,
  slideInterval = 5000,
  fallbackBanners = []
}) => {
    
  const { banners, loading, error, refetch } = useBannerAPI(apiUrl);

  // Use fallback banners if API fails or returns empty
  const displayBanners = (error || banners.length === 0) && fallbackBanners.length > 0 ? fallbackBanners : banners;

  if (loading) {
    return (
      <div className="w-full mt-0">
        <BannerSlider banners={[]} autoSlide={false} />
      </div>
    );
  }

  if (error && fallbackBanners.length === 0) {
    return (
      <div className="w-full p-8 text-center">
        <div className="text-red-500 mb-4">Error loading banners: {error}</div>
        <button
          onClick={refetch}
          className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 transition-colors"
        >
          Retry
        </button>
      </div>
    );
  }

  return (
    <div className="w-full mt-0">
      <BannerSlider 
        banners={displayBanners} 
        autoSlide={autoSlide} 
        slideInterval={slideInterval}
      />
    </div>
  );
};

export default BannerContainer;
