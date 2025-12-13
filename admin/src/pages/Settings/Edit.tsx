import React, { useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../../lib/api';
import type { Setting } from '../../types';

interface FormData {
  key: string;
  value: string;
  type: 'string' | 'number' | 'boolean' | 'json';
  description: string;
}

export function SettingEditPage() {
  const { key: keyParam } = useParams<{ key?: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const { data: setting, isLoading, error } = useQuery<Setting>({
    queryKey: ['setting', keyParam],
    queryFn: () => api.get<Setting>(`/settings/${keyParam}`).then(r => r.data),
    enabled: !!keyParam,
  });

  const createMutation = useMutation({
    mutationFn: (payload: Omit<Setting, 'updated_at'>) =>
      api.post<Setting>('/settings', payload).then(r => r.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['settings'] });
      navigate('/settings');
    },
  });

  const updateMutation = useMutation({
    mutationFn: (payload: Partial<Setting>) =>
      api.put<Setting>(`/settings/${keyParam}`, payload).then(r => r.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['settings'] });
      navigate('/settings');
    },
  });

  const [formData, setFormData] = React.useState<FormData>({
    key: '',
    value: '',
    type: 'string',
    description: '',
  });

  useEffect(() => {
    if (setting) {
      setFormData({
        key: setting.key,
        value: setting.value,
        type: setting.type as 'string' | 'number' | 'boolean' | 'json',
        description: setting.description || '',
      });
    }
  }, [setting]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    const { name, value, type } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: type === 'checkbox' ? (e.target as HTMLInputElement).checked : value,
    }));
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const payload = {
      key: formData.key,
      value: formData.value,
      type: formData.type,
      description: formData.description || null,
    };
    if (keyParam) {
      updateMutation.mutate(payload);
    } else {
      createMutation.mutate(payload);
    }
  };

  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>Error: {(error as Error).message}</div>;

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold mb-6">{keyParam ? 'Edit Setting' : 'New Setting'}</h1>

      <form onSubmit={handleSubmit} className="space-y-4 max-w-xl">
        <div>
          <label className="block text-sm font-medium mb-1">Key</label>
          <input
            name="key"
            type="text"
            value={formData.key}
            onChange={handleChange}
            required
            disabled={!!keyParam}
            className="w-full border rounded px-3 py-2 disabled:bg-gray-100"
          />
          {keyParam && <p className="text-xs text-gray-500 mt-1">Key cannot be changed after creation.</p>}
        </div>

        <div>
          <label className="block text-sm font-medium mb-1">Type</label>
          <select
            name="type"
            value={formData.type}
            onChange={handleChange}
            className="w-full border rounded px-3 py-2"
          >
            <option value="string">String</option>
            <option value="number">Number</option>
            <option value="boolean">Boolean</option>
            <option value="json">JSON</option>
          </select>
        </div>

        <div>
          <label className="block text-sm font-medium mb-1">Value</label>
          <textarea
            name="value"
            value={formData.value}
            onChange={handleChange}
            required
            rows={4}
            className="w-full border rounded px-3 py-2 font-mono text-sm"
            placeholder={
              formData.type === 'boolean' ? 'e.g., true or false' :
              formData.type === 'number' ? 'e.g., 123.45' :
              formData.type === 'json' ? 'e.g., {"key":"value"}' :
              'Plain text value'
            }
          />
        </div>

        <div>
          <label className="block text-sm font-medium mb-1">Description (optional)</label>
          <textarea
            name="description"
            value={formData.description}
            onChange={handleChange}
            rows={2}
            className="w-full border rounded px-3 py-2"
            placeholder="What this setting controls"
          />
        </div>

        <div className="flex gap-2">
          <button
            type="submit"
            disabled={createMutation.isPending || updateMutation.isPending}
            className="bg-gold text-white px-4 py-2 rounded hover:bg-yellow-600 transition disabled:opacity-50"
          >
            {createMutation.isPending || updateMutation.isPending ? 'Saving...' : keyParam ? 'Update' : 'Create'}
          </button>
          <button
            type="button"
            onClick={() => navigate('/settings')}
            className="border border-gray-300 px-4 py-2 rounded hover:bg-gray-50 transition"
          >
            Cancel
          </button>
        </div>
      </form>
    </div>
  );
}
