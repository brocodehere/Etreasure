import React from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '../lib/api';
import { ImageUploader } from '../components/ImageUploader';

type MediaItem = {
  id: number;
  path: string;
  url: string;
  mime_type: string;
  file_size_bytes: number;
  width?: number;
  height?: number;
  created_at: string;
};

type MediaListResponse = {
  items: Array<{
    id: number;
    path: string;
    url: string;
    mime_type: string;
    file_size_bytes: number;
    width?: number;
    height?: number;
    created_at: string;
  }>;
  nextCursor?: number;
};

export const MediaLibraryPage: React.FC = () => {
  const qc = useQueryClient();
  const { data, isLoading, error } = useQuery<MediaListResponse>({
    queryKey: ['media', { first: 50 }],
    queryFn: () => api.get<MediaListResponse>(`/media?first=50`),
  });

  async function handleDelete(id: number) {
    await api.delete(`/media/${id}`);
    qc.invalidateQueries({ queryKey: ['media'] });
  }

  return (
    <div className="space-y-6">
      <header>
        <h1 className="text-2xl font-playfair text-maroon">Media Library</h1>
        <p className="text-sm text-dark/70 mt-1">Upload images and manage assets.</p>
      </header>

      <ImageUploader onUploaded={() => qc.invalidateQueries({ queryKey: ['media'] })} />

      {isLoading && (
        <div className="flex items-center justify-center py-12">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-maroon"></div>
          <span className="ml-2 text-dark/70">Loading media...</span>
        </div>
      )}
      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4">
          <p className="text-sm text-red-600">Error loading media: {(error as Error).message}</p>
        </div>
      )}

      {data && (
        <>
          {data.items.length === 0 ? (
            <div className="bg-white border border-gold/30 rounded-lg p-8 shadow-card text-center">
              <div className="text-dark/60 mb-4">
                <svg className="w-16 h-16 mx-auto mb-4 text-gold/40" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                </svg>
                <p className="text-lg font-medium mb-2">No media files yet</p>
                <p className="text-sm">Upload your first image using the uploader above</p>
              </div>
            </div>
          ) : (
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-4">
          {data.items.map((m) => (
            <div key={m.id} className="bg-white border border-gold/30 rounded-lg p-2 shadow-card">
              {m.mime_type.startsWith('image/') ? (
                <img
                  src={m.url}
                  alt={m.path}
                  className="w-full h-32 object-cover rounded"
                  loading="lazy"
                />
              ) : (
                <div className="h-32 flex items-center justify-center text-xs text-dark/60">{m.mime_type}</div>
              )}
              <div className="mt-2 flex items-center justify-between text-xs text-dark/70">
                <span>#{m.id}</span>
                <button
                  onClick={() => handleDelete(m.id)}
                  className="text-maroon hover:underline"
                  aria-label={`Delete media ${m.id}`}
                >
                  Delete
                </button>
              </div>
            </div>
          ))}
        </div>
          )}
        </>
      )}
    </div>
  );
};
