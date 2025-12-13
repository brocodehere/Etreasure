// src/components/CategoryRow.tsx
import React, { useCallback } from 'react';
import { getPublicImageUrl } from '../../lib/useMediaUpload';

type Category = {
  uuid_id: string;
  slug: string;
  name: string;
  description?: string | null;
  parent_id?: string | null;
  sort_order: number;
  image_id?: number | null;
  image_path?: string | null;
  image_url?: string | null; // Add the new image_url field
};

export const CategoryRow: React.FC<{ cat: Category; children?: Category[]; onDelete: (id: string) => void; onEdit: (category: Category) => void }> = React.memo(({ cat, children = [], onDelete, onEdit }) => {
  const handleDelete = useCallback(() => {
    try {
      onDelete(cat.uuid_id);
    } catch (err) {
      console.error('Error deleting category:', err);
    }
  }, [cat.uuid_id, onDelete]);

  const handleChildDelete = useCallback((childId: string) => {
    try {
      onDelete(childId);
    } catch (err) {
      console.error('Error deleting child category:', err);
    }
  }, [onDelete]);

  return (
    <li>
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-3">
          {(cat.image_url || cat.image_path) && (
            <img 
              src={cat.image_url || getPublicImageUrl(cat.image_path)}
              alt={cat.name}
              className="w-8 h-8 object-cover rounded"
            />
          )}
          <div>
            <div className="font-medium">{cat.name || 'Unnamed'}</div>
            <div className="text-xs text-dark/60">/{cat.slug || 'no-slug'}</div>
          </div>
        </div>
        <div className="flex space-x-2">
          <button 
            onClick={() => onEdit(cat)} 
            className="text-gold text-sm hover:text-gold/80"
          >
            Edit
          </button>
          <button onClick={handleDelete} className="text-maroon text-sm hover:text-maroon/80">Delete</button>
        </div>
      </div>
      {children && children.length > 0 && (
        <ul className="mt-2 ml-4 list-disc">
          {children.map((c) => (
            <li key={c.uuid_id} className="flex items-center justify-between">
              <span>{c.name || 'Unnamed'}</span>
              <button onClick={() => handleChildDelete(c.uuid_id)} className="text-maroon text-sm">Delete</button>
            </li>
          ))}
        </ul>
      )}
    </li>
  );
});
