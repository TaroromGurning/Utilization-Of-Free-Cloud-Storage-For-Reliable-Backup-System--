import apiClient from './api';
import { useAuthStore } from '@/stores/authStore';
import router from '@/router';

export default {
    /**
   * Mengirim kredensial ke backend, menyimpan token, dan mengarahkan ke dashboard.
   * @param {string} username
   * @param {string} password
   */
  async login(username, password) {
    const authStore = useAuthStore()

    try {
        const response = await apiClient.post('auth/login', {
            username: username,
            password: password
        })

        const token = response.data.token
        if (token) {
            authStore.setToken(token)
            router.push('/Dashboard')
        }   
    } catch (error) {
        console.error("Login Failed:", error)
        throw error
    }
  },
  /**
   * menghapus token dari storage -> halaman login
   */
  logout() {
    const authStore = useAuthStore()
    authStore.clearToken()
    router.push('/login')
  }
}