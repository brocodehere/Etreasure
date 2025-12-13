import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { api } from '../../lib/api';
import type { User } from '../../types';

interface ListResponse {
  data: User[];
  next_cursor?: string;
}

export function UsersListPage() {
  const [cursor, setCursor] = useState<string | undefined>();
  const [limit] = useState(50);

  const { data, isLoading, error } = useQuery<ListResponse>({
    queryKey: ['users', cursor, limit],
    queryFn: () => {
      const url = cursor ? `/users?cursor=${cursor}&limit=${limit}` : `/users?limit=${limit}`;
      return api.get<ListResponse>(url).then((r: any) => r.data);
    },
  });

  const queryClient = useQueryClient();

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.delete(`/users/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
    },
  });

  const handleDelete = (id: string) => {
    if (confirm('Are you sure you want to delete this user?')) {
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
        <h1 className="text-2xl font-bold">Users & Roles</h1>
        <Link
          to="new"
          className="bg-gold text-white px-4 py-2 rounded hover:bg-yellow-600 transition"
        >
          New User
        </Link>
      </div>

      {data?.data && data.data.length > 0 ? (
        <>
          <ul className="space-y-4">
            {data.data.map((user: User) => (
              <li key={user.id} className="border rounded-lg p-4 bg-white shadow-sm">
                <div className="flex justify-between items-start">
                  <div>
                    <h3 className="font-semibold text-lg">
                      {user.first_name && user.last_name
                        ? `${user.first_name} ${user.last_name}`
                        : user.email}
                    </h3>
                    <p className="text-gray-600 text-sm mt-1">{user.email}</p>
                    <div className="mt-2 flex items-center gap-4 text-sm text-gray-600">
                      <span>Status: <span className={user.is_active ? 'text-green-600' : 'text-red-600'}>{user.is_active ? 'Active' : 'Inactive'}</span></span>
                      <span>Roles: {user.roles.length > 0 ? user.roles.join(', ') : 'None'}</span>
                    </div>
                    <div className="mt-1 text-xs text-gray-500">
                      Joined {new Date(user.created_at).toLocaleDateString()}
                    </div>
                  </div>
                  <div className="flex gap-2">
                    <Link
                      to={user.id}
                      className="text-gold hover:underline text-sm font-medium"
                    >
                      Edit
                    </Link>
                    <button
                      onClick={() => handleDelete(user.id)}
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
          No users yet. <Link to="new" className="text-gold hover:underline">Create the first user</Link>.
        </div>
      )}
    </div>
  );
}
