<template>
  <div class="app-layout">
    <!-- Header -->
    <header class="app-header">
      <div class="header-content">
        <div class="logo">
          <h1>G-Backup</h1>
        </div>
        
        <nav class="main-nav">
          <router-link to="/dashboard" class="nav-item">
            <span class="nav-text">Dashboard</span>
          </router-link>
          
          <router-link to="/remotes" class="nav-item">
            <span class="nav-text">Gdrive</span>
          </router-link>
          
          <router-link to="/scheduled" class="nav-item">
            <span class="nav-text">Scheduled Jobs</span>
          </router-link>
          
          <router-link to="/manual" class="nav-item">
            <span class="nav-text">Jobs</span>
          </router-link>
        </nav>

        <div class="header-actions">
          <button @click="handleLogout" class="logout-btn">
            Logout
          </button>
        </div>
      </div>
    </header>
    
    <!-- Main Content dengan Log Panel -->
    <div class="content-wrapper">
      <!-- Main Content -->
      <main class="app-main">
        <transition name="fade" mode="out-in">
          <router-view />
        </transition>
      </main>
      
      <!-- Log Panel (Sticky di kanan) -->
      <aside class="log-panel">
        <SimpleLiveLog />
      </aside>
    </div>
    
    <!-- Footer -->
    <footer class="app-footer">
      <p>© 2024 G-Backup System</p>
    </footer>

    <!-- ⭐ Logout Confirmation Modal -->
    <Teleport to="body">
      <Transition name="fade">
        <div v-if="showLogoutModal" class="modal-overlay" @click="showLogoutModal = false">
          <div class="modal-content" @click.stop>
            <div class="modal-header">
              <h3>Konfirmasi Logout</h3>
            </div>
            <div class="modal-body">
              <div class="warning-icon">⚠️</div>
              <p>Yakin ingin keluar dari sistem? Sesi Anda akan diakhiri sekarang.</p>
            </div>
            <div class="modal-footer">
              <button @click="showLogoutModal = false" class="btn-cancel">
                Batal
              </button>
              <button @click="confirmLogout" class="btn-confirm">
                Ya, Keluar
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { useAuthStore } from '@/stores/authStore';
import SimpleLiveLog from '@/components/LiveLog.vue';

const router = useRouter();
const authStore = useAuthStore();

// State untuk mengontrol visibilitas modal
const showLogoutModal = ref(false);

const handleLogout = () => {
  // Buka modal alih-alih menggunakan confirm() browser
  showLogoutModal.value = true;
};

const confirmLogout = () => {
  authStore.logout();
  showLogoutModal.value = false;
  router.push('/login');
};
</script>

<style scoped>
.app-layout {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background-color: #fafafa;
}

/* Header */
.app-header {
  background: #fff;
  border-bottom: 1px solid #e5e5e5;
  position: sticky;
  top: 0;
  z-index: 100;
}

.header-content {
  max-width: 100%;
  margin: 0 auto;
  padding: 1rem 2rem;
  display: flex;
  align-items: center;
  gap: 2rem;
}

.logo h1 {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 600;
  color: #1a1a1a;
  white-space: nowrap;
  letter-spacing: -0.02em;
}

.main-nav {
  display: flex;
  gap: 0.25rem;
  flex: 1;
  overflow-x: auto;
}

.nav-item {
  display: flex;
  align-items: center;
  padding: 0.625rem 1rem;
  color: #666;
  text-decoration: none;
  border-radius: 6px;
  transition: all 0.2s ease;
  white-space: nowrap;
  font-weight: 500;
  font-size: 0.9375rem;
}

.nav-item:hover {
  background: #f5f5f5;
  color: #1a1a1a;
}

.nav-item.router-link-active {
  background: #1a1a1a;
  color: white;
  font-weight: 500;
}

.nav-text {
  font-size: 0.9375rem;
}

.header-actions {
  display: flex;
  align-items: center;
}

.logout-btn {
  display: flex;
  align-items: center;
  padding: 0.625rem 1rem;
  background: transparent;
  color: #666;
  border: 1px solid #e5e5e5;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 500;
  font-size: 0.9375rem;
  transition: all 0.2s ease;
  white-space: nowrap;
}

.logout-btn:hover {
  background: #fef2f2;
  color: #dc2626;
  border-color: #dc2626;
}

/* Content Wrapper dengan Log Panel */
.content-wrapper {
  flex: 1;
  display: grid;
  grid-template-columns: 1fr 320px;
  gap: 1.5rem;
  padding: 1.5rem 2rem;
  max-width: 100%;
}

/* Main Content */
.app-main {
  min-width: 0;
}

/* Log Panel (Sticky) */
.log-panel {
  position: sticky;
  top: calc(60px + 1.5rem);
  height: calc(100vh - 60px - 3rem - 60px);
  min-width: 0;
}

/* Footer */
.app-footer {
  background: #fff;
  border-top: 1px solid #e5e5e5;
  padding: 1.25rem 2rem;
  text-align: center;
}

.app-footer p {
  margin: 0;
  font-size: 0.875rem;
  color: #666;
}

/* Transitions */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* Responsive - Hide log panel on smaller screens */
@media (max-width: 1200px) {
  .content-wrapper {
    grid-template-columns: 1fr;
  }
  
  .log-panel {
    display: none;
  }
}

/* Logout Confirmation Modal Styles */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
  backdrop-filter: blur(2px);
}

.modal-content {
  background: white;
  border-radius: 12px;
  width: 100%;
  max-width: 400px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.2);
  overflow: hidden;
  animation: modalSlide 0.3s ease-out;
}

@keyframes modalSlide {
  from { transform: translateY(20px); opacity: 0; }
  to { transform: translateY(0); opacity: 1; }
}

.modal-header {
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid #f0f0f0;
}

.modal-header h3 {
  margin: 0;
  font-size: 1.1rem;
  font-weight: 600;
  color: #1a1a1a;
}

.modal-body {
  padding: 2rem 1.5rem;
  text-align: center;
}

.warning-icon {
  font-size: 2.5rem;
  margin-bottom: 1rem;
}

.modal-body p {
  margin: 0;
  color: #4b5563;
  font-size: 0.95rem;
  line-height: 1.5;
}

.modal-footer {
  padding: 1rem 1.5rem;
  background: #f9fafb;
  border-top: 1px solid #f0f0f0;
  display: flex;
  gap: 0.75rem;
  justify-content: flex-end;
}

.btn-cancel {
  padding: 0.625rem 1.25rem;
  background: white;
  color: #374151;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-cancel:hover {
  background: #f3f4f6;
  border-color: #9ca3af;
}

.btn-confirm {
  padding: 0.625rem 1.25rem;
  background: #dc2626;
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 0.875rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-confirm:hover {
  background: #b91c1c;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(220, 38, 38, 0.2);
}

@media (max-width: 1024px) {
  .header-content {
    flex-direction: column;
    gap: 1rem;
    align-items: stretch;
  }
  
  .main-nav {
    width: 100%;
    justify-content: center;
  }

  .header-actions {
    width: 100%;
    justify-content: center;
  }

  .logout-btn {
    width: 100%;
  }
}
</style>