import React, { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useParams, useNavigate, Link } from 'react-router-dom';

interface ContentPage {
  id: string;
  title: string;
  slug: string;
  content: string;
  type: string;
  is_active: boolean;
  meta_title: string;
  meta_description: string;
  created_at: string;
  updated_at: string;
}

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

const apiRequest = async (url: string, options: RequestInit = {}) => {
  const token = localStorage.getItem('access_token');
  const headers = {
    'Content-Type': 'application/json',
    ...(token && { Authorization: `Bearer ${token}` }),
    ...options.headers,
  };

  const response = await fetch(`${API_BASE_URL}${url}`, {
    ...options,
    headers,
  });

  if (!response.ok) {
    throw new Error(`API Error: ${response.statusText}`);
  }

  return response.json();
};

export default function ContentEdit() {
  const { id } = useParams<{ id?: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const isEditing = !!id;

  const { data: contentData, isLoading } = useQuery({
    queryKey: ['content-page', id],
    queryFn: async () => {
      if (!id) return null;
      const response = await apiRequest(`/api/admin/content/pages/${id}`);
      return response;
    },
    enabled: !!id,
  });

  const [formData, setFormData] = useState<Partial<ContentPage>>({
    title: '',
    slug: '',
    content: '',
    type: 'about',
    is_active: true,
    meta_title: '',
    meta_description: '',
  });

  useEffect(() => {
    if (contentData) {
      setFormData(contentData);
    }
  }, [contentData]);

  const updateMutation = useMutation({
    mutationFn: async (data: Partial<ContentPage>) => {
      const url = data.id ? `/api/admin/content/pages/${data.id}` : '/api/admin/content/pages';
      const method = data.id ? 'PUT' : 'POST';
      
      return apiRequest(url, {
        method,
        body: JSON.stringify(data),
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['content-pages'] });
      if (!isEditing) {
        navigate('/content');
      } else {
        navigate('/content');
      }
      alert('Content saved successfully!');
    },
    onError: (error: any) => {
      alert(`Error saving content: ${error.message}`);
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    updateMutation.mutate(formData);
  };

  if (isLoading && isEditing) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <div className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center py-4">
            <div>
              <h1 className="text-2xl font-bold text-gray-900">
                {isEditing ? 'Edit Content' : 'Create New Content'}
              </h1>
              <p className="text-sm text-gray-600 mt-1">
                {isEditing ? 'Modify existing content page' : 'Create a new content page'}
              </p>
            </div>
            <div className="flex items-center space-x-4">
              <Link
                to="/content"
                className="bg-gray-200 hover:bg-gray-300 text-gray-800 px-4 py-2 rounded-lg text-sm font-medium transition-colors duration-200"
              >
                Cancel
              </Link>
            </div>
          </div>
        </div>
      </div>

      {/* Form */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        <div className="bg-white rounded-lg shadow">
          <form onSubmit={handleSubmit} className="space-y-6">
            {/* Basic Information */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Title *</label>
                <input
                  type="text"
                  value={formData.title || ''}
                  onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  placeholder="Enter content title"
                  required
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Slug *</label>
                <div className="flex items-center space-x-2">
                  <input
                    type="text"
                    value={formData.slug || ''}
                    onChange={(e) => setFormData({ ...formData, slug: e.target.value })}
                    className="flex-1 px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    placeholder="url-friendly-slug"
                    required
                  />
                  <span className="text-sm text-gray-500">/</span>
                </div>
              </div>
            </div>

            {/* Type Selection */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">Content Type *</label>
              <select
                value={formData.type || 'about'}
                onChange={(e) => setFormData({ ...formData, type: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                required
              >
                <option value="about">About</option>
                <option value="our-story">Our Story</option>
                <option value="artisans">Artisans</option>
                <option value="crafts">Crafts & Techniques</option>
                <option value="policy">Policy</option>
                <option value="contact">Contact</option>
                <option value="faq">FAQ</option>
              </select>
            </div>

            {/* SEO Fields */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Meta Title</label>
                <input
                  type="text"
                  value={formData.meta_title || ''}
                  onChange={(e) => setFormData({ ...formData, meta_title: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  placeholder="SEO meta title"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Meta Description</label>
                <textarea
                  value={formData.meta_description || ''}
                  onChange={(e) => setFormData({ ...formData, meta_description: e.target.value })}
                  rows={3}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  placeholder="SEO meta description"
                />
              </div>
            </div>

            {/* Content Editor */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">Content *</label>
              <textarea
                value={formData.content || ''}
                onChange={(e) => setFormData({ ...formData, content: e.target.value })}
                rows={15}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                placeholder="Enter content (HTML supported)"
                required
              />
            </div>

            {/* Status */}
            <div>
              <label className="flex items-center space-x-3">
                <input
                  type="checkbox"
                  checked={formData.is_active || false}
                  onChange={(e) => setFormData({ ...formData, is_active: e.target.checked })}
                  className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                />
                <span className="ml-2 text-sm font-medium text-gray-700">Active</span>
              </label>
            </div>

            {/* Submit Buttons */}
            <div className="flex justify-end space-x-4">
              <Link
                to="/content"
                className="bg-gray-200 hover:bg-gray-300 text-gray-800 px-6 py-3 rounded-lg text-sm font-medium transition-colors duration-200"
              >
                Cancel
              </Link>
              <button
                type="submit"
                disabled={updateMutation.isPending}
                className="bg-blue-600 hover:bg-blue-700 disabled:bg-blue-300 text-white px-6 py-3 rounded-lg text-sm font-medium transition-colors duration-200"
              >
                {updateMutation.isPending ? 'Saving...' : (isEditing ? 'Update Content' : 'Create Content')}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}

