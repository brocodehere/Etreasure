import React from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { api } from '../../lib/api';
import type { Setting } from '../../types';

interface ListResponse {
  data: Setting[];
}

export function SettingsListPage() {
  const { data, isLoading, error } = useQuery<ListResponse>({
    queryKey: ['settings'],
    queryFn: () => api.get<ListResponse>('/settings').then(r => r.data),
  });

  const queryClient = useQueryClient();

  const deleteMutation = useMutation({
    mutationFn: (key: string) => api.delete(`/settings/${key}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['settings'] });
    },
  });

  const handleDelete = (key: string) => {
    if (confirm(`Are you sure you want to delete setting "${key}"?`)) {
      deleteMutation.mutate(key);
    }
  };

  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>Error: {(error as Error).message}</div>;

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">Settings</h1>
        <Link
          to="new"
          className="bg-gold text-white px-4 py-2 rounded hover:bg-yellow-600 transition"
        >
          New Setting
        </Link>
      </div>

      {data?.data && data.data.length > 0 ? (
        <ul className="space-y-4">
          {data.data.map((setting) => (
            <li key={setting.key} className="border rounded-lg p-4 bg-white shadow-sm">
              <div className="flex justify-between items-start">
                <div>
                  <h3 className="font-semibold text-lg">{setting.key}</h3>
                  <p className="text-gray-600 text-sm mt-1">
                    Type: <span className="font-mono">{setting.type}</span>
                  </p>
                  <p className="text-gray-700 text-sm mt-1 break-all">
                    Value: <span className="font-mono">{setting.value}</span>
                  </p>
                  {setting.description && (
                    <p className="text-gray-500 text-sm mt-1">{setting.description}</p>
                  )}
                  <div className="mt-1 text-xs text-gray-500">
                    Updated {new Date(setting.updated_at).toLocaleString()}
                  </div>
                </div>
                <div className="flex gap-2">
                  <Link
                    to={setting.key}
                    className="text-gold hover:underline text-sm font-medium"
                  >
                    Edit
                  </Link>
                  <button
                    onClick={() => handleDelete(setting.key)}
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
      ) : (
        <div className="text-center py-12 text-gray-500">
          No settings yet. <Link to="new" className="text-gold hover:underline">Create the first setting</Link>.
        </div>
      )}
    </div>
  );
}
