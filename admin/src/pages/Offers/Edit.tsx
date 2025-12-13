import React, { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../../lib/api';
import type { Offer } from '../../types';

interface FormData {
  title: string;
  description: string;
  discount_type: 'percentage' | 'fixed';
  discount_value: number;
  applies_to: 'all' | 'products' | 'categories' | 'collections';
  applies_to_ids: string[];
  min_order_amount: string;
  usage_limit: string;
  is_active: boolean;
  starts_at: string;
  ends_at: string;
}

export function OfferEditPage() {
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

  const { data: offer, isLoading, error } = useQuery<Offer>({
    queryKey: ['offer', id],
    queryFn: () => api.get<Offer>(`/offers/${id}`),
    enabled: !!id,
  });

  const createMutation = useMutation({
    mutationFn: (payload: Omit<Offer, 'id' | 'created_at' | 'updated_at' | 'usage_count'>) =>
      api.post<Offer>('/offers', payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['offers'] });
      showSuccessMessage('Offer created successfully!');
      navigate('/offers');
    },
  });

  const updateMutation = useMutation({
    mutationFn: (payload: Partial<Offer>) =>
      api.put<Offer>(`/offers/${id}`, payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['offers'] });
      showSuccessMessage('Offer updated successfully!');
      navigate('/offers');
    },
  });

  const [formData, setFormData] = React.useState<FormData>({
    title: '',
    description: '',
    discount_type: 'percentage',
    discount_value: 0,
    applies_to: 'all',
    applies_to_ids: [],
    min_order_amount: '',
    usage_limit: '',
    is_active: true,
    starts_at: '',
    ends_at: '',
  });

  useEffect(() => {
    if (offer) {
      setFormData({
        title: offer.title,
        description: offer.description || '',
        discount_type: offer.discount_type as 'percentage' | 'fixed',
        discount_value: offer.discount_value,
        applies_to: offer.applies_to as 'all' | 'products' | 'categories' | 'collections',
        applies_to_ids: offer.applies_to_ids,
        min_order_amount: offer.min_order_amount?.toString() || '',
        usage_limit: offer.usage_limit?.toString() || '',
        is_active: offer.is_active,
        starts_at: new Date(offer.starts_at).toISOString().slice(0, 16),
        ends_at: new Date(offer.ends_at).toISOString().slice(0, 16),
      });
    }
  }, [offer]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    const { name, value, type } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: type === 'checkbox' 
        ? (e.target as HTMLInputElement).checked 
        : name === 'discount_value' 
          ? parseFloat(value) || 0 
          : value,
    }));
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const payload = {
      title: formData.title,
      description: formData.description || undefined,
      discount_type: formData.discount_type,
      discount_value: formData.discount_value,
      applies_to: formData.applies_to,
      applies_to_ids: formData.applies_to_ids,
      min_order_amount: formData.min_order_amount ? parseFloat(formData.min_order_amount) || undefined : undefined,
      usage_limit: formData.usage_limit ? parseInt(formData.usage_limit, 10) || undefined : undefined,
      is_active: formData.is_active,
      starts_at: new Date(formData.starts_at).toISOString(),
      ends_at: new Date(formData.ends_at).toISOString(),
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

      <h1 className="text-2xl font-bold mb-6">{id ? 'Edit Offer' : 'New Offer'}</h1>

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
          <label className="block text-sm font-medium mb-1">Description</label>
          <textarea
            name="description"
            value={formData.description}
            onChange={handleChange}
            rows={3}
            className="w-full border rounded px-3 py-2"
          />
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium mb-1">Discount Type</label>
            <select
              name="discount_type"
              value={formData.discount_type}
              onChange={handleChange}
              className="w-full border rounded px-3 py-2"
            >
              <option value="percentage">Percentage</option>
              <option value="fixed">Fixed Amount</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium mb-1">
              {formData.discount_type === 'percentage' ? 'Percentage (%)' : 'Amount ($)'}
            </label>
            <input
              name="discount_value"
              type="number"
              step="any"
              min="0"
              value={formData.discount_value}
              onChange={handleChange}
              required
              className="w-full border rounded px-3 py-2"
            />
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium mb-1">Applies To</label>
          <select
            name="applies_to"
            value={formData.applies_to}
            onChange={handleChange}
            className="w-full border rounded px-3 py-2"
          >
            <option value="all">All Products</option>
            <option value="products">Specific Products</option>
            <option value="categories">Specific Categories</option>
            <option value="collections">Specific Collections</option>
          </select>
        </div>

        {formData.applies_to !== 'all' && (
          <div>
            <label className="block text-sm font-medium mb-1">
              {formData.applies_to === 'products' ? 'Product IDs' : formData.applies_to === 'categories' ? 'Category IDs' : 'Collection IDs'}
            </label>
            <input
              name="applies_to_ids"
              type="text"
              value={formData.applies_to_ids.join(',')}
              onChange={(e) => setFormData(prev => ({ ...prev, applies_to_ids: e.target.value.split(',').map(s => s.trim()).filter(Boolean) }))}
              placeholder="uuid1,uuid2,uuid3"
              className="w-full border rounded px-3 py-2"
            />
          </div>
        )}

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium mb-1">Minimum Order Amount (optional)</label>
            <input
              name="min_order_amount"
              type="number"
              step="any"
              min="0"
              value={formData.min_order_amount}
              onChange={handleChange}
              placeholder="0.00"
              className="w-full border rounded px-3 py-2"
            />
          </div>

          <div>
            <label className="block text-sm font-medium mb-1">Usage Limit (optional)</label>
            <input
              name="usage_limit"
              type="number"
              min="0"
              value={formData.usage_limit}
              onChange={handleChange}
              placeholder="No limit"
              className="w-full border rounded px-3 py-2"
            />
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium mb-1">Starts At</label>
            <input
              name="starts_at"
              type="datetime-local"
              value={formData.starts_at}
              onChange={handleChange}
              required
              className="w-full border rounded px-3 py-2"
            />
          </div>

          <div>
            <label className="block text-sm font-medium mb-1">Ends At</label>
            <input
              name="ends_at"
              type="datetime-local"
              value={formData.ends_at}
              onChange={handleChange}
              required
              className="w-full border rounded px-3 py-2"
            />
          </div>
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
            onClick={() => navigate('/offers')}
            className="border border-gray-300 px-4 py-2 rounded hover:bg-gray-50 transition"
          >
            Cancel
          </button>
        </div>
      </form>
    </div>
  );
}
