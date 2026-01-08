<template>
  <div class="setup-container">
    <div class="setup-card">
      <!-- Header -->
      <div class="setup-header">
        <h1>Setup Awal</h1>
        <p>Buat akun administrator untuk memulai sistem</p>
      </div>
      
      <!-- Form -->
      <form @submit.prevent="handleSubmit" class="setup-form">
        <div class="form-group">
          <label for="username">Username Administrator</label>
          <input 
            type="text" 
            id="username" 
            v-model="username" 
            placeholder="Masukkan username"
            required 
            :disabled="loading"
            autocomplete="username"
          >
        </div>

        <div class="form-group">
          <label for="password">Password</label>
          <input 
            type="password" 
            id="password" 
            v-model="password" 
            placeholder="Minimal 6 karakter"
            required 
            :disabled="loading"
            autocomplete="new-password"
          >
        </div>
        
        <div v-if="error" class="error-message">
          {{ error }}
        </div>
        
        <button type="submit" :disabled="loading" class="btn-submit">
          <span v-if="loading" class="loading-spinner"></span>
          {{ loading ? 'Memproses...' : 'Buat Akun Administrator' }}
        </button>

        <p class="info-text">Setelah setup selesai, Anda akan diarahkan ke halaman login</p>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { useAppStore } from '@/stores/app';
import { useRouter } from 'vue-router';

const appStore = useAppStore();
const router = useRouter();

const username = ref('');
const password = ref('');
const error = ref(null);
const loading = ref(false);

const handleSubmit = async () => {
  error.value = null;
  loading.value = true;
  
  try {
    const response = await fetch('/api/v1/setup/register', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ 
        username: username.value, 
        password: password.value 
      }),
    });

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || 'Gagal mendaftar admin pertama.');
    }

    // Setup complete
    appStore.setSetupComplete();
    
    console.log('✅ Setup completed successfully');
    
    // Redirect ke login
    router.push({ 
      path: '/login', 
      query: { setup: 'complete' } 
    });
    
  } catch (err) {
    error.value = err.message;
    console.error('❌ Setup error:', err);
  } finally {
    loading.value = false;
  }
};
</script>

<style scoped>
.setup-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #fafafa;
  padding: 1rem;
}

.setup-card {
  background: white;
  border: 1px solid #e5e5e5;
  border-radius: 8px;
  padding: 3rem 2.5rem;
  width: 100%;
  max-width: 480px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

.setup-header {
  text-align: center;
  margin-bottom: 2.5rem;
}

.setup-header h1 {
  font-size: 1.75rem;
  font-weight: 600;
  color: #1a1a1a;
  margin: 0 0 0.75rem 0;
  letter-spacing: -0.02em;
}

.setup-header p {
  font-size: 0.9375rem;
  color: #666;
  margin: 0;
  line-height: 1.5;
}

.setup-form {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-group label {
  font-size: 0.875rem;
  font-weight: 500;
  color: #333;
}

.form-group input {
  padding: 0.75rem 1rem;
  font-size: 0.9375rem;
  border: 1px solid #e5e5e5;
  border-radius: 6px;
  transition: all 0.2s;
  background: white;
}

.form-group input:focus {
  outline: none;
  border-color: #1a1a1a;
  box-shadow: 0 0 0 3px rgba(26, 26, 26, 0.05);
}

.form-group input:disabled {
  background: #f8f8f8;
  cursor: not-allowed;
  color: #999;
}

.form-group input::placeholder {
  color: #999;
}

.error-message {
  background: #fef2f2;
  color: #dc2626;
  padding: 0.875rem 1rem;
  border-radius: 6px;
  font-size: 0.875rem;
  border: 1px solid #fee2e2;
  margin: 0;
}

.btn-submit {
  background: #1a1a1a;
  color: white;
  border: none;
  padding: 0.875rem 1.5rem;
  border-radius: 6px;
  font-size: 0.9375rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  margin-top: 0.5rem;
}

.btn-submit:hover:not(:disabled) {
  background: #333;
}

.btn-submit:disabled {
  background: #e5e5e5;
  color: #999;
  cursor: not-allowed;
}

.loading-spinner {
  width: 16px;
  height: 16px;
  border: 2px solid #fff;
  border-top-color: transparent;
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.info-text {
  text-align: center;
  font-size: 0.8125rem;
  color: #666;
  margin: 0.5rem 0 0 0;
  line-height: 1.5;
}

@media (max-width: 540px) {
  .setup-card {
    padding: 2rem 1.5rem;
  }

  .setup-header h1 {
    font-size: 1.5rem;
  }

  .setup-header p {
    font-size: 0.875rem;
  }
}
</style>