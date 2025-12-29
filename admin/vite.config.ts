import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react-swc';

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5174,
    proxy: {
      '/api/admin': {
        target: 'https://etreasure-1.onrender.com',
        changeOrigin: true,
      },
      '/api/cart': {
        target: 'https://etreasure-1.onrender.com',
        changeOrigin: true,
      },
      '/api/wishlist': {
        target: 'https://etreasure-1.onrender.com',
        changeOrigin: true,
      },
      '/uploads': {
        target: 'https://etreasure-1.onrender.com',
        changeOrigin: true,
      },
    },
  },
  build: {
    sourcemap: false,
  },
});
