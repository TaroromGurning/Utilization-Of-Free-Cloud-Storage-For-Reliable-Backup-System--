import axios from 'axios'
import { useAuthStore } from '@/stores/authStore'

// Membuat instance Axios
const apiClient = axios.create({
  // baseURL akan otomatis menggunakan proxy Vite (ke /api)
  baseURL: '/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
})

// --- AXIOS INTERCEPTOR (PENTING) ---
// Ini berjalan sebelum setiap request API dikirim
apiClient.interceptors.request.use(
  (config) => {
    const authStore = useAuthStore()
    const token = authStore.token

    // Jika token ada, tambahkan ke header Authorization
    if (token) {
      config.headers['Authorization'] = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

export default apiClient