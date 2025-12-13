import React, { useState } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '../lib/api';
import { ImageUploader } from './ImageUploader';

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

type Props = {
  selectedId?: number | null;
  onSelect: (media: MediaItem | null) => void;
  className?: string;
};

export const MediaSelector: React.FC<Props> = ({ selectedId, onSelect, className = "" }) => {
  const qc = useQueryClient();
  const [isOpen, setIsOpen] = useState(false);
  
  const { data, isLoading, error } = useQuery<{ items: MediaItem[] }>({
    queryKey: ['media'],
    queryFn: () => api.get<{ items: MediaItem[] }>('/media'),
  });

  const selectedMedia = data?.items.find(m => m.id === selectedId);

  if (!isOpen) {
    return (
      <div className={`space-y-2 ${className}`}>
        {selectedMedia ? (
          <div className="relative group">
            <img 
              src={selectedMedia.url}
              alt="Selected category image"
              className="w-24 h-24 object-cover rounded border-2 border-gold/40"
            />
            <button
              type="button"
              onClick={() => onSelect(null)}
              className="absolute -top-2 -right-2 bg-red-500 text-white rounded-full w-6 h-6 flex items-center justify-center text-xs hover:bg-red-600"
            >
              ×
            </button>
          </div>
        ) : (
          <div className="w-24 h-24 border-2 border-dashed border-gold/40 rounded flex items-center justify-center">
            <button
              type="button"
              onClick={() => setIsOpen(true)}
              className="text-gold hover:text-gold/80 text-sm"
            >
              Add Image
            </button>
          </div>
        )}
        <button
          type="button"
          onClick={() => setIsOpen(true)}
          className="text-xs text-maroon hover:text-maroon/80"
        >
          {selectedMedia ? 'Change Image' : 'Select Image'}
        </button>
      </div>
    );
  }

  return (
    <div className={`fixed inset-0 bg-black/50 flex items-center justify-center z-50 ${className}`}>
      <div className="bg-white rounded-lg p-6 max-w-4xl max-h-[80vh] overflow-y-auto w-full">
        <div className="flex justify-between items-center mb-4">
          <h3 className="text-lg font-semibold">Select Image</h3>
          <button
            type="button"
            onClick={() => setIsOpen(false)}
            className="text-gray-500 hover:text-gray-700"
          >
            ×
          </button>
        </div>

        <ImageUploader onUploaded={() => qc.invalidateQueries({ queryKey: ['media'] })} />

        {isLoading && (
          <div className="text-center py-8">Loading media...</div>
        )}

        {error && (
          <div className="text-center py-8 text-red-500">
            Failed to load media
          </div>
        )}

        {data && (
          <div className="grid grid-cols-4 gap-4 mt-4">
            {data.items.map((media) => (
              <div
                key={media.id}
                className={`relative group cursor-pointer border-2 rounded overflow-hidden ${
                  selectedId === media.id ? 'border-gold' : 'border-gray-200 hover:border-gold/50'
                }`}
                onClick={() => {
                  onSelect(media);
                  setIsOpen(false);
                }}
              >
                <img
                  src={media.url}
                  alt={media.path}
                  className="w-full h-24 object-cover"
                />
                {selectedId === media.id && (
                  <div className="absolute top-1 right-1 bg-gold text-white rounded-full w-6 h-6 flex items-center justify-center text-xs">
                    ✓
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};
