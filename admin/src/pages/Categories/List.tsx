import React, { useCallback, useMemo, useState, useRef, useEffect } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '../../lib/api';
import { CategoryRow } from './CategoryRow';
import { MediaSelector } from '../../components/MediaSelector';
import { LoadingState } from '../../components/LoadingSpinner';

// Category type
type Category = {
  uuid_id: string;
  slug: string;
  name: string;
  description?: string | null;
  parent_id?: string | null;
  sort_order: number;
  image_id?: number | null;
  image_path?: string | null;
};

function slugify(input: string) {
  return input.toLowerCase().trim().replace(/[^a-z0-9\s-]/g, '').replace(/\s+/g, '-').replace(/-+/g, '-');
}

/**
 * NOTE: This version assumes your GET /categories returns
 * a minimal payload for lists. If your backend returns more,
 * consider adding query params to limit fields server-side.
 */

export const CategoriesListPage: React.FC = () => {
  const qc = useQueryClient();
  
  // Toast message state
  const [showSuccess, setShowSuccess] = useState(false);
  const [successMessage, setSuccessMessage] = useState('');

  // Show success message function
  const showSuccessMessage = (message: string) => {
    setSuccessMessage(message);
    setShowSuccess(true);
    setTimeout(() => {
      setShowSuccess(false);
    }, 3000); // Hide after 3 seconds
  };

  // Query: tuned for performance
  const { data, isLoading, error } = useQuery<{ items: Category[] }>({
    queryKey: ['categories'],
    queryFn: () => api.get<{ items: Category[] }>('/categories'),
    staleTime: 1000 * 60 * 5, // 5 minutes
    gcTime: 1000 * 60 * 30, // 30 minutes
    refetchOnWindowFocus: false,
    refetchOnReconnect: false,
    retry: (failureCount: number, error: any) => {
      // Don't retry on 4xx errors
      if (error?.status >= 400 && error?.status < 500) return false;
      return failureCount < 3;
    },
    select: (res) => ({ items: res.items.map((it: any) => ({
      uuid_id: it.uuid_id,
      slug: it.slug,
      name: it.name,
      description: it.description ?? null,
      parent_id: it.parent_id ?? null,
      sort_order: it.sort_order ?? 0,
    })) }),
  });

  // form state
  const [name, setName] = useState<string>('');
  const [slug, setSlug] = useState<string>('');
  const [description, setDescription] = useState<string>('');
  const [parentId, setParentId] = useState<string | ''>('');
  const [imageId, setImageId] = useState<number | null>(null);
  const [editingId, setEditingId] = useState<string | null>(null);
  const userEditedSlug = useRef(false);

  // If user hasn't typed slug manually, auto-fill slug when name changes
  useEffect(() => {
    if (!userEditedSlug.current) {
      setSlug(name ? slugify(name) : '');
    }
  }, [name]);

  const markSlugEdited = useCallback(() => {
    userEditedSlug.current = true;
  }, []);

  // Optimistic create mutation
  const createMutation = useMutation({
    mutationFn: async (payload: { slug: string; name: string; description?: string; parent_id?: string; image_id?: number | null }) => {
            try {
        const result = await api.post('/categories', payload);
                return result;
      } catch (error) {
        console.error('API call failed:', error);
        throw error;
      }
    },
    onMutate: async (payload) => {
      await qc.cancelQueries({ queryKey: ['categories'] });

      const previous = qc.getQueryData<{ items: Category[] }>(['categories']);

      // optimistic ID (negative temporary)
      const tempId = Math.floor(Math.random() * -1000000);

      const optimisticItem: Category = {
        uuid_id: `temp-${tempId}`,
        slug: payload.slug,
        name: payload.name,
        description: payload.description ?? null,
        parent_id: payload.parent_id ?? null,
        sort_order: 0,
      };

      qc.setQueryData<{ items: Category[] } | undefined>(['categories'], (old) => {
        if (!old || !old.items) return { items: [optimisticItem] };
        return { items: [optimisticItem, ...old.items] };
      });

      return { previous };
    },
    onError: (_err, _variables, context: any) => {
      console.error('Mutation error:', _err);
      // rollback
      if (context?.previous) {
        qc.setQueryData(['categories'], context.previous);
      }
    },
    onSuccess: (data) => {
            qc.invalidateQueries({ queryKey: ['categories'] });
      // Show success message
      showSuccessMessage('Category created successfully!');
      // reset form
      setName(''); setSlug(''); setDescription(''); setParentId(''); setImageId(null); userEditedSlug.current = false;
    },
  });

  // Optimistic update mutation
  const updateMutation = useMutation({
    mutationFn: async ({ id, payload }: { id: string; payload: { slug: string; name: string; description?: string; parent_id?: string; image_id?: number | null } }) => {
      return await api.put(`/categories/${id}`, payload);
    },
    onMutate: async ({ id, payload }) => {
      await qc.cancelQueries({ queryKey: ['categories'] });
      const previous = qc.getQueryData<{ items: Category[] }>(['categories']);
      qc.setQueryData<{ items: Category[] } | undefined>(['categories'], (old) => {
        if (!old) return { items: [] };
        return {
          items: old.items.map((it) =>
            it.uuid_id === id
              ? { ...it, ...payload, parent_id: payload.parent_id ?? null }
              : it
          ),
        };
      });
      return { previous };
    },
    onError: (_err, _variables, context: any) => {
      if (context?.previous) {
        qc.setQueryData(['categories'], context.previous);
      }
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['categories'] });
      showSuccessMessage('Category updated successfully!');
      resetForm();
    },
  });

  // Optimistic delete
  const deleteMutation = useMutation({
    mutationFn: async (id: string) => api.delete(`/categories/${id}`),
    onMutate: async (id: string) => {
      await qc.cancelQueries({ queryKey: ['categories'] });
      const previous = qc.getQueryData<{ items: Category[] }>(['categories']);
      qc.setQueryData<{ items: Category[] } | undefined>(['categories'], (old) => {
        if (!old) return { items: [] };
        return { items: old.items.filter((it) => it.uuid_id !== id) };
      });
      return { previous };
    },
    onError: (_err, _variables, context: any) => {
      if (context?.previous) {
        qc.setQueryData(['categories'], context.previous);
      }
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['categories'] });
    },
  });

  const items: Category[] = data?.items ?? [];

  // Memoized roots & children map (keeps work minimal)
  const roots = useMemo(() => items.filter((it) => it.parent_id == null), [items]);

  const childrenMap = useMemo(() => {
    const map = new Map<string, Category[]>();
    for (const it of items) {
      if (it.parent_id != null) {
        const arr = map.get(it.parent_id) ?? [];
        arr.push(it);
        map.set(it.parent_id, arr);
      }
    }
    return map;
  }, [items]);

  // Handlers - memoized
  const handleCreate = useCallback(() => {
        if (!name.trim()) {
            // consider toast or validation UI
      return;
    }
    const payload = {
      slug: slug || slugify(name),
      name,
      description: description || undefined,
      parent_id: parentId === '' ? undefined : parentId,
      image_id: imageId,
    };
        createMutation.mutate(payload);
  }, [name, slug, description, parentId, imageId, createMutation]);

  const resetForm = useCallback(() => {
    setName('');
    setSlug('');
    setDescription('');
    setParentId('');
    setImageId(null);
    setEditingId(null);
    userEditedSlug.current = false;
  }, []);

  const handleEdit = useCallback((category: Category) => {
    setName(category.name);
    setSlug(category.slug);
    setDescription(category.description || '');
    setParentId(category.parent_id || '');
    setImageId(category.image_id || null);
    setEditingId(category.uuid_id);
    userEditedSlug.current = true;
  }, []);

  const handleUpdate = useCallback(() => {
    if (!editingId || !name.trim()) return;
    
    const payload = {
      slug: slug || slugify(name),
      name,
      description: description || undefined,
      parent_id: parentId === '' ? undefined : parentId,
      image_id: imageId,
    };
    
    updateMutation.mutate({ id: editingId, payload });
  }, [editingId, name, slug, description, parentId, imageId, updateMutation]);

  const handleDelete = useCallback((id: string) => {
    if (!confirm('Delete category? This cannot be undone.')) return;
    try {
      deleteMutation.mutate(id);
    } catch (err) {
      console.error('Error deleting category:', err);
      alert('Failed to delete category. Please try again.');
    }
  }, [deleteMutation]);

  return (
    <LoadingState isLoading={isLoading} error={error}>
      <div className="space-y-6">
      {/* Success Message */}
      {showSuccess && (
        <div className="fixed top-4 right-4 z-50 animate-pulse">
          <div className="bg-green-500 text-white px-6 py-3 rounded-lg shadow-lg flex items-center space-x-2">
            <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
            </svg>
            <span className="font-medium">{successMessage}</span>
          </div>
        </div>
      )}

      <header>
        <h1 className="text-2xl font-playfair text-maroon">Categories</h1>
        <p className="text-sm text-dark/70 mt-1">Create and manage category hierarchy.</p>
      </header>

      <div className="bg-white border border-gold/30 rounded-lg p-4 shadow-card">
        <h2 className="text-lg font-playfair text-maroon mb-4">
          {editingId ? 'Edit Category' : 'Create New Category'}
        </h2>
        <div className="grid grid-cols-1 md:grid-cols-4 gap-3">
          <input
            placeholder="Name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="rounded-md border border-gold/40 bg-cream/60 px-3 py-2 text-sm"
          />
          <input
            placeholder="Slug (auto if blank)"
            value={slug}
            onChange={(e) => { markSlugEdited(); setSlug(e.target.value); }}
            className="rounded-md border border-gold/40 bg-cream/60 px-3 py-2 text-sm"
          />
          <input
            placeholder="Description"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            className="rounded-md border border-gold/40 bg-cream/60 px-3 py-2 text-sm"
          />
          <select
            value={parentId}
            onChange={(e) => {
              const val = e.target.value;
              setParentId(val === '' ? '' : val);
            }}
            className="rounded-md border border-gold/40 bg-cream/60 px-3 py-2 text-sm"
          >
            <option value="">No parent</option>
            {roots.map((r) => (
              <option key={r.uuid_id} value={r.uuid_id}>{r.name}</option>
            ))}
          </select>
        </div>
        <div className="mt-3">
          <label className="text-sm font-medium text-dark/70 block mb-1">Category Image</label>
          <MediaSelector 
            selectedId={imageId}
            onSelect={(media) => setImageId(media?.id || null)}
          />
        </div>
        <div className="mt-3">
          <button
            onClick={editingId ? handleUpdate : handleCreate}
            disabled={createMutation.isPending || updateMutation.isPending}
            className="inline-flex items-center justify-center rounded-md bg-maroon text-cream text-sm font-medium px-3 py-2 hover:bg-maroon/90 disabled:opacity-60 mr-2"
          >
            {createMutation.isPending || updateMutation.isPending 
              ? (editingId ? 'Updating…' : 'Creating…')
              : (editingId ? 'Update Category' : 'Create Category')
            }
          </button>
          {editingId && (
            <button
              onClick={resetForm}
              className="inline-flex items-center justify-center rounded-md border border-gold/40 text-maroon text-sm font-medium px-3 py-2 hover:bg-gold/10"
            >
              Cancel
            </button>
          )}
        </div>
      </div>

      {isLoading && (
        <div className="flex items-center justify-center py-12">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-maroon"></div>
          <span className="ml-2 text-dark/70">Loading categories...</span>
        </div>
      )}

      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4">
          <p className="text-sm text-red-600">Error loading categories: {String((error as any)?.message || error)}</p>
        </div>
      )}

      {!isLoading && !error && items.length === 0 && (
        <div className="bg-white border border-gold/30 rounded-lg p-8 shadow-card text-center">
          <div className="text-dark/60">
            <svg className="w-16 h-16 mx-auto mb-4 text-gold/40" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
            </svg>
            <p className="text-lg font-medium mb-2">No categories yet</p>
            <p className="text-sm">Create your first category using the form above</p>
          </div>
        </div>
      )}

      {items.length > 0 && (
        <div className="bg-white border border-gold/30 rounded-lg p-4 shadow-card">
          <ul className="space-y-2">
            {roots.map((r) => (
              <CategoryRow 
                key={r.uuid_id} 
                cat={r} 
                children={childrenMap.get(r.uuid_id) ?? []} 
                onDelete={handleDelete}
                onEdit={handleEdit}
              />
            ))}
          </ul>
        </div>
      )}
      </div>
    </LoadingState>
  );
};
