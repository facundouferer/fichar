import { defineConfig } from 'astro/config';

export default defineConfig({
  preview: {
    allowedHosts: ['fichar.gar.com.ar', 'www.fichar.gar.com.ar']
  },
  vite: {
    build: {
      outDir: 'dist',
      emptyOutDir: true
    },
    define: {
      'import.meta.env.PUBLIC_API_URL': JSON.stringify(process.env.PUBLIC_API_URL || process.env.API_URL || 'http://localhost:8082')
    }
  }
});
