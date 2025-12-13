# Image Optimization Guide

This directory contains optimized images for Ethnic Treasure website. All images follow the performance optimization strategy:

## Image Format Priority
1. **AVIF** - Primary format (best compression, modern browsers)
2. **WebP** - Fallback format (good compression, wide support)
3. **JPG** - Final fallback (maximum compatibility)

## Responsive Image Sizes

### Hero Banner
- `hero-banner-400.avif/webp/jpg` - 400px wide (mobile)
- `hero-banner-800.avif/webp/jpg` - 800px wide (tablet)
- `hero-banner-1600.avif/webp/jpg` - 1600px wide (desktop)

### Product Images
- `{product-name}-400.avif/webp/jpg` - 400px wide (thumbnails)
- `{product-name}-600.avif/webp/jpg` - 600px wide (cards)
- `{product-name}-800.avif/webp/jpg` - 800px wide (gallery)
- `{product-name}-1200.avif/webp/jpg` - 1200px wide (main view)

### Category Images
- `{category}-400.avif/webp/jpg` - 400px wide (cards)
- `{category}-800.avif/webp/jpg` - 800px wide (hover states)

### Gallery Images
- `instagram-{id}-400.avif/webp/jpg` - 400px square

## Usage Example

```html
<picture>
  <source srcset="/images/product-1200.avif" media="(min-width: 1024px)" type="image/avif">
  <source srcset="/images/product-800.avif" media="(min-width: 640px)" type="image/avif">
  <source srcset="/images/product-400.avif" type="image/avif">
  <source srcset="/images/product-1200.webp" media="(min-width: 1024px)" type="image/webp">
  <source srcset="/images/product-800.webp" media="(min-width: 640px)" type="image/webp">
  <source srcset="/images/product-400.webp" type="image/webp">
  <img 
    src="/images/product-800.jpg" 
    alt="Product description"
    class="w-full h-full object-cover"
    loading="lazy"
    sizes="(max-width: 600px) 100vw, 33vw"
  />
</picture>
```

## Optimization Settings

### AVIF Settings
- Quality: 80-85%
- Effort: 6 (balanced speed/quality)
- Chroma subsampling: 4:2:0

### WebP Settings
- Quality: 85%
- Method: 6 (balanced speed/quality)
- Alpha quality: 90%

### JPG Settings
- Quality: 85%
- Progressive: enabled
- Optimized: enabled

## File Naming Convention

```
{section}-{name}-{size}.format
```

Examples:
- `hero-banner-1600.avif`
- `products-banarasi-silk-800.webp`
- `categories-sarees-400.avif`
- `gallery-instagram-1-400.webp`

## Alt Text Guidelines

All images must have descriptive alt text that:
- Describes the content for visually impaired users
- Includes relevant keywords naturally
- Is concise but informative
- Follows accessibility best practices

## Preloading Critical Images

Critical above-the-fold images should be preloaded:

```html
<link rel="preload" as="image" href="/images/hero-banner-1600.avif" type="image/avif">
```

## Lazy Loading

All non-hero images should use lazy loading:

```html
<img loading="lazy" ...>
```

## Image CDN Configuration

When deployed to Cloudflare Pages:
- Enable automatic WebP conversion
- Configure AVIF delivery for supported browsers
- Set appropriate cache headers (1 year immutable)
- Enable automatic compression
