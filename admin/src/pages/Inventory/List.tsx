import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { api } from '../../lib/api';
import type { InventoryItem } from '../../types';

interface ListResponse {
  data: InventoryItem[];
  next_cursor?: string;
}

export function InventoryListPage() {
  const [cursor, setCursor] = useState<string | undefined>();
  const [limit] = useState(50);

  const { data, isLoading, error } = useQuery<ListResponse>({
    queryKey: ['inventory', cursor, limit],
    queryFn: () =>
      api.get<ListResponse>('/inventory', { params: { cursor, limit } }).then(r => r.data),
  });

  const queryClient = useQueryClient();

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.delete(`/inventory/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['inventory'] });
    },
  });

  const handleDelete = (id: string) => {
    if (confirm('Are you sure you want to delete this inventory item?')) {
      deleteMutation.mutate(id);
    }
  };

  const loadMore = () => {
    if (data?.next_cursor) {
      setCursor(data.next_cursor);
    }
  };

  if (isLoading && !data) return <div>Loading...</div>;
  if (error) return <div>Error: {(error as Error).message}</div>;

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">Inventory</h1>
        <Link
          to="new"
          className="bg-gold text-white px-4 py-2 rounded hover:bg-yellow-600 transition"
        >
          New Item
        </Link>
      </div>

      {data?.data && data.data.length > 0 ? (
        <>
          <ul className="space-y-4">
            {data.data.map((item) => (
              <li key={item.id} className="border rounded-lg p-4 bg-white shadow-sm">
                <div className="flex justify-between items-start">
                  <div>
                    <h3 className="font-semibold text-lg">{item.sku}</h3>
                    <p className="text-gray-600 text-sm mt-1">
                      Product ID: {item.product_id}
                      {item.variant_id && ` • Variant ID: ${item.variant_id}`}
                    </p>
                    <div className="mt-2 grid grid-cols-3 gap-4 text-sm text-gray-600">
                      <span>Quantity: <span className={item.quantity > 0 ? 'text-green-600' : 'text-red-600'}>{item.quantity}</span></span>
                      <span>Reserved: <span className="text-yellow-600">{item.reserved}</span></span>
                      <span>Available: <span className={item.available > 0 ? 'text-green-600' : 'text-red-600'}>{item.available}</span></span>
                    </div>
                    {item.location && <p className="text-gray-500 text-sm mt-1">Location: {item.location}</p>}
                    {item.cost_price && <p className="text-gray-500 text-sm">Cost: ${item.cost_price.toFixed(2)}</p>}
                    <div className="mt-1 text-xs text-gray-500">
                      Updated {new Date(item.updated_at).toLocaleString()}
                    </div>
                  </div>
                  <div className="flex gap-2">
                    <Link
                      to={item.id}
                      className="text-gold hover:underline text-sm font-medium"
                    >
                      Edit
                    </Link>
                    <button
                      onClick={() => handleDelete(item.id)}
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

          {data.next_cursor && (
            <div className="mt-6 text-center">
              <button
                onClick={loadMore}
                disabled={isLoading}
                className="bg-maroon text-white px-4 py-2 rounded hover:bg-red-900 transition disabled:opacity-50"
              >
                {isLoading ? 'Loading...' : 'Load More'}
              </button>
            </div>
          )}
        </>
      ) : (
        <div className="text-center py-12 text-gray-500">
          No inventory items yet. <Link to="new" className="text-gold hover:underline">Create the first item</Link>.
        </div>
      )}
    </div>
  );
}
