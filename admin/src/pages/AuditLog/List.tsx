import React, { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '../../lib/api';
import type { AuditLog } from '../../types';

interface ListResponse {
  data: AuditLog[];
  next_cursor?: string;
}

export function AuditLogListPage() {
  const [cursor, setCursor] = useState<string | undefined>();
  const [limit] = useState(50);

  const { data, isLoading, error } = useQuery<ListResponse>({
    queryKey: ['audit-log', cursor, limit],
    queryFn: () => {
      const url = cursor ? `/audit-log?cursor=${cursor}&limit=${limit}` : `/audit-log?limit=${limit}`;
      return api.get<ListResponse>(url).then((r: any) => r.data);
    },
  });

  const loadMore = () => {
    if (data?.next_cursor) {
      setCursor(data.next_cursor);
    }
  };

  if (isLoading && !data) return <div>Loading...</div>;
  if (error) return <div>Error: {(error as Error).message}</div>;

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold mb-6">Audit Log</h1>

      {data?.data && data.data.length > 0 ? (
        <>
          <ul className="space-y-4">
            {data.data.map((log: AuditLog) => (
              <li key={log.id} className="border rounded-lg p-4 bg-white shadow-sm">
                <div className="flex justify-between items-start">
                  <div>
                    <h3 className="font-semibold text-lg">{log.action}</h3>
                    <p className="text-gray-600 text-sm mt-1">
                      {log.resource_type} {log.resource_id ? `(${log.resource_id})` : ''}
                    </p>
                    <div className="mt-2 text-sm text-gray-600">
                      <span>By: {log.user_email}</span>
                      <span className="mx-2">•</span>
                      <span>IP: {log.ip_address}</span>
                    </div>
                    {log.details && (
                      <details className="mt-2">
                        <summary className="text-xs text-gray-500 cursor-pointer hover:text-gray-700">Details</summary>
                        <pre className="text-xs text-gray-600 mt-1 bg-gray-50 p-2 rounded overflow-auto">
                          {JSON.stringify(log.details, null, 2)}
                        </pre>
                      </details>
                    )}
                    <div className="mt-1 text-xs text-gray-500">
                      {new Date(log.created_at).toLocaleString()}
                    </div>
                  </div>
                </div>
              </li>
            ))}
          </ul>

          {data.next_cursor && (
            <div className="mt-6 text-center">
              <button
                onClick={loadMore}
                disabled={isLoading}
                className="bg-maroon text-white px-4 py-2 rounded hover:bg-red-900 transition disabled:opacity-50"
              >
                {isLoading ? 'Loading...' : 'Load More'}
              </button>
            </div>
          )}
        </>
      ) : (
        <div className="text-center py-12 text-gray-500">
          No audit logs yet.
        </div>
      )}
    </div>
  );
}
