import vue from '@vitejs/plugin-vue'
import { defineConfig } from 'vite'

export default defineConfig({
  base: './',
  plugins: [vue()],
  server: {
    port: 6699,
  },
  build: {
    outDir: 'dist',
    assetsDir: 'assets',
  },
})
