import type { FC } from 'react';

interface Props {
  id: number;
  name: string;
  price: number;
  image?: string;
}

const AddToCartButton: FC<Props> = ({ id, name, price, image }) => {
  const addToCart = () => {
    const raw = localStorage.getItem('cart');
    const cart = raw ? JSON.parse(raw) : [];
    const existing = cart.find((item: any) => item.id === id);
    if (existing) {
      existing.quantity += 1;
    } else {
      cart.push({ id, name, price, image, quantity: 1 });
    }
    localStorage.setItem('cart', JSON.stringify(cart));
    window.dispatchEvent(new StorageEvent('storage', { key: 'cart', newValue: JSON.stringify(cart) }));
  };

  return (
    <button
      type="button"
      onClick={addToCart}
      className="btn-primary w-full mt-4 rounded-lg py-3 text-center"
    >
      Add to Cart
    </button>
  );
};

export default AddToCartButton;
