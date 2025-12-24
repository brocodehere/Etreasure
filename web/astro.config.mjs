// @ts-check
import { defineConfig } from 'astro/config';
import node from '@astrojs/node';
import tailwind from '@astrojs/tailwind';
import react from '@astrojs/react';

// https://astro.build/config
export default defineConfig({
  integrations: [
    tailwind(),
    react()
  ],





  // ✅ Keep output static (best for performance + Vercel/Cloudflare)
  output: 'static',

  // ✅ Vite config (safe + minimal)
  vite: {
  resolve: {
    alias: {
      '@lib': '/src/lib',
      '@components': '/src/components',
    },
  },
  optimizeDeps: {
    force: false,  // Changed from true to false
    include: ['react', 'react-dom'],
  },
  esbuild: {
    target: 'es2015',  // Add this line
  },
  build: {
    rollupOptions: {
      external: [],  // Add this line
    },
  },
  server: {
    proxy: {
      '/api': {
        target: 'https://etreasure-1.onrender.com',
        changeOrigin: true,
        secure: false,
      },
    },
  },
},
  // ✅ Image service (safe)
  image: {
    service: {
      entrypoint: 'astro/assets/services/sharp',
      config: {
        limitInputPixels: false,
      },
    },
  },

  // ✅ Good for SEO & performance
  compressHTML: true,
});
