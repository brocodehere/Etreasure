import React, { useState } from 'react';
import { useMutation } from '@tanstack/react-query';
import { api } from '../lib/api';

interface PreviewRequest {
  type: 'product' | 'banner' | 'category' | 'offer';
  content: Record<string, any>;
}

export function PreviewPage() {
  const [type, setType] = useState<'product' | 'banner' | 'category' | 'offer'>('product');
  const [content, setContent] = useState<Record<string, any>>({
    title: '',
    description: '',
    price: 0,
    currency: 'USD',
    image_url: '',
  });
  const [previewHtml, setPreviewHtml] = useState('');

  const previewMutation = useMutation({
    mutationFn: (payload: PreviewRequest) =>
      api.post<{ html: string }>('/preview', payload).then((r: any) => r.data),
    onSuccess: (data: any) => {
      setPreviewHtml(data.html);
    },
  });

  const handlePreview = () => {
    previewMutation.mutate({ type, content });
  };

  const handleFieldChange = (field: string, value: any) => {
    setContent(prev => ({ ...prev, [field]: value }));
  };

  const renderForm = () => {
    switch (type) {
      case 'product':
        return (
          <>
            <div>
              <label className="block text-sm font-medium mb-1">Title</label>
              <input
                type="text"
                value={content.title || ''}
                onChange={(e) => handleFieldChange('title', e.target.value)}
                className="w-full border rounded px-3 py-2"
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">Description</label>
              <textarea
                value={content.description || ''}
                onChange={(e) => handleFieldChange('description', e.target.value)}
                rows={3}
                className="w-full border rounded px-3 py-2"
              />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium mb-1">Price</label>
                <input
                  type="number"
                  step="any"
                  value={content.price || ''}
                  onChange={(e) => handleFieldChange('price', parseFloat(e.target.value) || 0)}
                  className="w-full border rounded px-3 py-2"
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">Currency</label>
                <select
                  value={content.currency || 'USD'}
                  onChange={(e) => handleFieldChange('currency', e.target.value)}
                  className="w-full border rounded px-3 py-2"
                >
                  <option value="USD">USD</option>
                  <option value="EUR">EUR</option>
                  <option value="GBP">GBP</option>
                </select>
              </div>
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">Image URL</label>
              <input
                type="url"
                value={content.image_url || ''}
                onChange={(e) => handleFieldChange('image_url', e.target.value)}
                placeholder="https://example.com/image.jpg"
                className="w-full border rounded px-3 py-2"
              />
            </div>
          </>
        );
      case 'banner':
        return (
          <>
            <div>
              <label className="block text-sm font-medium mb-1">Title</label>
              <input
                type="text"
                value={content.title || ''}
                onChange={(e) => handleFieldChange('title', e.target.value)}
                className="w-full border rounded px-3 py-2"
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">Subtitle</label>
              <input
                type="text"
                value={content.subtitle || ''}
                onChange={(e) => handleFieldChange('subtitle', e.target.value)}
                className="w-full border rounded px-3 py-2"
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">Image URL</label>
              <input
                type="url"
                value={content.image_url || ''}
                onChange={(e) => handleFieldChange('image_url', e.target.value)}
                placeholder="https://example.com/banner.jpg"
                className="w-full border rounded px-3 py-2"
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">Link URL</label>
              <input
                type="url"
                value={content.link_url || ''}
                onChange={(e) => handleFieldChange('link_url', e.target.value)}
                placeholder="https://example.com/products"
                className="w-full border rounded px-3 py-2"
              />
            </div>
          </>
        );
      case 'category':
        return (
          <>
            <div>
              <label className="block text-sm font-medium mb-1">Name</label>
              <input
                type="text"
                value={content.name || ''}
                onChange={(e) => handleFieldChange('name', e.target.value)}
                className="w-full border rounded px-3 py-2"
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">Description</label>
              <textarea
                value={content.description || ''}
                onChange={(e) => handleFieldChange('description', e.target.value)}
                rows={3}
                className="w-full border rounded px-3 py-2"
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">Image URL</label>
              <input
                type="url"
                value={content.image_url || ''}
                onChange={(e) => handleFieldChange('image_url', e.target.value)}
                placeholder="https://example.com/category.jpg"
                className="w-full border rounded px-3 py-2"
              />
            </div>
          </>
        );
      case 'offer':
        return (
          <>
            <div>
              <label className="block text-sm font-medium mb-1">Title</label>
              <input
                type="text"
                value={content.title || ''}
                onChange={(e) => handleFieldChange('title', e.target.value)}
                className="w-full border rounded px-3 py-2"
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">Description</label>
              <textarea
                value={content.description || ''}
                onChange={(e) => handleFieldChange('description', e.target.value)}
                rows={3}
                className="w-full border rounded px-3 py-2"
              />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium mb-1">Discount Type</label>
                <select
                  value={content.discount_type || 'percentage'}
                  onChange={(e) => handleFieldChange('discount_type', e.target.value)}
                  className="w-full border rounded px-3 py-2"
                >
                  <option value="percentage">Percentage</option>
                  <option value="fixed">Fixed Amount</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">Discount Value</label>
                <input
                  type="number"
                  step="any"
                  value={content.discount_value || ''}
                  onChange={(e) => handleFieldChange('discount_value', parseFloat(e.target.value) || 0)}
                  className="w-full border rounded px-3 py-2"
                />
              </div>
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">Image URL</label>
              <input
                type="url"
                value={content.image_url || ''}
                onChange={(e) => handleFieldChange('image_url', e.target.value)}
                placeholder="https://example.com/offer.jpg"
                className="w-full border rounded px-3 py-2"
              />
            </div>
          </>
        );
      default:
        return null;
    }
  };

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold mb-6">Preview</h1>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div>
          <div className="mb-4">
            <label className="block text-sm font-medium mb-1">Preview Type</label>
            <select
              value={type}
              onChange={(e) => setType(e.target.value as any)}
              className="w-full border rounded px-3 py-2"
            >
              <option value="product">Product</option>
              <option value="banner">Banner</option>
              <option value="category">Category</option>
              <option value="offer">Offer</option>
            </select>
          </div>

          <div className="space-y-4">
            {renderForm()}
          </div>

          <button
            onClick={handlePreview}
            disabled={previewMutation.isPending}
            className="mt-6 bg-gold text-white px-4 py-2 rounded hover:bg-yellow-600 transition disabled:opacity-50"
          >
            {previewMutation.isPending ? 'Generating...' : 'Preview'}
          </button>
        </div>

        <div>
          <h2 className="text-lg font-semibold mb-4">Preview Output</h2>
          {previewHtml ? (
            <div className="border rounded-lg p-4 bg-gray-50">
              <div dangerouslySetInnerHTML={{ __html: previewHtml }} />
            </div>
          ) : (
            <div className="border rounded-lg p-4 bg-gray-50 text-gray-500">
              Configure content and click Preview to see the rendered output.
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
