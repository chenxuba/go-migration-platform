import { defineConfig } from 'vite';

export default defineConfig({
  base: '/training-games/',
  server: {
    host: '0.0.0.0',
    port: 6677,
  },
  preview: {
    host: '0.0.0.0',
    port: 6777,
  },
});
