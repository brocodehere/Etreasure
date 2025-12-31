import type { APIRoute } from 'astro';

// Static pages that should always be included
const staticPages = [
  {
    url: '/',
    changefreq: 'daily',
    priority: 1.0,
    lastmod: new Date().toISOString()
  },
  {
    url: '/shop',
    changefreq: 'daily',
    priority: 0.9,
    lastmod: new Date().toISOString()
  },
  {
    url: '/categories',
    changefreq: 'weekly',
    priority: 0.8,
    lastmod: new Date().toISOString()
  },
  {
    url: '/about',
    changefreq: 'monthly',
    priority: 0.7,
    lastmod: new Date().toISOString()
  },
  {
    url: '/contact',
    changefreq: 'monthly',
    priority: 0.7,
    lastmod: new Date().toISOString()
  },
  {
    url: '/bestsellers',
    changefreq: 'weekly',
    priority: 0.8,
    lastmod: new Date().toISOString()
  }
];

export const GET: APIRoute = async ({ site }) => {
  if (!site) {
    return new Response('Site configuration missing', { status: 500 });
  }

  const baseUrl = site.origin;
  const sitemap: string[] = [];

  // Add XML header
  sitemap.push('<?xml version="1.0" encoding="UTF-8"?>');
  sitemap.push('<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">');

  // Add static pages
  for (const page of staticPages) {
    sitemap.push('<url>');
    sitemap.push(`<loc>${baseUrl}${page.url}</loc>`);
    sitemap.push(`<lastmod>${page.lastmod}</lastmod>`);
    sitemap.push(`<changefreq>${page.changefreq}</changefreq>`);
    sitemap.push(`<priority>${page.priority}</priority>`);
    sitemap.push('</url>');
  }

  try {
    // Fetch dynamic data from API
    const API_URL = import.meta.env.PUBLIC_API_URL || 'http://localhost:8080';
    
    // Fetch products
    const productsResponse = await fetch(`${API_URL}/api/products?limit=1000`);
    if (productsResponse.ok) {
      const productsData = await productsResponse.json();
      const products = productsData.items || [];
      
      // Add product pages
      for (const product of products) {
        const productSlug = product.slug || product.id || product.uuid_id;
        if (productSlug) {
          sitemap.push('<url>');
          sitemap.push(`<loc>${baseUrl}/product/${productSlug}</loc>`);
          sitemap.push(`<lastmod>${product.updated_at || new Date().toISOString()}</lastmod>`);
          sitemap.push('<changefreq>weekly</changefreq>');
          sitemap.push('<priority>0.8</priority>');
          sitemap.push('</url>');
        }
      }
    }

    // Fetch categories
    const categoriesResponse = await fetch(`${API_URL}/api/public/categories`);
    if (categoriesResponse.ok) {
      const categoriesData = await categoriesResponse.json();
      const categories = categoriesData.items || [];
      
      // Add category pages
      for (const category of categories) {
        const categorySlug = category.slug || category.name?.toLowerCase().replace(/\s+/g, '-');
        if (categorySlug) {
          sitemap.push('<url>');
          sitemap.push(`<loc>${baseUrl}/shop/${categorySlug}</loc>`);
          sitemap.push(`<lastmod>${category.updated_at || new Date().toISOString()}</lastmod>`);
          sitemap.push('<changefreq>weekly</changefreq>');
          sitemap.push('<priority>0.7</priority>');
          sitemap.push('</url>');
        }
      }
    }

  } catch (error) {
        // Continue with static pages only if API fails
  }

  sitemap.push('</urlset>');

  return new Response(sitemap.join('\n'), {
    headers: {
      'Content-Type': 'application/xml',
      'Cache-Control': 'public, max-age=3600, s-maxage=3600'
    }
  });
};
