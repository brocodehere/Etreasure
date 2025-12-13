import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { api } from '../../lib/api';
import type { Customer } from '../../types';

interface ListResponse {
  data: Customer[];
  next_cursor?: string;
}

export function CustomersListPage() {
  const [cursor, setCursor] = useState<string | undefined>();
  const [limit] = useState(50);

  const { data, isLoading, error } = useQuery<ListResponse>({
    queryKey: ['customers', cursor, limit],
    queryFn: () =>
      api.get<ListResponse>('/customers', { params: { cursor, limit } }).then(r => r.data),
  });

  const queryClient = useQueryClient();

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.delete(`/customers/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['customers'] });
    },
  });

  const handleDelete = (id: string) => {
    if (confirm('Are you sure you want to delete this customer?')) {
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
        <h1 className="text-2xl font-bold">Customers</h1>
        <Link
          to="new"
          className="bg-gold text-white px-4 py-2 rounded hover:bg-yellow-600 transition"
        >
          New Customer
        </Link>
      </div>

      {data?.data && data.data.length > 0 ? (
        <>
          <ul className="space-y-4">
            {data.data.map((customer) => (
              <li key={customer.id} className="border rounded-lg p-4 bg-white shadow-sm">
                <div className="flex justify-between items-start">
                  <div>
                    <h3 className="font-semibold text-lg">
                      {customer.first_name && customer.last_name
                        ? `${customer.first_name} ${customer.last_name}`
                        : customer.email}
                    </h3>
                    <p className="text-gray-600 text-sm mt-1">{customer.email}</p>
                    {customer.phone && <p className="text-gray-500 text-sm">{customer.phone}</p>}
                    <div className="mt-2 text-sm text-gray-600">
                      {customer.addresses.length} address{customer.addresses.length !== 1 ? 'es' : ''}
                    </div>
                    <div className="mt-1 text-xs text-gray-500">
                      Joined {new Date(customer.created_at).toLocaleDateString()}
                    </div>
                  </div>
                  <div className="flex gap-2">
                    <Link
                      to={customer.id}
                      className="text-gold hover:underline text-sm font-medium"
                    >
                      Edit
                    </Link>
                    <button
                      onClick={() => handleDelete(customer.id)}
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
          No customers yet. <Link to="new" className="text-gold hover:underline">Create the first customer</Link>.
        </div>
      )}
    </div>
  );
}
