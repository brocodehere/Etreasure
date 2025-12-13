import type { FC, ReactNode } from 'react';

interface Props {
  title: string;
  description?: string | null;
  priceCents: number;
  currency: string;
  availability: string;
  addToCartSlot?: ReactNode;
}

const ProductDetails: FC<Props> = ({
  title,
  description,
  priceCents,
  currency,
  availability,
  addToCartSlot,
}) => {
  const price = (priceCents / 100).toFixed(2);
  return (
    <div className="space-y-6">
      <h1 className="font-playfair text-4xl lg:text-5xl font-bold text-maroon mb-4">{title}</h1>
      <div className="flex items-baseline space-x-4 mb-4">
        <span className="text-4xl font-bold text-gold">
          {currency} {price}
        </span>
        <span className="text-sm px-3 py-1 rounded-full border border-gold/40 text-dark/80">
          {availability === 'InStock' ? 'In stock' : 'Out of stock'}
        </span>
      </div>
      {description && (
        <p className="text-dark/70 leading-relaxed text-sm md:text-base">{description}</p>
      )}
      {addToCartSlot}
    </div>
  );
};

export default ProductDetails;
