import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  root: 'web',
  server: {
    proxy: {
      '/simulate': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  }
})
