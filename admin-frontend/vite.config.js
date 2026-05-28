import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  base: '/admin-ui/',
  server: {
    proxy: {
      '/admin/': 'http://localhost:8000'
    }
  }
})
