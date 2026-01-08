import { fileURLToPath, URL } from 'node:url' 
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue()],

  resolve: {
    alias: {
      // Mengatur '@' agar menunjuk ke direktori /src
      '@': fileURLToPath(new URL('./src', import.meta.url)) 
    }
  },

  server: {
    // Port untuk frontend (misalnya, 5173)
    host: '0.0.0.0',
    port: 5173, 
    proxy: {
      // Semua request yang dimulai dengan /api
      '/api': {
        // Akan diteruskan ke backend Golang Anda
        target: 'http://localhost:8080', // Ganti jika IP backend/port berbeda
        changeOrigin: true, // Wajib untuk CORS
      }
    }
  }
})