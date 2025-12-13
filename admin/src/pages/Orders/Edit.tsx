import React, { useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../../lib/api';
import type { Order, Address, FormAddress, CreateOrderLineItem } from '../../types';

interface FormData {
  customer_id: string;
  currency: string;
  line_items: CreateOrderLineItem[];
  shipping_address: FormAddress | null;
  billing_address: FormAddress | null;
  notes: string;
  status?: string;
}

export function OrderEditPage() {
  const { id } = useParams<{ id?: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const { data: order, isLoading, error } = useQuery<Order>({
    queryKey: ['order', id],
    queryFn: () => api.get<Order>(`/orders/${id}`).then(r => {
      const orderData = r;
      if (!orderData) {
        throw new Error('Order data not found');
      }
      return orderData;
    }),
    enabled: !!id,
  });

  const createMutation = useMutation({
    mutationFn: (payload: Omit<Order, 'id' | 'order_number' | 'created_at' | 'updated_at'>) =>
      api.post<Order>('/orders', payload).then(r => r),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['orders'] });
      navigate('/orders');
    },
  });

  const updateMutation = useMutation({
    mutationFn: (payload: Partial<Order>) =>
      api.put<Order>(`/orders/${id}`, payload).then(r => r),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['orders'] });
      navigate('/orders');
    },
  });

  const [formData, setFormData] = React.useState<FormData>({
    customer_id: '',
    currency: 'USD',
    line_items: [{ product_id: '', variant_id: '', quantity: 1, price: 0 }],
    shipping_address: null,
    billing_address: null,
    notes: '',
    status: undefined,
  });

  useEffect(() => {
    if (order) {
      setFormData({
        customer_id: order.user_id?.toString() || '',
        currency: order.currency,
        line_items: order.line_items ? order.line_items.map(li => ({
          product_id: li.product_id,
          variant_id: li.variant_id || '',
          quantity: li.quantity,
          price: li.price,
        })) : [{ product_id: '', variant_id: '', quantity: 1, price: 0 }],
        shipping_address: order.shipping_name ? {
          street: order.shipping_address_line1 || '',
          city: order.shipping_city || '',
          state: order.shipping_state || '',
          country: order.shipping_country || '',
          postal_code: order.shipping_pin_code || '',
        } : null,
        billing_address: order.billing_name ? {
          street: order.billing_address_line1 || '',
          city: order.billing_city || '',
          state: order.billing_state || '',
          country: order.billing_country || '',
          postal_code: order.billing_pin_code || '',
        } : null,
        notes: order.notes || '',
        status: id ? order.status : undefined,
      });
    }
  }, [order, id]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    const { name, value, type } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: type === 'checkbox' ? (e.target as HTMLInputElement).checked : value,
    }));
  };

  const handleLineItemChange = (index: number, field: keyof CreateOrderLineItem, value: any) => {
    setFormData(prev => {
      const items = [...prev.line_items];
      items[index] = { ...items[index], [field]: value };
      return { ...prev, line_items: items };
    });
  };

  const addLineItem = () => {
    setFormData(prev => ({
      ...prev,
      line_items: [...prev.line_items, { product_id: '', variant_id: '', quantity: 1, price: 0 }],
    }));
  };

  const removeLineItem = (index: number) => {
    setFormData(prev => ({
      ...prev,
      line_items: prev.line_items.filter((_, i) => i !== index),
    }));
  };

  const handleAddressChange = (type: 'shipping' | 'billing', field: keyof FormAddress, value: any) => {
    setFormData(prev => ({
      ...prev,
      [`${type}_address`]: {
        ...(prev[`${type}_address`] || {}),
        [field]: value,
      },
    }));
  };

  const calculateTotal = (lineItems: CreateOrderLineItem[]) => {
    return lineItems.reduce((acc, item) => acc + item.price * item.quantity, 0);
  };

  const calculateSubtotal = (lineItems: CreateOrderLineItem[]) => {
    return lineItems.reduce((acc, item) => acc + item.price * item.quantity, 0);
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const payload = {
      status: (formData.status || 'pending') as 'pending' | 'pending_payment' | 'processing' | 'shipped' | 'delivered' | 'cancelled' | 'paid' | 'just_arrived',
      user_id: formData.customer_id ? parseInt(formData.customer_id) : undefined,
      currency: formData.currency,
      line_items: formData.line_items.map(li => ({
        id: '', // Will be generated by backend
        order_id: '', // Will be set by backend
        product_id: li.product_id,
        variant_id: li.variant_id,
        title: '', // Will be populated by backend
        sku: '', // Will be populated by backend
        quantity: li.quantity,
        price: li.price,
        total: li.price * li.quantity,
      })),
      shipping_address: formData.shipping_address,
      billing_address: formData.billing_address,
      notes: formData.notes || undefined,
      total_price: calculateTotal(formData.line_items),
      subtotal: calculateSubtotal(formData.line_items),
      tax_amount: 0,
      shipping_amount: 0,
      discount_amount: 0,
      customer_name: '',
      customer_email: '',
      customer_phone: '',
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
      <h1 className="text-2xl font-bold mb-6">{id ? 'Edit Order' : 'New Order'}</h1>

      <form onSubmit={handleSubmit} className="space-y-6">
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium mb-1">Customer ID (optional)</label>
            <input
              name="customer_id"
              type="text"
              value={formData.customer_id}
              onChange={handleChange}
              placeholder="UUID"
              className="w-full border rounded px-3 py-2"
            />
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">Currency</label>
            <select
              name="currency"
              value={formData.currency}
              onChange={handleChange}
              className="w-full border rounded px-3 py-2"
            >
              <option value="USD">USD</option>
              <option value="EUR">EUR</option>
              <option value="GBP">GBP</option>
            </select>
          </div>
        </div>

        {id && (
          <div>
            <label className="block text-sm font-medium mb-1">Status</label>
            <select
              name="status"
              value={formData.status}
              onChange={handleChange}
              className="w-full border rounded px-3 py-2"
            >
              <option value="pending">Pending</option>
              <option value="processing">Processing</option>
              <option value="shipped">Shipped</option>
              <option value="delivered">Delivered</option>
              <option value="cancelled">Cancelled</option>
            </select>
          </div>
        )}

        <div>
          <h3 className="text-lg font-semibold mb-2">Line Items</h3>
          {formData.line_items.map((item, idx) => (
            <div key={idx} className="border rounded p-3 mb-2 bg-gray-50">
              <div className="grid grid-cols-4 gap-2">
                <input
                  type="text"
                  placeholder="Product ID"
                  value={item.product_id}
                  onChange={(e) => handleLineItemChange(idx, 'product_id', e.target.value)}
                  className="border rounded px-2 py-1"
                  required
                />
                <input
                  type="text"
                  placeholder="Variant ID"
                  value={item.variant_id}
                  onChange={(e) => handleLineItemChange(idx, 'variant_id', e.target.value)}
                  className="border rounded px-2 py-1"
                />
                <input
                  type="number"
                  min="1"
                  placeholder="Qty"
                  value={item.quantity}
                  onChange={(e) => handleLineItemChange(idx, 'quantity', parseInt(e.target.value, 10))}
                  className="border rounded px-2 py-1"
                  required
                />
                <input
                  type="number"
                  step="any"
                  min="0"
                  placeholder="Price"
                  value={item.price}
                  onChange={(e) => handleLineItemChange(idx, 'price', parseFloat(e.target.value))}
                  className="border rounded px-2 py-1"
                  required
                />
              </div>
              <button
                type="button"
                onClick={() => removeLineItem(idx)}
                className="mt-2 text-xs text-red-600 hover:underline"
              >
                Remove
              </button>
            </div>
          ))}
          <button
            type="button"
            onClick={addLineItem}
            className="text-sm text-gold hover:underline"
          >
            + Add Line Item
          </button>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <h3 className="text-lg font-semibold mb-2">Shipping Address</h3>
            <AddressForm address={formData.shipping_address} onChange={(field, value) => handleAddressChange('shipping', field, value)} />
          </div>
          <div>
            <h3 className="text-lg font-semibold mb-2">Billing Address</h3>
            <AddressForm address={formData.billing_address} onChange={(field, value) => handleAddressChange('billing', field, value)} />
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium mb-1">Notes</label>
          <textarea
            name="notes"
            value={formData.notes}
            onChange={handleChange}
            rows={3}
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
            onClick={() => navigate('/orders')}
            className="border border-gray-300 px-4 py-2 rounded hover:bg-gray-50 transition"
          >
            Cancel
          </button>
        </div>
      </form>
    </div>
  );
}

interface AddressFormProps {
  address: FormAddress | null;
  onChange: (field: keyof FormAddress, value: any) => void;
}

function AddressForm({ address, onChange }: AddressFormProps) {
  const handleChange = (field: keyof FormAddress, e: React.ChangeEvent<HTMLInputElement>) => {
    onChange(field, e.target.value);
  };
  return (
    <div className="space-y-2">
      <input type="text" placeholder="Street Address" value={address?.street || ''} onChange={(e) => handleChange('street', e)} className="w-full border rounded px-2 py-1" />
      <input type="text" placeholder="City" value={address?.city || ''} onChange={(e) => handleChange('city', e)} className="w-full border rounded px-2 py-1" />
      <input type="text" placeholder="State" value={address?.state || ''} onChange={(e) => handleChange('state', e)} className="w-full border rounded px-2 py-1" />
      <input type="text" placeholder="Country" value={address?.country || ''} onChange={(e) => handleChange('country', e)} className="w-full border rounded px-2 py-1" />
      <input type="text" placeholder="Postal Code" value={address?.postal_code || ''} onChange={(e) => handleChange('postal_code', e)} className="w-full border rounded px-2 py-1" />
    </div>
  );
}
