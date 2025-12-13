import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { api } from '../../lib/api';
import type { Banner } from '../../types';

interface ListResponse {
  data: Banner[];
  next_cursor?: string;
}

export function BannersListPage() {
  const [cursor, setCursor] = useState<string | undefined>();
  const [limit] = useState(50);

  const { data, isLoading, error, refetch } = useQuery<ListResponse>({
    queryKey: ['banners'],
    queryFn: () => {
      return api.get<ListResponse>('/banners').then(response => {
        return response;
      });
    },
    staleTime: 0, // Always fresh data
  });

  const queryClient = useQueryClient();

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.delete(`/banners/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['banners'] });
    },
  });

  const handleDelete = (id: string) => {
    if (confirm('Are you sure you want to delete this banner?')) {
      deleteMutation.mutate(id);
    }
  };

  const handleRefresh = () => {
    refetch();
  };

  const loadMore = () => {
    // For now, just refetch - pagination can be implemented later
    refetch();
  };

  if (isLoading && !data) return <div>Loading...</div>;
  if (error) return <div>Error: {(error as Error).message}</div>;


  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">Banners</h1>
        <div className="flex gap-2">
          <button
            onClick={handleRefresh}
            className="border border-gray-300 text-gray-700 px-4 py-2 rounded hover:bg-gray-50 transition"
          >
            Refresh
          </button>
          <Link
            to="new"
            className="bg-gold text-white px-4 py-2 rounded hover:bg-yellow-600 transition"
          >
            New Banner
          </Link>
        </div>
      </div>

      {data && data.data && data.data.length > 0 ? (
        <>
          <ul className="space-y-4">
            {data.data.map((banner: Banner) => (
              <li key={banner.id} className="border rounded-lg p-4 bg-white shadow-sm">
                <div className="flex justify-between items-start">
                  <div>
                    <h3 className="font-semibold text-lg">{banner.title}</h3>
                    {banner.link_url && (
                      <a
                        href={banner.link_url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-blue-600 hover:underline text-sm"
                      >
                        {banner.link_url}
                      </a>
                    )}
                    <div className="mt-2 flex items-center gap-4 text-sm text-gray-600">
                      <span>Sort: {banner.sort_order}</span>
                      <span className={banner.is_active ? 'text-green-600' : 'text-red-600'}>
                        {banner.is_active ? 'Active' : 'Inactive'}
                      </span>
                      {banner.starts_at && <span>Starts: {new Date(banner.starts_at).toLocaleDateString()}</span>}
                      {banner.ends_at && <span>Ends: {new Date(banner.ends_at).toLocaleDateString()}</span>}
                    </div>
                    {banner.image_url && (
                      <img
                        src={banner.image_url}
                        alt={banner.title}
                        className="mt-3 h-20 w-auto rounded border"
                      />
                    )}
                  </div>
                  <div className="flex gap-2">
                    <Link
                      to={banner.id}
                      className="text-gold hover:underline text-sm font-medium"
                    >
                      Edit
                    </Link>
                    <button
                      onClick={() => handleDelete(banner.id)}
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

          {/* Pagination can be implemented later */}
        </>
      ) : (
        <div className="text-center py-12 text-gray-500">
          No banners yet. <Link to="new" className="text-gold hover:underline">Create the first banner</Link>.
        </div>
      )}
    </div>
  );
}
