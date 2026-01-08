<template>
  <nav class="navbar">
    <div class="navbar-header">
      <h3>G-Backup</h3>
    </div>
    
    <ul class="nav-links">
      <li>
        <router-link to="/remotes">Monitoring Remote</router-link>
      </li>
      <li>
        <router-link to="/scheduled-jobs">Scheduled Jobs</router-link>
      </li>
      <li>
        <router-link to="/manual-jobs">Manual Jobs</router-link>
      </li>
      </ul>
    
    <div class="navbar-footer">
      <button @click="handleLogout" class="logout-button">Logout</button>
    </div>
  </nav>
</template>

<script setup>
import { useAuthStore } from '@/stores/authStore'
import { useRouter } from 'vue-router'
import authService from '@/services/authService' // Import authService untuk logout

const authStore = useAuthStore()
const router = useRouter()

// Fungsi Logout
function handleLogout() {
  authService.logout()
  // Router akan otomatis mengarahkan ke /login
  // (Jika authService tidak mengarahkan, tambahkan router.push('/login'))
}
</script>

<style scoped>
/* Styling sederhana untuk sidebar */
.navbar {
  width: 240px;
  height: 100vh;
  background: #2c3e50; /* Warna gelap */
  color: white;
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  flex-shrink: 0; /* Mencegah sidebar menyusut */
}
.navbar-header h3 {
  text-align: center;
  margin: 0 0 2rem 0;
  font-size: 1.5rem;
}
.nav-links {
  list-style: none;
  padding: 0;
  margin: 0;
  flex-grow: 1; /* Mendorong footer ke bawah */
}
.nav-links li {
  margin-bottom: 0.5rem;
}
.nav-links a {
  color: #bdc3c7; /* Warna abu-abu terang */
  text-decoration: none;
  display: block;
  padding: 0.75rem 1rem;
  border-radius: 4px;
  transition: background-color 0.2s;
}
.nav-links a:hover {
  background-color: #34495e;
}
/* Style untuk link yang sedang aktif */
.nav-links a.router-link-exact-active {
  background-color: #1abc9c; /* Warna aksen */
  color: white;
  font-weight: bold;
}
.navbar-footer {
  margin-top: auto;
}
.logout-button {
  background-color: #e74c3c;
  color: white;
  border: none;
  padding: 0.75rem;
  cursor: pointer;
  width: 100%;
  border-radius: 4px;
}
.logout-button:hover {
  background-color: #c0392b;
}
</style>