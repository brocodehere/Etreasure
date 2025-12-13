import React, { useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../../lib/api';
import type { InventoryItem } from '../../types';

interface FormData {
  product_id: string;
  variant_id: string;
  sku: string;
  quantity: number;
  location: string;
  cost_price: string;
}

export function InventoryEditPage() {
  const { id } = useParams<{ id?: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const { data: item, isLoading, error } = useQuery<InventoryItem>({
    queryKey: ['inventory', id],
    queryFn: () => api.get<InventoryItem>(`/inventory/${id}`).then(r => r.data),
    enabled: !!id,
  });

  const createMutation = useMutation({
    mutationFn: (payload: Omit<InventoryItem, 'id' | 'reserved' | 'available' | 'updated_at'>) =>
      api.post<InventoryItem>('/inventory', payload).then(r => r.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['inventory'] });
      navigate('/inventory');
    },
  });

  const updateMutation = useMutation({
    mutationFn: (payload: Partial<InventoryItem>) =>
      api.put<InventoryItem>(`/inventory/${id}`, payload).then(r => r.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['inventory'] });
      navigate('/inventory');
    },
  });

  const adjustMutation = useMutation({
    mutationFn: (payload: { quantity: number; reason: string }) =>
      api.post<InventoryItem>(`/inventory/${id}/adjust`, payload).then(r => r.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['inventory'] });
      queryClient.invalidateQueries({ queryKey: ['inventory', id] });
    },
  });

  const [formData, setFormData] = React.useState<FormData>({
    product_id: '',
    variant_id: '',
    sku: '',
    quantity: 0,
    location: '',
    cost_price: '',
  });
  const [adjustQty, setAdjustQty] = React.useState(0);
  const [adjustReason, setAdjustReason] = React.useState('');

  useEffect(() => {
    if (item) {
      setFormData({
        product_id: item.product_id,
        variant_id: item.variant_id || '',
        sku: item.sku,
        quantity: item.quantity,
        location: item.location || '',
        cost_price: item.cost_price?.toString() || '',
      });
    }
  }, [item]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value, type } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: type === 'number' ? parseInt(value, 10) || 0 : value,
    }));
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const payload = {
      product_id: formData.product_id,
      variant_id: formData.variant_id || null,
      sku: formData.sku,
      quantity: formData.quantity,
      location: formData.location || null,
      cost_price: formData.cost_price ? parseFloat(formData.cost_price) : null,
    };
    if (id) {
      updateMutation.mutate(payload);
    } else {
      createMutation.mutate(payload);
    }
  };

  const handleAdjust = (e: React.FormEvent) => {
    e.preventDefault();
    if (!id) return;
    adjustMutation.mutate({ quantity: adjustQty, reason: adjustReason });
    setAdjustQty(0);
    setAdjustReason('');
  };

  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>Error: {(error as Error).message}</div>;

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold mb-6">{id ? 'Edit Inventory Item' : 'New Inventory Item'}</h1>

      <form onSubmit={handleSubmit} className="space-y-4 max-w-xl">
        <div>
          <label className="block text-sm font-medium mb-1">Product ID</label>
          <input
            name="product_id"
            type="text"
            value={formData.product_id}
            onChange={handleChange}
            required
            className="w-full border rounded px-3 py-2"
          />
        </div>

        <div>
          <label className="block text-sm font-medium mb-1">Variant ID (optional)</label>
          <input
            name="variant_id"
            type="text"
            value={formData.variant_id}
            onChange={handleChange}
            placeholder="UUID"
            className="w-full border rounded px-3 py-2"
          />
        </div>

        <div>
          <label className="block text-sm font-medium mb-1">SKU</label>
          <input
            name="sku"
            type="text"
            value={formData.sku}
            onChange={handleChange}
            required
            className="w-full border rounded px-3 py-2"
          />
        </div>

        <div>
          <label className="block text-sm font-medium mb-1">Quantity</label>
          <input
            name="quantity"
            type="number"
            min="0"
            value={formData.quantity}
            onChange={handleChange}
            required
            className="w-full border rounded px-3 py-2"
          />
        </div>

        <div>
          <label className="block text-sm font-medium mb-1">Location (optional)</label>
          <input
            name="location"
            type="text"
            value={formData.location}
            onChange={handleChange}
            placeholder="Warehouse A, Shelf 12"
            className="w-full border rounded px-3 py-2"
          />
        </div>

        <div>
          <label className="block text-sm font-medium mb-1">Cost Price (optional)</label>
          <input
            name="cost_price"
            type="number"
            step="any"
            min="0"
            value={formData.cost_price}
            onChange={handleChange}
            placeholder="0.00"
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
            onClick={() => navigate('/inventory')}
            className="border border-gray-300 px-4 py-2 rounded hover:bg-gray-50 transition"
          >
            Cancel
          </button>
        </div>
      </form>

      {id && (
        <div className="mt-8 max-w-xl">
          <h2 className="text-lg font-semibold mb-4">Adjust Inventory</h2>
          <form onSubmit={handleAdjust} className="space-y-4">
            <div>
              <label className="block text-sm font-medium mb-1">Adjustment (positive to add, negative to subtract)</label>
              <input
                type="number"
                value={adjustQty}
                onChange={(e) => setAdjustQty(parseInt(e.target.value, 10) || 0)}
                required
                className="w-full border rounded px-3 py-2"
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">Reason</label>
              <input
                type="text"
                value={adjustReason}
                onChange={(e) => setAdjustReason(e.target.value)}
                required
                placeholder="e.g., Stocktake, Return, Damage"
                className="w-full border rounded px-3 py-2"
              />
            </div>
            <button
              type="submit"
              disabled={adjustMutation.isPending}
              className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700 transition disabled:opacity-50"
            >
              {adjustMutation.isPending ? 'Adjusting...' : 'Adjust'}
            </button>
          </form>
        </div>
      )}
    </div>
  );
}
