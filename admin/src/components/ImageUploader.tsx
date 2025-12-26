import React, { useCallback, useRef, useState } from 'react';
import { uploadImage } from '../lib/useMediaUpload';

type Props = {
  onUploaded?: (result: { key: string; url: string }) => void;
  type?: 'product' | 'banner' | 'category';
};

export const ImageUploader: React.FC<Props> = ({ onUploaded, type = 'product' }) => {
  const inputRef = useRef<HTMLInputElement | null>(null);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleFiles = useCallback(async (files: FileList | null) => {
    if (!files || files.length === 0) return;
    setUploading(true);
    setError(null);
    try {
      const uploadPromises = Array.from(files).map(file => uploadImage(file, type));
      const results = await Promise.all(uploadPromises);
      results.forEach(result => onUploaded?.(result));
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Upload error');
    } finally {
      setUploading(false);
    }
  }, [onUploaded, type]);

  return (
    <div className="border-2 border-dashed border-gold/50 rounded-lg p-4 bg-cream/40">
      <div className="flex items-center justify-between gap-4">
        <div>
          <div className="text-sm font-medium text-dark">Upload images</div>
          <div className="text-xs text-dark/60">AVIF/WEBP/JPG/PNG up to ~5MB. Multiple files supported.</div>
        </div>
        <div className="flex items-center gap-2">
          <input
            ref={inputRef}
            type="file"
            accept="image/*"
            multiple
            className="hidden"
            onChange={(e) => handleFiles(e.target.files)}
          />
          <button
            type="button"
            onClick={() => inputRef.current?.click()}
            disabled={uploading}
            className="inline-flex items-center justify-center rounded-md bg-maroon text-cream text-sm font-medium px-3 py-2 hover:bg-maroon/90 disabled:opacity-60"
          >
            {uploading ? 'Uploading…' : 'Choose Files'}
          </button>
        </div>
      </div>
      {error && <p className="mt-2 text-sm text-red-600">{error}</p>}
    </div>
  );
};
