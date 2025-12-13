import React, { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '../../lib/api';

interface SimpleOrder {
  id: string;
  order_number: string;
  customer_name: string;
  status: string;
  total_price: number;
  created_at: string;
}

interface ListResponse {
  data: SimpleOrder[];
  next_cursor?: string;
}

interface Order {
  id: string;
  order_number: string;
  customer_name: string;
  status: string;
  total_price: number;
  created_at: string;
}

export function SimpleOrdersPage() {
  const [limit] = useState(10);

  const { data, isLoading, error } = useQuery({
    queryKey: ['orders-simple', limit],
    queryFn: () => {
      return api.get<{data: Order[]}>(`/orders?limit=${limit}`).then((r: any) => {
        return r.data;
      }).catch((err) => {
        console.error('Simple Orders API error:', err);
        throw err;
      });
    },
  });

  
  if (isLoading) return <div className="p-6">Loading...</div>;
  if (error) return <div className="p-6">Error: {(error as Error).message}</div>;

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold mb-6">Simple Orders</h1>
      
      <div className="mb-4 p-4 bg-gray-100 rounded">
        <p>Loading: {isLoading ? 'true' : 'false'}</p>
        <p>Error: {error ? (error as Error).message : 'none'}</p>
        <p>Data length: {data?.data?.length || 0}</p>
        <p>Raw data: {JSON.stringify(data, null, 2)}</p>
      </div>

      {data?.data && data.data.length > 0 ? (
        <div className="space-y-4">
          {data.data.map((order: Order) => (
            <div key={order.id} className="bg-white p-4 border rounded">
              <h3 className="font-semibold">{order.order_number}</h3>
              <p>Customer: {order.customer_name}</p>
              <p>Status: {order.status}</p>
              <p>Total: {order.total_price}</p>
              <p>Date: {new Date(order.created_at).toLocaleDateString()}</p>
            </div>
          ))}
        </div>
      ) : (
        <div className="text-center py-12 text-gray-500">
          No orders found.
        </div>
      )}
    </div>
  );
}
