import { useState, type FC } from 'react';

interface ImageVariant {
  url: string;
  alt?: string;
}

interface Props {
  hero: ImageVariant;
  images?: ImageVariant[];
}

const ProductGallery: FC<Props> = ({ hero, images = [] }) => {
  const [mainImage, setMainImage] = useState<ImageVariant>(hero);
  const [isZoomed, setIsZoomed] = useState(false);

  // Combine hero image with additional images, removing duplicates
  const allImages = [hero, ...images.filter(img => img.url !== hero.url)];

  const handleImageClick = (image: ImageVariant) => {
    setMainImage(image);
    setIsZoomed(false);
  };

  const handleZoomToggle = () => {
    setIsZoomed(!isZoomed);
  };

  const alt = mainImage.alt || 'Product image';

  return (
    <div className="space-y-4">
      {/* Main Image */}
      <div className="relative overflow-hidden rounded-lg bg-cream">
        <div className={`transition-transform duration-300 cursor-zoom-in ${isZoomed ? 'scale-150' : 'scale-100'}`}
             onClick={handleZoomToggle}>
          <picture>
            <source srcSet={mainImage.url} type="image/webp" />
            <img
              src={mainImage.url}
              alt={alt}
              className="w-full h-auto object-cover"
              loading="eager"
            />
          </picture>
        </div>
        
        {/* Zoom Indicator */}
        <div className="absolute top-4 right-4 bg-white/80 px-3 py-1 rounded-full text-sm text-dark/80">
          {isZoomed ? 'Click to unzoom' : 'Click to zoom'}
        </div>
      </div>

      {/* Thumbnail Gallery */}
      {allImages.length > 1 && (
        <div className="grid grid-cols-4 gap-4">
          {allImages.map((image, index) => (
            <button
              key={index}
              className={`relative overflow-hidden rounded-lg border-2 transition-all duration-300 ${
                mainImage.url === image.url ? 'border-gold shadow-gold' : 'border-transparent hover:border-gold/50'
              }`}
              onClick={() => handleImageClick(image)}
            >
              <img
                src={image.url}
                alt={`${image.alt || 'Product image'} ${index + 1}`}
                className="w-full h-20 object-cover transition-transform duration-300 hover:scale-105"
                loading="lazy"
              />
            </button>
          ))}
        </div>
      )}
    </div>
  );
};

export default ProductGallery;
