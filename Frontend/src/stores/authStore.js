import { defineStore } from "pinia";
import { ref } from "vue";

export const useAuthStore = defineStore('auth', () => {
    const token = ref(sessionStorage.getItem('token') || null)

    function isAuthenticated() {
        return token.value !== null
    }

    function setToken(newToken) {
        token.value = newToken
        sessionStorage.setItem('token', newToken)
    }

    function clearToken() {
        token.value = null
        sessionStorage.removeItem('token')
    }

    // ⭐ BARU: Logout function
    function logout() {
        token.value = null
        sessionStorage.removeItem('token')
        console.log('✅ User logged out')
    }

    return {
        token, 
        isAuthenticated, 
        setToken, 
        clearToken, 
        logout  // ⭐ Export logout
    }
})