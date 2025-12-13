import React, { useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../../lib/api';
import type { User } from '../../types';

interface FormData {
  email: string;
  first_name: string;
  last_name: string;
  password: string;
  is_active: boolean;
  roles: string[];
}

export function UserEditPage() {
  const { id } = useParams<{ id?: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const { data: user, isLoading, error } = useQuery<User>({
    queryKey: ['user', id],
    queryFn: () => api.get<User>(`/users/${id}`).then((r: any) => r.data),
    enabled: !!id,
  });

  const { data: rolesData } = useQuery<{ data: { id: string; name: string }[] }>({
    queryKey: ['roles'],
    queryFn: () => api.get('/roles').then((r: any) => r.data),
  });

  const createMutation = useMutation({
    mutationFn: (payload: Omit<User, 'id' | 'created_at' | 'updated_at'>) =>
      api.post<User>('/users', payload).then((r: any) => r.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      navigate('/users');
    },
  });

  const updateMutation = useMutation({
    mutationFn: (payload: Partial<User>) =>
      api.put<User>(`/users/${id}`, payload).then((r: any) => r.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      navigate('/users');
    },
  });

  const [formData, setFormData] = React.useState<FormData>({
    email: '',
    first_name: '',
    last_name: '',
    password: '',
    is_active: true,
    roles: [],
  });

  useEffect(() => {
    if (user) {
      setFormData({
        email: user.email,
        first_name: user.first_name || '',
        last_name: user.last_name || '',
        password: '',
        is_active: user.is_active,
        roles: user.roles,
      });
    }
  }, [user]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value, type } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: type === 'checkbox' ? (e.target as HTMLInputElement).checked : value,
    }));
  };

  const handleRoleChange = (roleName: string, checked: boolean) => {
    setFormData(prev => ({
      ...prev,
      roles: checked
        ? [...prev.roles, roleName]
        : prev.roles.filter(r => r !== roleName),
    }));
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const payload = {
      email: formData.email,
      first_name: formData.first_name || null,
      last_name: formData.last_name || null,
      ...(formData.password ? { password: formData.password } : {}),
      is_active: formData.is_active,
      roles: formData.roles,
    };
    if (id) {
      updateMutation.mutate(payload);
    } else {
      createMutation.mutate(payload);
    }
  };

  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>Error: {(error as Error).message}</div>;

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold mb-6">{id ? 'Edit User' : 'New User'}</h1>

      <form onSubmit={handleSubmit} className="space-y-4 max-w-xl">
        <div>
          <label className="block text-sm font-medium mb-1">Email</label>
          <input
            name="email"
            type="email"
            value={formData.email}
            onChange={handleChange}
            required
            className="w-full border rounded px-3 py-2"
          />
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium mb-1">First Name</label>
            <input
              name="first_name"
              type="text"
              value={formData.first_name}
              onChange={handleChange}
              className="w-full border rounded px-3 py-2"
            />
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">Last Name</label>
            <input
              name="last_name"
              type="text"
              value={formData.last_name}
              onChange={handleChange}
              className="w-full border rounded px-3 py-2"
            />
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium mb-1">
            Password {id ? '(leave blank to keep current)' : ''}
          </label>
          <input
            name="password"
            type="password"
            value={formData.password}
            onChange={handleChange}
            minLength={8}
            required={!id}
            className="w-full border rounded px-3 py-2"
          />
        </div>

        <div className="flex items-center gap-2">
          <input
            name="is_active"
            type="checkbox"
            checked={formData.is_active}
            onChange={handleChange}
            className="rounded"
          />
          <label htmlFor="is_active" className="text-sm font-medium">Active</label>
        </div>

        <div>
          <h3 className="text-lg font-semibold mb-2">Roles</h3>
          {rolesData?.data ? (
            <div className="space-y-2">
              {rolesData.data.map((role) => (
                <label key={role.id} className="flex items-center gap-2">
                  <input
                    type="checkbox"
                    checked={formData.roles.includes(role.name)}
                    onChange={(e) => handleRoleChange(role.name, e.target.checked)}
                    className="rounded"
                  />
                  <span className="text-sm">{role.name}</span>
                </label>
              ))}
            </div>
          ) : (
            <div className="text-sm text-gray-500">Loading roles...</div>
          )}
        </div>

        <div className="flex gap-2">
          <button
            type="submit"
            disabled={createMutation.isPending || updateMutation.isPending}
            className="bg-gold text-white px-4 py-2 rounded hover:bg-yellow-600 transition disabled:opacity-50"
          >
            {createMutation.isPending || updateMutation.isPending ? 'Saving...' : id ? 'Update' : 'Create'}
          </button>
          <button
            type="button"
            onClick={() => navigate('/users')}
            className="border border-gray-300 px-4 py-2 rounded hover:bg-gray-50 transition"
          >
            Cancel
          </button>
        </div>
      </form>
    </div>
  );
}
