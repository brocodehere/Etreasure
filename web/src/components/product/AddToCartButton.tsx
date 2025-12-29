import type { FC } from 'react';
import { addToCart as addToCartAPI } from '../../lib/api';

interface Props {
  id: string;
  name: string;
  price: number;
  image?: string;
}

const AddToCartButton: FC<Props> = ({ id, name, price, image }) => {
  const addToCart = async () => {
    try {
      // Backend will handle session creation automatically
      await addToCartAPI(id, 1);
      // Dispatch event to update cart count in header
      window.dispatchEvent(new Event('cart-updated'));
      // Show success message
      // Success is handled by cart count update in header
    } catch (error) {
      console.error('Failed to add to cart:', error);
      // Error is handled silently - user can see cart count doesn't update
    }
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
