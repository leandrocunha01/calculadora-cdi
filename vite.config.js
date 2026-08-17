import path from 'path'
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  root: 'web',
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './web')
    }
  },
  server: {
    proxy: {
      '/simulate': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  }
})
