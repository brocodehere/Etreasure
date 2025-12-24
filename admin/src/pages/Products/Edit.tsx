import React, { useState, useEffect, useMemo } from 'react';
import { useNavigate, useParams, Link } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../../lib/api';
import { ImageUploader } from '../../components/ImageUploader';
import { getPublicImageUrl } from '../../lib/useMediaUpload';

// Error boundary component
class ErrorBoundary extends React.Component<
  { children: React.ReactNode },
  { hasError: boolean; error?: Error }
> {
  constructor(props: { children: React.ReactNode }) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error: Error) {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {}

  render() {
    if (this.state.hasError) {
      return (
        <div className="space-y-6">
          <header>
            <h1 className="text-2xl font-playfair text-maroon">Product Error</h1>
          </header>
          <div className="bg-red-50 border border-red-200 rounded-lg p-4">
            <p className="text-sm text-red-600 font-medium">Something went wrong</p>
            <p className="text-xs text-red-500 mt-1">{this.state.error?.message}</p>
            <div className="mt-3 flex gap-2">
              <Link
                to="/products"
                className="text-xs bg-red-100 text-red-700 px-2 py-1 rounded hover:bg-red-200"
              >
                Back to Products
              </Link>
              <button
                onClick={() => window.location.reload()}
                className="text-xs bg-red-100 text-red-700 px-2 py-1 rounded hover:bg-red-200"
              >
                Retry
              </button>
            </div>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}

function slugify(input: string) {
  return input
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9\s-]/g, '')
    .replace(/\s+/g, '-')
    .replace(/-+/g, '-');
}

type Variant = {
  sku: string;
  title: string;
  price_cents: number;
  compare_at_price_cents?: number | null;
  currency: string;
  stock_quantity: number;
};

type ImageLink = { media_id: number; sort_order: number };

type UpsertPayload = {
  slug: string;
  title: string;
  subtitle?: string | null;
  description?: string | null;
  category_id?: string | null;
  published: boolean;
  publish_at?: string | null;
  unpublish_at?: string | null;
  seo_title?: string | null;
  seo_description?: string | null;
  variants: Variant[];
  images: ImageLink[];
};

type ProductResponse = {
  product: {
    uuid_id: string;
    slug: string;
    title: string;
    subtitle?: string | null;
    description?: string | null;
    category_id?: string | null;
    published: boolean;
    publish_at?: string | null;
    unpublish_at?: string | null;
    seo_title?: string | null;
    seo_description?: string | null;
  };
  variants: Array<{
    sku: string;
    title: string;
    price_cents: number;
    compare_at_price_cents?: number | null;
    currency: string;
    stock_quantity: number;
  }>;
  images: Array<{ media_id: number; sort_order: number }>;
};

type Category = {
  uuid_id: string;
  slug: string;
  name: string;
  description?: string | null;
  parent_id?: string | null;
  sort_order: number;
};

type MediaListResponse = {
  items: Array<{
    id: number;
    path: string;
    url: string;
    mime_type: string;
  }>;
  nextCursor?: number;
};

export const ProductEditPage: React.FC = () => {
  return (
    <ErrorBoundary>
      <ProductEditPageInner />
    </ErrorBoundary>
  );
};

const ProductEditPageInner: React.FC = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const qc = useQueryClient();

  const isNew = id === 'new' || !id;

  // Debug logging
  
  const [title, setTitle] = useState('');
  const [slug, setSlug] = useState('');
  const [subtitle, setSubtitle] = useState('');
  const [description, setDescription] = useState('');
  const [published, setPublished] = useState(false);
  const [publishAt, setPublishAt] = useState<string | ''>('');
  const [unpublishAt, setUnpublishAt] = useState<string | ''>('');
  const [seoTitle, setSeoTitle] = useState('');
  const [seoDescription, setSeoDescription] = useState('');
  const [variants, setVariants] = useState<Variant[]>([
    { sku: '', title: 'Default', price_cents: 0, compare_at_price_cents: undefined, currency: 'INR', stock_quantity: 0 },
  ]);
  const [categoryId, setCategoryId] = useState<string>('');
  const [images, setImages] = useState<ImageLink[]>([]);
  const [showSuccess, setShowSuccess] = useState(false);
  const [successMessage, setSuccessMessage] = useState('');

  // All hooks must be called before any early returns
  const { data: prodData, isLoading: loadingProduct } = useQuery<ProductResponse>({
    queryKey: ['product', id],
    queryFn: () => {
      if (!id || id === 'new') throw new Error('Invalid product ID');
      return api.get<ProductResponse>(`/products/${id}`);
    },
    enabled: !isNew,
    staleTime: 2 * 60 * 1000, // 2 minutes
    retry: (failureCount: number, error: any) => {
      // Don't retry on 4xx errors or invalid ID
      if (error?.status >= 400 && error?.status < 500) return false;
      if (error?.message?.includes('Invalid product ID')) return false;
      return failureCount < 2;
    },
  });

  const { data: categories } = useQuery<{ items: Category[] }>({
    queryKey: ['categories'],
    queryFn: () => api.get<{ items: Category[] }>(`/categories`),
    staleTime: 10 * 60 * 1000, // 10 minutes
  });

  const { data: media } = useQuery<MediaListResponse>({
    queryKey: ['media', { first: 20 }],
    queryFn: () => api.get<MediaListResponse>(`/media?first=20`),
    staleTime: 5 * 60 * 1000, // 5 minutes
  });

  const saveMutation = useMutation({
    mutationFn: async () => {
      const payload: UpsertPayload = {
        slug: slug || slugify(title),
        title,
        subtitle: subtitle || null,
        description: description || null,
        published,
        publish_at: publishAt ? new Date(publishAt).toISOString() : null,
        unpublish_at: unpublishAt ? new Date(unpublishAt).toISOString() : null,
        seo_title: seoTitle || null,
        seo_description: seoDescription || null,
        variants,
        category_id: categoryId || null,
        images,
      };
      
            
      try {
        if (isNew) {
          const res = await api.post<{ uuid_id: string }>(`/products`, payload);
          return res.uuid_id;
        } else {
          await api.put(`/products/${id}`, payload);
          return Number(id);
        }
      } catch (err) {
        throw err;
      }
    },
    onSuccess: (newId) => {
      qc.invalidateQueries({ queryKey: ['products'] });
      if (isNew) {
        // Show success message first, then navigate after delay
        showSuccessMessage('Product added successfully!');
        setTimeout(() => {
          navigate('/products');
        }, 2000); // Navigate after 2 seconds
      } else {
        // Show success message for existing product updates
        showSuccessMessage('Product updated successfully!');
      }
    },
  });

  // Effects and computed values
  useEffect(() => {
    if (prodData && id !== 'new') {
      const p = prodData.product;
      setTitle(p.title || '');
      setSlug(p.slug || '');
      setSubtitle(p.subtitle || '');
      setDescription(p.description || '');
      setPublished(p.published);
      setPublishAt(p.publish_at || '');
      setUnpublishAt(p.unpublish_at || '');
      setSeoTitle(p.seo_title || '');
      setSeoDescription(p.seo_description || '');
      setVariants(
        (prodData.variants || []).map((v) => ({
          sku: v.sku || '',
          title: v.title || '',
          price_cents: v.price_cents || 0,
          compare_at_price_cents: v.compare_at_price_cents ?? undefined,
          currency: v.currency || 'INR',
          stock_quantity: v.stock_quantity || 0,
        })),
      );
      setCategoryId(p.category_id || '');
      setImages(prodData.images || []);
    }
  }, [prodData, isNew]);

  useEffect(() => {
    if (!isNew && !slug) {
      setSlug(slugify(title));
    }
  }, [title]);

  const canSave = useMemo(() => title.trim().length > 0 && slug.trim().length > 0, [title, slug]);

  // Show success message function
  const showSuccessMessage = (message: string) => {
    setSuccessMessage(message);
    setShowSuccess(true);
    setTimeout(() => {
      setShowSuccess(false);
    }, 3000); // Hide after 3 seconds
  };

  // Show error state for invalid product ID (only for existing products)
  const hasInvalidId = !isNew && (!id || id === 'undefined' || id === '' || id === null);
  
  // Show loading state while fetching data
  const shouldShowLoading = !isNew && loadingProduct;

  // Early returns after all hooks are defined
  if (hasInvalidId) {
    return (
      <div className="space-y-6">
        <header>
          <h1 className="text-2xl font-playfair text-maroon">Product Not Found</h1>
        </header>
        <div className="bg-red-50 border border-red-200 rounded-lg p-4">
          <p className="text-sm text-red-600 font-medium">Invalid product ID</p>
          <p className="text-xs text-red-500 mt-1">The product you're trying to edit doesn't exist.</p>
          <div className="mt-3 flex gap-2">
            <Link
              to="/products"
              className="text-xs bg-red-100 text-red-700 px-2 py-1 rounded hover:bg-red-200"
            >
              Back to Products
            </Link>
            <button
              onClick={() => window.location.reload()}
              className="text-xs bg-red-100 text-red-700 px-2 py-1 rounded hover:bg-red-200"
            >
              Retry
            </button>
          </div>
        </div>
      </div>
    );
  }

  // Show loading state while fetching data
  if (shouldShowLoading) {
    return (
      <div className="space-y-6">
        <header>
          <h1 className="text-2xl font-playfair text-maroon">Edit Product</h1>
        </header>
        <div className="flex items-center justify-center py-12">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-maroon"></div>
          <span className="ml-2 text-dark/70">Loading product...</span>
        </div>
      </div>
    );
  }

  function handleCategoryChange(categoryUuid: string) {
    setCategoryId(categoryUuid);
  }

  function addVariant() {
    setVariants((v) => [...v, { sku: '', title: '', price_cents: 0, compare_at_price_cents: undefined, currency: 'INR', stock_quantity: 0 }]);
  }

  function removeVariant(ix: number) {
    setVariants((v) => v.filter((_, i) => i !== ix));
  }

  function addImage(mediaId: number) {
    setImages((imgs) => {
      const nextOrder = imgs.length > 0 ? Math.max(...imgs.map((i) => i.sort_order)) + 1 : 0;
      if (imgs.some((i) => i.media_id === mediaId)) return imgs;
      return [...imgs, { media_id: mediaId, sort_order: nextOrder }];
    });
  }

  function removeImage(mediaId: number) {
    setImages((imgs) => imgs.filter((i) => i.media_id !== mediaId));
  }

  return (
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

      <header className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-playfair text-maroon">{isNew ? 'New Product' : 'Edit Product'}</h1>
          <p className="text-sm text-dark/70 mt-1">Title, variants, categories, and images.</p>
        </div>
        <button
          onClick={() => saveMutation.mutate()}
          disabled={!canSave || saveMutation.isPending}
          className="inline-flex items-center justify-center rounded-md bg-maroon text-cream text-sm font-medium px-3 py-2 hover:bg-maroon/90 disabled:opacity-60"
        >
          {saveMutation.isPending ? 'Saving…' : 'Save'}
        </button>
      </header>

      {shouldShowLoading && !isNew ? (
        <div className="flex items-center justify-center py-12">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-maroon"></div>
          <span className="ml-2 text-dark/70">Loading product...</span>
        </div>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Left: core fields */}
          <div className="lg:col-span-2 space-y-6">
            <div className="bg-white border border-gold/30 rounded-lg p-4 shadow-card space-y-3">
              <div>
                <label className="block text-sm font-medium mb-1">Title</label>
                <input value={title} onChange={(e) => setTitle(e.target.value)} className="w-full rounded-md border border-gold/40 bg-cream/60 px-3 py-2 text-sm" />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">Slug</label>
                <input value={slug} onChange={(e) => setSlug(e.target.value)} className="w-full rounded-md border border-gold/40 bg-cream/60 px-3 py-2 text-sm" />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">Subtitle</label>
                <input value={subtitle} onChange={(e) => setSubtitle(e.target.value)} className="w-full rounded-md border border-gold/40 bg-cream/60 px-3 py-2 text-sm" />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">Description</label>
                <textarea value={description} onChange={(e) => setDescription(e.target.value)} className="w-full rounded-md border border-gold/40 bg-cream/60 px-3 py-2 text-sm" rows={5} />
              </div>

              <div className="border-t pt-4">
                <h3 className="text-lg font-semibold mb-3">SEO Settings</h3>
                <div className="space-y-3">
                  <div>
                    <label className="block text-sm font-medium mb-1">SEO Title</label>
                    <input 
                      value={seoTitle} 
                      onChange={(e) => setSeoTitle(e.target.value)} 
                      placeholder="Custom SEO title for search engines"
                      className="w-full rounded-md border border-gold/40 bg-cream/60 px-3 py-2 text-sm" 
                    />
                    <p className="text-xs text-gray-500 mt-1">Optional: Custom title for search engines (60 characters recommended)</p>
                  </div>
                  <div>
                    <label className="block text-sm font-medium mb-1">SEO Description</label>
                    <textarea 
                      value={seoDescription} 
                      onChange={(e) => setSeoDescription(e.target.value)} 
                      placeholder="Custom SEO description for search engines"
                      className="w-full rounded-md border border-gold/40 bg-cream/60 px-3 py-2 text-sm" 
                      rows={3}
                    />
                    <p className="text-xs text-gray-500 mt-1">Optional: Description for search engines (160 characters recommended)</p>
                  </div>
                </div>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <label className="flex items-center gap-2 text-sm">
                  <input type="checkbox" checked={published} onChange={(e) => setPublished(e.target.checked)} />
                  Published
                </label>
                <div>
                  <label className="block text-sm mb-1">Publish At</label>
                  <input type="datetime-local" value={publishAt} onChange={(e) => setPublishAt(e.target.value)} className="w-full rounded-md border border-gold/40 bg-cream/60 px-3 py-2 text-sm" />
                </div>
                <div>
                  <label className="block text-sm mb-1">Unpublish At</label>
                  <input type="datetime-local" value={unpublishAt} onChange={(e) => setUnpublishAt(e.target.value)} className="w-full rounded-md border border-gold/40 bg-cream/60 px-3 py-2 text-sm" />
                </div>
              </div>
            </div>

            {/* Variants */}
            <div className="bg-white border border-gold/30 rounded-lg p-4 shadow-card">
              <div className="flex items-center justify-between mb-3">
                <h2 className="text-sm font-medium">Variants</h2>
                <button onClick={addVariant} className="text-maroon text-sm">Add</button>
              </div>
              <div className="overflow-x-auto">
                <table className="min-w-full text-sm">
                  <thead>
                    <tr className="bg-cream/60">
                      <th className="text-left px-3 py-2">SKU</th>
                      <th className="text-left px-3 py-2">Title</th>
                      <th className="text-left px-3 py-2">Price (cents)</th>
                      <th className="text-left px-3 py-2">Compare@ (cents)</th>
                      <th className="text-left px-3 py-2">Stock</th>
                      <th className="text-left px-3 py-2">Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {variants.map((v, i) => (
                      <tr key={i} className="border-t border-gold/20">
                        <td className="px-3 py-2"><input value={v.sku} onChange={(e) => setVariants((arr) => arr.map((x, ix) => ix===i? { ...x, sku: e.target.value } : x))} className="w-40 rounded-md border border-gold/40 bg-cream/60 px-2 py-1" /></td>
                        <td className="px-3 py-2"><input value={v.title} onChange={(e) => setVariants((arr) => arr.map((x, ix) => ix===i? { ...x, title: e.target.value } : x))} className="w-40 rounded-md border border-gold/40 bg-cream/60 px-2 py-1" /></td>
                        <td className="px-3 py-2"><input type="number" value={v.price_cents} onChange={(e) => setVariants((arr) => arr.map((x, ix) => ix===i? { ...x, price_cents: Number(e.target.value) } : x))} className="w-32 rounded-md border border-gold/40 bg-cream/60 px-2 py-1" /></td>
                        <td className="px-3 py-2"><input type="number" value={v.compare_at_price_cents ?? ''} onChange={(e) => setVariants((arr) => arr.map((x, ix) => ix===i? { ...x, compare_at_price_cents: e.target.value === '' ? undefined : Number(e.target.value) } : x))} className="w-32 rounded-md border border-gold/40 bg-cream/60 px-2 py-1" /></td>
                        <td className="px-3 py-2"><input type="number" value={v.stock_quantity} onChange={(e) => setVariants((arr) => arr.map((x, ix) => ix===i? { ...x, stock_quantity: Number(e.target.value) } : x))} className="w-24 rounded-md border border-gold/40 bg-cream/60 px-2 py-1" /></td>
                        <td className="px-3 py-2"><button onClick={() => removeVariant(i)} className="text-maroon">Remove</button></td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          </div>

          {/* Right: categories + images */}
          <div className="space-y-6">
            <div className="bg-white border border-gold/30 rounded-lg p-4 shadow-card">
              <h2 className="text-sm font-medium mb-3">Category</h2>
              {!categories ? (
                <div className="flex items-center py-4">
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-maroon mr-2"></div>
                  <span className="text-sm text-dark/70">Loading categories...</span>
                </div>
              ) : (
                <div>
                  <select 
                    value={categoryId} 
                    onChange={(e) => handleCategoryChange(e.target.value)}
                    className="w-full rounded-md border border-gold/40 bg-cream/60 px-3 py-2 text-sm"
                  >
                    <option value="">Select a category</option>
                    {(categories?.items || []).map((c) => (
                      <option key={c.uuid_id} value={c.uuid_id}>
                        {c.name}
                      </option>
                    ))}
                  </select>
                </div>
              )}
            </div>

            <div className="bg-white border border-gold/30 rounded-lg p-4 shadow-card">
              <h2 className="text-sm font-medium mb-3">Product Images</h2>
              
              {/* Upload new images */}
              <div className="mb-4">
                <ImageUploader 
                  type="product"
                  onUploaded={(result) => {
                    // Refetch media to show newly uploaded image
                    qc.refetchQueries({ queryKey: ['media'] });
                  }}
                />
              </div>
              
              {/* Existing media library */}
              <div className="border-t pt-3">
                <h3 className="text-xs text-dark/70 mb-2">Or select from media library</h3>
                {!media ? (
                <div className="flex items-center py-4">
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-maroon mr-2"></div>
                  <span className="text-sm text-dark/70">Loading media...</span>
                </div>
              ) : (
                <>
                  <div className="grid grid-cols-3 gap-2 max-h-64 overflow-y-auto">
                    {(media?.items || []).map((m) => (
                      <button key={m.id} type="button" onClick={() => addImage(m.id)} className="group border border-gold/30 rounded overflow-hidden">
                        <img src={getPublicImageUrl(m.url || m.path)} alt={m.path} className="w-full h-20 object-cover group-hover:opacity-80" />
                      </button>
                    ))}
                  </div>
                  {images.length > 0 && (
                    <div className="mt-3">
                      <h3 className="text-xs text-dark/70 mb-2">Selected</h3>
                      <div className="flex flex-wrap gap-2">
                        {images.map((im) => {
                          const mediaItem = media?.items.find((x) => x.id === im.media_id);
                          return (
                            <div key={im.media_id} className="relative border border-gold/30 rounded">
                              {mediaItem ? (
                                <img src={getPublicImageUrl(mediaItem.url || mediaItem.path)} alt={String(im.media_id)} className="w-20 h-20 object-cover rounded" />
                              ) : (
                                <div className="w-20 h-20 bg-cream" />
                              )}
                              <button onClick={() => removeImage(im.media_id)} className="absolute top-1 right-1 text-xs bg-maroon text-cream rounded px-1">x</button>
                            </div>
                          );
                        })}
                      </div>
                    </div>
                  )}
                </>
              )}
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
