import React, { useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../../lib/api';
import type { Customer, Address } from '../../types';

interface FormData {
  email: string;
  first_name: string;
  last_name: string;
  phone: string;
  addresses: Omit<Address, 'id' | 'customer_id'>[];
}

export function CustomerEditPage() {
  const { id } = useParams<{ id?: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const { data: customer, isLoading, error } = useQuery<Customer>({
    queryKey: ['customer', id],
    queryFn: () => api.get<Customer>(`/customers/${id}`).then(r => r.data),
    enabled: !!id,
  });

  const createMutation = useMutation({
    mutationFn: (payload: Omit<Customer, 'id' | 'created_at' | 'updated_at'>) =>
      api.post<Customer>('/customers', payload).then(r => r.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['customers'] });
      navigate('/customers');
    },
  });

  const updateMutation = useMutation({
    mutationFn: (payload: Partial<Customer>) =>
      api.put<Customer>(`/customers/${id}`, payload).then(r => r.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['customers'] });
      navigate('/customers');
    },
  });

  const [formData, setFormData] = React.useState<FormData>({
    email: '',
    first_name: '',
    last_name: '',
    phone: '',
    addresses: [{ type: 'shipping', first_name: '', last_name: '', company: '', address1: '', address2: '', city: '', province: '', country: '', postal_code: '', phone: '' }],
  });

  useEffect(() => {
    if (customer) {
      setFormData({
        email: customer.email,
        first_name: customer.first_name || '',
        last_name: customer.last_name || '',
        phone: customer.phone || '',
        addresses: customer.addresses.map(addr => ({
          type: addr.type,
          first_name: addr.first_name,
          last_name: addr.last_name,
          company: addr.company || '',
          address1: addr.address1,
          address2: addr.address2 || '',
          city: addr.city,
          province: addr.province || '',
          country: addr.country,
          postal_code: addr.postal_code,
          phone: addr.phone || '',
        })),
      });
    }
  }, [customer]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
  };

  const handleAddressChange = (index: number, field: keyof Address, value: any) => {
    setFormData(prev => {
      const addresses = [...prev.addresses];
      addresses[index] = { ...addresses[index], [field]: value };
      return { ...prev, addresses };
    });
  };

  const addAddress = () => {
    setFormData(prev => ({
      ...prev,
      addresses: [...prev.addresses, { type: 'shipping', first_name: '', last_name: '', company: '', address1: '', address2: '', city: '', province: '', country: '', postal_code: '', phone: '' }],
    }));
  };

  const removeAddress = (index: number) => {
    setFormData(prev => ({
      ...prev,
      addresses: prev.addresses.filter((_, i) => i !== index),
    }));
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const payload = {
      email: formData.email,
      first_name: formData.first_name || null,
      last_name: formData.last_name || null,
      phone: formData.phone || null,
      addresses: formData.addresses,
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
      <h1 className="text-2xl font-bold mb-6">{id ? 'Edit Customer' : 'New Customer'}</h1>

      <form onSubmit={handleSubmit} className="space-y-6">
        <div className="grid grid-cols-2 gap-4">
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
          <div>
            <label className="block text-sm font-medium mb-1">Phone</label>
            <input
              name="phone"
              type="tel"
              value={formData.phone}
              onChange={handleChange}
              className="w-full border rounded px-3 py-2"
            />
          </div>
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
          <h3 className="text-lg font-semibold mb-2">Addresses</h3>
          {formData.addresses.map((addr, idx) => (
            <div key={idx} className="border rounded p-3 mb-2 bg-gray-50">
              <div className="grid grid-cols-2 gap-2 mb-2">
                <div>
                  <label className="block text-xs font-medium mb-1">Type</label>
                  <select
                    value={addr.type}
                    onChange={(e) => handleAddressChange(idx, 'type', e.target.value)}
                    className="w-full border rounded px-2 py-1"
                  >
                    <option value="shipping">Shipping</option>
                    <option value="billing">Billing</option>
                  </select>
                </div>
                <div className="flex justify-end">
                  <button
                    type="button"
                    onClick={() => removeAddress(idx)}
                    className="text-xs text-red-600 hover:underline mt-6"
                  >
                    Remove
                  </button>
                </div>
              </div>
              <div className="grid grid-cols-2 gap-2">
                <input type="text" placeholder="First Name" value={addr.first_name} onChange={(e) => handleAddressChange(idx, 'first_name', e.target.value)} className="border rounded px-2 py-1" />
                <input type="text" placeholder="Last Name" value={addr.last_name} onChange={(e) => handleAddressChange(idx, 'last_name', e.target.value)} className="border rounded px-2 py-1" />
                <input type="text" placeholder="Company" value={addr.company} onChange={(e) => handleAddressChange(idx, 'company', e.target.value)} className="border rounded px-2 py-1" />
                <input type="text" placeholder="Phone" value={addr.phone} onChange={(e) => handleAddressChange(idx, 'phone', e.target.value)} className="border rounded px-2 py-1" />
                <input type="text" placeholder="Address 1" value={addr.address1} onChange={(e) => handleAddressChange(idx, 'address1', e.target.value)} className="border rounded px-2 py-1 col-span-2" />
                <input type="text" placeholder="Address 2" value={addr.address2} onChange={(e) => handleAddressChange(idx, 'address2', e.target.value)} className="border rounded px-2 py-1 col-span-2" />
                <input type="text" placeholder="City" value={addr.city} onChange={(e) => handleAddressChange(idx, 'city', e.target.value)} className="border rounded px-2 py-1" />
                <input type="text" placeholder="Province" value={addr.province} onChange={(e) => handleAddressChange(idx, 'province', e.target.value)} className="border rounded px-2 py-1" />
                <input type="text" placeholder="Country" value={addr.country} onChange={(e) => handleAddressChange(idx, 'country', e.target.value)} className="border rounded px-2 py-1" />
                <input type="text" placeholder="Postal Code" value={addr.postal_code} onChange={(e) => handleAddressChange(idx, 'postal_code', e.target.value)} className="border rounded px-2 py-1" />
              </div>
            </div>
          ))}
          <button
            type="button"
            onClick={addAddress}
            className="text-sm text-gold hover:underline"
          >
            + Add Address
          </button>
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
            onClick={() => navigate('/customers')}
            className="border border-gray-300 px-4 py-2 rounded hover:bg-gray-50 transition"
          >
            Cancel
          </button>
        </div>
      </form>
    </div>
  );
}
