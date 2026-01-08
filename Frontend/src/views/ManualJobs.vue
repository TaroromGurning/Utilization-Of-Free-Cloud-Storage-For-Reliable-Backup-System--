<template>
  <div class="manual-jobs-view">
    <div class="header">
      <div>
        <h1>Manual Jobs</h1>
        <p class="subtitle">View and manage all manual backup jobs</p>
      </div>  
    </div>

    <!-- Loading State -->
    <div v-if="isLoading" class="status-message">
      <span class="loading-dot"></span>
      Memuat data...
    </div>
    
    <!-- Error State -->
    <div v-if="errorMessage" class="status-message error">
      {{ errorMessage }}
      <button @click="fetchData" class="retry-btn">Coba Lagi</button>
    </div>

    <!-- Empty State -->
    <div v-if="!isLoading && jobs && jobs.length === 0 && !errorMessage" class="empty-state">
      <p>Belum ada manual job yang terdaftar</p>
      <router-link to="/create" class="btn-create">
        Buat Manual Job
      </router-link>
    </div>

    <!-- Jobs Table -->
    <div v-if="!isLoading && jobs.length > 0" class="table-container">
      <table class="jobs-table">
        <thead>
          <tr>
            <th>Nama Job</th>
            <th>Object</th>
            <th>GDrive</th>
            <th>Created At</th>
            <th>Last Run</th>
            <th>Status</th>
            <th>Action</th>
          </tr>
        </thead>
        <tbody>
          <ManualJobRow
            v-for="job in jobs"
            :key="job.id"
            :job="job"
            @trigger="handleTrigger"
            @view-script="handleViewScript"
            @delete="handleDeleteJob"
          />
        </tbody>
      </table>
    </div>

    <!-- Script Preview Modal -->
    <ScriptPreview
      :is-visible="isModalVisible"
      :job-id="currentJobId"
      :script-content="currentScript"
      @close="closeModal"
    />

    <!-- Toast Notification -->
    <Transition name="toast">
      <div
        v-if="toast.show"
        :class="getToastClass(toast.type)"
        class="toast-notification"
      >
        <div class="toast-content">
          <svg v-if="toast.type === 'success'" class="toast-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
          </svg>
          <svg v-else-if="toast.type === 'error'" class="toast-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
          <svg v-else class="toast-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <p class="toast-message">{{ toast.message }}</p>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import jobService from '@/services/jobService';
import ManualJobRow from '@/components/ManualJobRow.vue';
import ScriptPreview from '@/components/ScriptPreview.vue';

// State untuk jobs list
const jobs = ref([]);
const isLoading = ref(true);
const errorMessage = ref(null);

// State untuk Script Preview Modal
const isModalVisible = ref(false);
const currentScript = ref('');
const currentJobId = ref(null);

// Toast notification state
const toast = ref({
  show: false,
  message: '',
  type: 'success', // 'success' | 'error' | 'info'
});

// Fetch manual jobs
async function fetchData() {
  isLoading.value = true;
  errorMessage.value = null;
  
  try {
    const data = await jobService.getManualJobs();
    jobs.value = Array.isArray(data) ? data : [];
    console.log('✅ Manual jobs loaded:', jobs.value.length);
  } catch (error) {
    errorMessage.value = error.response?.data?.error || 'Gagal memuat data manual jobs.';
    console.error("Fetch Manual Jobs Error:", error);
  } finally {
    isLoading.value = false;
  }
}

// Handle Trigger Job
async function handleTrigger(jobId) {
  if (!confirm(`Apakah Anda yakin ingin menjalankan job ID ${jobId} sekarang?`)) {
    return;
  }
  
  try {
    await jobService.triggerManualJob(jobId);
    showToast(`Job ${jobId} berhasil di-trigger! Status akan diperbarui.`, 'success');
    
    // Refresh job list setelah trigger
    setTimeout(() => {
      fetchData();
    }, 1000);
  } catch (error) {
    const errorMsg = error.response?.data?.error || error.message || String(error);
    showToast(`Gagal men-trigger job: ${errorMsg}`, 'error');
  }
}

// Handle View Script
async function handleViewScript(jobId) {
  try {
    isModalVisible.value = true;
    currentScript.value = 'Loading script...';
    currentJobId.value = jobId;
    
    const data = await jobService.getJobScript(jobId);
    currentScript.value = data.script_preview || data.script || 'No script available';
  } catch (error) {
    isModalVisible.value = false;
    const errorMsg = error.response?.data?.error || error.message;
    showToast(`Gagal memuat pratinjau script: ${errorMsg}`, 'error');
  }
}

// Handle Delete Job
async function handleDeleteJob(jobId, jobName) {
  if (!confirm(`Apakah Anda yakin ingin menghapus job "${jobName}"?`)) {
    return;
  }

  try {
    await jobService.deleteJob(jobId);
    showToast('Job berhasil dihapus!', 'success');
    fetchData(); // Refresh job list
  } catch (error) {
    const errorMsg = error.response?.data?.error || error.message || 'Gagal menghapus job';
    showToast(errorMsg, 'error');
  }
}

// Close Script Preview Modal
function closeModal() {
  isModalVisible.value = false;
  currentScript.value = '';
  currentJobId.value = null;
}

// Toast Notification Helper
function showToast(message, type = 'success') {
  toast.value = { show: true, message, type };
  setTimeout(() => {
    toast.value.show = false;
  }, 3000);
}

// Get toast CSS class
function getToastClass(type) {
  const classes = {
    success: 'toast-success',
    error: 'toast-error',
    info: 'toast-info'
  };
  return classes[type] || 'toast-info';
}

// Initial load
onMounted(fetchData);
</script>

<style scoped>
.manual-jobs-view {
  max-width: 1400px;
  margin: 0 auto;
  padding: 2rem 1.5rem;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 2.5rem;
  gap: 1rem;
}

h1 {
  font-size: 1.75rem;
  font-weight: 600;
  color: #1a1a1a;
  margin: 0 0 0.5rem 0;
  letter-spacing: -0.02em;
}

.subtitle {
  font-size: 0.95rem;
  color: #666;
  margin: 0;
  font-weight: 400;
}

.btn-add {
  display: inline-flex;
  align-items: center;
  background: #1a1a1a;
  color: white;
  padding: 0.625rem 1.125rem;
  border-radius: 6px;
  text-decoration: none;
  font-weight: 500;
  font-size: 0.9375rem;
  transition: all 0.2s;
  white-space: nowrap;
}

.btn-add:hover {
  background: #333;
}

.status-message {
  padding: 1rem;
  border-radius: 6px;
  font-size: 0.9375rem;
  background: #f8f8f8;
  color: #666;
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.status-message.error {
  background: #fef2f2;
  color: #dc2626;
  border: 1px solid #fee2e2;
  justify-content: space-between;
}

.loading-dot {
  width: 8px;
  height: 8px;
  background: #666;
  border-radius: 50%;
  animation: pulse 1.5s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.3; }
}

.retry-btn {
  background: transparent;
  border: 1px solid #dc2626;
  color: #dc2626;
  padding: 0.375rem 0.875rem;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.875rem;
  transition: all 0.2s;
}

.retry-btn:hover {
  background: #dc2626;
  color: white;
}

.empty-state {
  text-align: center;
  padding: 3rem 1rem;
}

.empty-state p {
  color: #999;
  font-size: 0.9375rem;
  margin-bottom: 1.5rem;
}

.btn-create {
  display: inline-block;
  background: #1a1a1a;
  color: white;
  padding: 0.625rem 1.125rem;
  border-radius: 6px;
  text-decoration: none;
  font-weight: 500;
  font-size: 0.9375rem;
  transition: all 0.2s;
}

.btn-create:hover {
  background: #333;
}

.table-container {
  background: #fff;
  border: 1px solid #e5e5e5;
  border-radius: 8px;
  overflow: hidden;
}

.jobs-table {
  width: 100%;
  border-collapse: collapse;
}

.jobs-table thead {
  background: #fafafa;
}

.jobs-table th {
  padding: 0.875rem 1rem;
  text-align: left;
  font-size: 0.8125rem;
  font-weight: bold;
  color: #000000;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  border-bottom: 1px solid #e5e5e5;
}

.jobs-table td {
  padding: 1rem;
  border-bottom: 1px solid #f0f0f0;
  font-size: 0.9375rem;
  color: #333;
}

.jobs-table tbody tr:last-child td {
  border-bottom: none;
}

.jobs-table tbody tr:hover {
  background: #fafafa;
}

/* Toast Notification */
.toast-notification {
  position: fixed;
  bottom: 2rem;
  right: 2rem;
  background: white;
  border-radius: 8px;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.1);
  z-index: 1001;
  animation: slideIn 0.3s ease-out;
}

@keyframes slideIn {
  from {
    transform: translateX(400px);
    opacity: 0;
  }
  to {
    transform: translateX(0);
    opacity: 1;
  }
}

.toast-notification.toast-success {
  border-left: 4px solid #22c55e;
}

.toast-notification.toast-error {
  border-left: 4px solid #ef4444;
}

.toast-notification.toast-info {
  border-left: 4px solid #3b82f6;
}

.toast-content {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem 1.5rem;
}

.toast-icon {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
}

.toast-notification.toast-success .toast-icon {
  color: #22c55e;
}

.toast-notification.toast-error .toast-icon {
  color: #ef4444;
}

.toast-notification.toast-info .toast-icon {
  color: #3b82f6;
}

.toast-message {
  margin: 0;
  color: #1a1a1a;
  font-weight: 500;
  font-size: 0.9375rem;
  white-space: nowrap;
}

.toast-enter-active,
.toast-leave-active {
  transition: all 0.3s ease;
}

.toast-enter-from {
  transform: translateX(400px);
  opacity: 0;
}

.toast-leave-to {
  transform: translateX(400px);
  opacity: 0;
}

@media (max-width: 768px) {
  .manual-jobs-view {
    padding: 1.5rem 1rem;
  }
  
  .header {
    flex-direction: column;
    align-items: stretch;
  }
  
  h1 {
    font-size: 1.5rem;
  }
  
  .btn-add {
    justify-content: center;
  }
  
  .table-container {
    overflow-x: auto;
  }
  
  .jobs-table {
    min-width: 800px;
  }

  .toast-notification {
    bottom: 1rem;
    right: 1rem;
    left: 1rem;
  }
}
</style>