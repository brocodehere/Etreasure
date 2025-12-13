import React, { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../../lib/api';
import { ImageUploader } from '../../components/ImageUploader';
import type { Banner } from '../../types';

interface FormData {
  title: string;
  image_url: string;
  link_url: string;
  is_active: boolean;
  sort_order: number;
  starts_at: string;
  ends_at: string;
}

export function BannerEditPage() {
  const { id } = useParams<{ id?: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  
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

  const { data: banner, isLoading, error } = useQuery<Banner>({
    queryKey: ['banner', id],
    queryFn: () => api.get<Banner>(`/banners/${id}`),
    enabled: !!id,
  });

  const createMutation = useMutation({
    mutationFn: (payload: {
      title: string;
      image_url: string;
      link_url?: string | null;
      is_active: boolean;
      sort_order: number;
      starts_at?: string | null;
      ends_at?: string | null;
    }) => api.post<Banner>('/banners', payload).catch(err => {
      console.error('Banner creation error:', err);
      throw err;
    }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['banners'] });
      queryClient.refetchQueries({ queryKey: ['banners'] });
      // Show success message first, then navigate after delay
      showSuccessMessage('Banner created successfully!');
      setTimeout(() => {
        navigate('/banners');
      }, 2000); // Navigate after 2 seconds
    },
  });

  const updateMutation = useMutation({
    mutationFn: (payload: Partial<Banner>) =>
      api.put<Banner>(`/banners/${id}`, payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['banners'] });
      queryClient.refetchQueries({ queryKey: ['banners'] });
      // Show success message for banner updates
      showSuccessMessage('Banner updated successfully!');
      navigate('/banners');
    },
  });

  const [formData, setFormData] = React.useState<FormData>({
    title: '',
    image_url: '',
    link_url: '',
    is_active: true,
    sort_order: 0,
    starts_at: '',
    ends_at: '',
  });

  useEffect(() => {
    if (banner) {
      setFormData({
        title: banner.title,
        image_url: banner.image_url,
        link_url: banner.link_url || '',
        is_active: banner.is_active,
        sort_order: banner.sort_order,
        starts_at: banner.starts_at ? new Date(banner.starts_at).toISOString().slice(0, 16) : '',
        ends_at: banner.ends_at ? new Date(banner.ends_at).toISOString().slice(0, 16) : '',
      });
    }
  }, [banner]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    const { name, value, type } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: type === 'checkbox' ? (e.target as HTMLInputElement).checked : value,
    }));
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    
    // Validate required fields
    if (!formData.title.trim()) {
      alert('Title is required');
      return;
    }
    // Image URL is now optional - can be set via upload or manual input
    
    const payload = {
      title: formData.title.trim(),
      image_url: formData.image_url.trim(),
      link_url: formData.link_url.trim() || undefined,
      is_active: formData.is_active,
      sort_order: Number(formData.sort_order) || 0,
      starts_at: formData.starts_at ? new Date(formData.starts_at).toISOString() : undefined,
      ends_at: formData.ends_at ? new Date(formData.ends_at).toISOString() : undefined,
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

      <h1 className="text-2xl font-bold mb-6">{id ? 'Edit Banner' : 'New Banner'}</h1>

      <form onSubmit={handleSubmit} className="space-y-4 max-w-xl">
        <div>
          <label className="block text-sm font-medium mb-1">Title</label>
          <input
            name="title"
            type="text"
            value={formData.title}
            onChange={handleChange}
            required
            className="w-full border rounded px-3 py-2"
          />
        </div>

        <div>
          <label className="block text-sm font-medium mb-1">Banner Image</label>
          <ImageUploader 
            type="banner"
            onUploaded={(result) => {
              setFormData(prev => ({ ...prev, image_url: result.url }));
            }}
          />
          {formData.image_url && (
            <div className="mt-2">
              <img 
                src={formData.image_url} 
                alt="Banner preview" 
                className="h-20 w-auto border rounded"
                onError={(e) => {
                  const target = e.target as HTMLImageElement;
                  target.style.display = 'none';
                }}
              />
            </div>
          )}
        </div>

        <div>
          <label className="block text-sm font-medium mb-1">Or enter Image URL manually</label>
          <input
            name="image_url"
            type="url"
            value={formData.image_url}
            onChange={handleChange}
            placeholder="https://example.com/image.jpg"
            className="w-full border rounded px-3 py-2"
          />
        </div>

        <div>
          <label className="block text-sm font-medium mb-1">Link URL (optional)</label>
          <input
            name="link_url"
            type="url"
            value={formData.link_url}
            onChange={handleChange}
            placeholder="https://example.com"
            className="w-full border rounded px-3 py-2"
          />
        </div>

        <div className="flex items-center gap-4">
          <label className="flex items-center gap-2">
            <input
              name="is_active"
              type="checkbox"
              checked={formData.is_active}
              onChange={handleChange}
              className="rounded"
            />
            <span className="text-sm font-medium">Active</span>
          </label>
        </div>

        <div>
          <label className="block text-sm font-medium mb-1">Sort Order</label>
          <input
            name="sort_order"
            type="number"
            value={formData.sort_order}
            onChange={handleChange}
            className="w-full border rounded px-3 py-2"
          />
        </div>

        <div>
          <label className="block text-sm font-medium mb-1">Starts At (optional)</label>
          <input
            name="starts_at"
            type="datetime-local"
            value={formData.starts_at}
            onChange={handleChange}
            className="w-full border rounded px-3 py-2"
          />
        </div>

        <div>
          <label className="block text-sm font-medium mb-1">Ends At (optional)</label>
          <input
            name="ends_at"
            type="datetime-local"
            value={formData.ends_at}
            onChange={handleChange}
            className="w-full border rounded px-3 py-2"
          />
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
            onClick={() => navigate('/banners')}
            className="border border-gray-300 px-4 py-2 rounded hover:bg-gray-50 transition"
          >
            Cancel
          </button>
        </div>
      </form>
    </div>
  );
}
