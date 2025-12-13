import type { FC } from 'react';

interface RelatedProduct {
  id: number;
  slug: string;
  title: string;
  price_cents: number;
  currency: string;
}

interface Props {
  items: RelatedProduct[];
}

const RelatedProducts: FC<Props> = ({ items }) => {
  if (!items?.length) return null;
  return (
    <section className="py-16 bg-white">
      <div className="container mx-auto px-4">
        <div className="text-center mb-12">
          <h2 className="font-playfair text-3xl font-bold text-maroon mb-4">You May Also Like</h2>
          <div className="w-24 h-1 bg-gold mx-auto" />
        </div>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
          {items.map((p) => (
            <a
              key={p.id}
              href={`/product/${p.slug}`}
              className="bg-white rounded-2xl shadow-lg overflow-hidden group hover:shadow-xl transition-all duration-300 transform hover:-translate-y-1"
            >
              <div className="p-6">
                <h3 className="font-playfair text-xl font-bold text-maroon mb-2 group-hover:text-gold transition-colors duration-300">
                  {p.title}
                </h3>
                <span className="text-2xl font-bold text-gold">
                  {p.currency} {(p.price_cents / 100).toFixed(2)}
                </span>
              </div>
            </a>
          ))}
        </div>
      </div>
    </section>
  );
};

export default RelatedProducts;
