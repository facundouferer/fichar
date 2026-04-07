import { defineConfig } from 'astro/config';

export default defineConfig({
  vite: {
    build: {
      outDir: 'dist',
      emptyOutDir: true
    }
  }
});
