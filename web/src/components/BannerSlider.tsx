import React, { useState, useEffect } from 'react';

interface Banner {
  id: string;
  image: string;
  link?: string;
  alt?: string;
}

interface BannerSliderProps {
  banners: Banner[];
  autoSlide?: boolean;
  slideInterval?: number;
}

const BannerSlider: React.FC<BannerSliderProps> = ({ 
  banners, 
  autoSlide = true, 
  slideInterval = 5000 
}) => {
  const [currentIndex, setCurrentIndex] = useState(0);
  const [isLoading, setIsLoading] = useState(banners.length === 0);

  useEffect(() => {
    if (banners.length > 0) {
      setIsLoading(false);
    }
  }, [banners]);

  useEffect(() => {
    if (!autoSlide || banners.length <= 1) return;

    const interval = setInterval(() => {
      setCurrentIndex((prevIndex) => 
        prevIndex === banners.length - 1 ? 0 : prevIndex + 1
      );
    }, slideInterval);

    return () => clearInterval(interval);
  }, [autoSlide, slideInterval, banners.length]);

  const goToSlide = (index: number) => {
    setCurrentIndex(index);
  };

  const goToPrevious = () => {
    setCurrentIndex(currentIndex === 0 ? banners.length - 1 : currentIndex - 1);
  };

  const goToNext = () => {
    setCurrentIndex(currentIndex === banners.length - 1 ? 0 : currentIndex + 1);
  };

  if (isLoading) {
    return (
      <div className="relative w-full h-[70vh] bg-gray-200 animate-pulse">
        <div className="absolute inset-0 flex items-center justify-center">
          <div className="text-gray-400">Loading banners...</div>
        </div>
      </div>
    );
  }

  if (!banners || banners.length === 0) {
    return (
      <div className="relative w-full h-[70vh] bg-gray-100 flex items-center justify-center">
        <div className="text-gray-500">No banners available</div>
      </div>
    );
  }

  const currentBanner = banners[currentIndex];

  return (
    <div className="relative w-full overflow-hidden z-10 pt-0">
      {/* Banner Container */}
      <div className="relative w-full h-[70vh] ">
        <div className="relative w-full h-full">
          {/* Banner Image */}
          <img
            src={currentBanner.image}
            alt={currentBanner.alt || `Banner ${currentIndex + 1}`}
            className="w-full h-full object-cover"
            style={{
              objectPosition: 'center',
            }}
          />
          
          {/* Optional overlay for better text visibility */}
          <div className="absolute inset-0 bg-black bg-opacity-0 pointer-events-none" />
        </div>

        {/* Explore More Button - Positioned at bottom-left */}
        {currentBanner.link && (
          <div className="absolute bottom-4 left-4 z-20">
            <a
              href={currentBanner.link}
              className="inline-flex items-center px-4 py-2 sm:px-6 sm:py-3 bg-white hover:bg-gray-100 text-gray-900 font-semibold rounded-lg shadow-lg transition-all duration-300 hover:shadow-xl hover:scale-105 focus:outline-none focus:ring-2 focus:ring-white focus:ring-offset-2 focus:ring-offset-gray-900"
              target="_blank"
              rel="noopener noreferrer"
            >
              <span className="text-sm sm:text-base">Explore More</span>
              <svg
                className="w-4 h-4 sm:w-5 sm:h-5 ml-2"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M9 5l7 7-7 7"
                />
              </svg>
            </a>
          </div>
        )}

        {/* Navigation Arrows - Only show if more than 1 banner */}
        {banners.length > 1 && (
          <>
            <button
              onClick={goToPrevious}
              className="absolute left-2 sm:left-4 top-1/2 -translate-y-1/2 z-10 p-2 sm:p-3 bg-white hover:bg-gray-100 rounded-full shadow-lg transition-all duration-300 hover:shadow-xl hover:scale-110 focus:outline-none focus:ring-2 focus:ring-white"
              aria-label="Previous banner"
            >
              <svg
                className="w-5 h-5 sm:w-6 sm:h-6 text-gray-800"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M15 19l-7-7 7-7"
                />
              </svg>
            </button>
            <button
              onClick={goToNext}
              className="absolute right-2 sm:right-4 top-1/2 -translate-y-1/2 z-10 p-2 sm:p-3 bg-white hover:bg-gray-100 rounded-full shadow-lg transition-all duration-300 hover:shadow-xl hover:scale-110 focus:outline-none focus:ring-2 focus:ring-white"
              aria-label="Next banner"
            >
              <svg
                className="w-5 h-5 sm:w-6 sm:h-6 text-gray-800"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M9 5l7 7-7 7"
                />
              </svg>
            </button>
          </>
        )}

        {/* Dots Indicator - Only show if more than 1 banner */}
        {banners.length > 1 && (
          <div className="absolute bottom-4 right-4 z-10 flex space-x-2">
            {banners.map((_, index) => (
              <button
                key={index}
                onClick={() => goToSlide(index)}
                className={`w-2 h-2 sm:w-3 sm:h-3 rounded-full transition-all duration-300 focus:outline-none ${
                  index === currentIndex
                    ? 'bg-white scale-125 shadow-lg'
                    : 'bg-white bg-opacity-50 hover:bg-opacity-75'
                }`}
                aria-label={`Go to banner ${index + 1}`}
              />
            ))}
          </div>
        )}
      </div>

      {/* Progress Bar - Only show if more than 1 banner and autoSlide is enabled */}
      {banners.length > 1 && autoSlide && (
        <div className="absolute bottom-0 left-0 right-0 h-1 bg-black bg-opacity-20 z-10">
          <div
            className="h-full bg-white transition-all duration-100 ease-linear"
            style={{
              width: '0%',
              animation: `progress ${slideInterval}ms linear infinite`,
            }}
          />
        </div>
      )}
    </div>
  );
};

export default BannerSlider;
