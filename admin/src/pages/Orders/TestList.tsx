import React from 'react';

export function TestOrdersPage() {
  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold mb-4">Test Orders Page</h1>
      <div className="bg-white p-4 rounded border">
        <p>If you can see this, the admin panel routing is working.</p>
        <p className="mt-2">The issue is likely in the OrdersListPage component.</p>
      </div>
    </div>
  );
}
