// @ts-check
import { defineConfig } from 'astro/config';

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

    // ✅ Dev-only API proxy (ignored in build)
    server: {
      proxy: {
        '/api': {
          target: 'http://localhost:8080',
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
