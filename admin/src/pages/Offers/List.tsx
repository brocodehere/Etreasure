import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { api } from '../../lib/api';
import type { Offer } from '../../types';

interface ListResponse {
  items: Offer[];
  next_cursor?: string;
}

export function OffersListPage() {
  const [cursor, setCursor] = useState<string | undefined>();
  const [limit] = useState(50);

  // Toast message state
  const [showSuccess, setShowSuccess] = useState(false);
  const [successMessage, setSuccessMessage] = useState('');

  // Show success message function
  const showSuccessMessage = (message: string) => {
    setSuccessMessage(message);
    setShowSuccess(true);
    setTimeout(() => {
      setShowSuccess(false);
    }, 3000); // Hide after 3 seconds
  };

  const { data, isLoading, error } = useQuery<Offer[]>({
    queryKey: ['offers'],
    queryFn: () => {
      const url = cursor ? `/offers?cursor=${cursor}&limit=${limit}` : `/offers?limit=${limit}`;
      return api.get<ListResponse>(url).then(response => {
        return response.items;
      });
    },
    staleTime: 0, // Always fresh data
  });

  const queryClient = useQueryClient();

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.delete(`/offers/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['offers'] });
      showSuccessMessage('Offer deleted successfully!');
    },
  });

  const handleDelete = (id: string) => {
    if (confirm('Are you sure you want to delete this offer?')) {
      deleteMutation.mutate(id);
    }
  };

  const loadMore = () => {
    // For now, just refetch - pagination can be implemented later
    queryClient.invalidateQueries({ queryKey: ['offers'] });
  };

  if (isLoading && !data) return <div>Loading...</div>;
  if (error) return <div>Error: {(error as Error).message}</div>;

  return (
    <div className="p-6">
      {/* Success Message */}
      {showSuccess && (
        <div className="fixed top-4 right-4 z-50 animate-pulse">
          <div className="bg-green-500 text-white px-6 py-3 rounded-lg shadow-lg flex items-center space-x-2">
            <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
            </svg>
            <span className="font-medium">{successMessage}</span>
          </div>
        </div>
      )}

      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">Offers</h1>
        <Link
          to="new"
          className="bg-gold text-white px-4 py-2 rounded hover:bg-yellow-600 transition"
        >
          New Offer
        </Link>
      </div>

      {data && data.length > 0 ? (
        <>
          <ul className="space-y-4">
            {data.map((offer: Offer) => (
              <li key={offer.id} className="border rounded-lg p-4 bg-white shadow-sm">
                <div className="flex justify-between items-start">
                  <div>
                    <h3 className="font-semibold text-lg">{offer.title}</h3>
                    {offer.description && <p className="text-gray-600 text-sm mt-1">{offer.description}</p>}
                    <div className="mt-2 flex items-center gap-4 text-sm text-gray-600">
                      <span>
                        Discount: {offer.discount_type === 'percentage' ? `${offer.discount_value}%` : `$${offer.discount_value}`}
                      </span>
                      <span>Applies to: {offer.applies_to}</span>
                      <span className={offer.is_active ? 'text-green-600' : 'text-red-600'}>
                        {offer.is_active ? 'Active' : 'Inactive'}
                      </span>
                      {offer.usage_limit && (
                        <span>
                          Usage: {offer.usage_count}/{offer.usage_limit}
                        </span>
                      )}
                    </div>
                    <div className="mt-1 text-xs text-gray-500">
                      Starts: {new Date(offer.starts_at).toLocaleString()}
                      {' • '}
                      Ends: {new Date(offer.ends_at).toLocaleString()}
                    </div>
                  </div>
                  <div className="flex gap-2">
                    <Link
                      to={offer.id}
                      className="text-gold hover:underline text-sm font-medium"
                    >
                      Edit
                    </Link>
                    <button
                      onClick={() => handleDelete(offer.id)}
                      className="text-red-600 hover:underline text-sm font-medium"
                      disabled={deleteMutation.isPending}
                    >
                      {deleteMutation.isPending ? 'Deleting...' : 'Delete'}
                    </button>
                  </div>
                </div>
              </li>
            ))}
          </ul>
        </>
      ) : (
        <div className="text-center py-12 text-gray-500">
          No offers yet. <Link to="new" className="text-gold hover:underline">Create the first offer</Link>.
        </div>
      )}
    </div>
  );
}
