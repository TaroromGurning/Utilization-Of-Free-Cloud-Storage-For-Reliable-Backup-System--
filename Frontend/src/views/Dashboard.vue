<template>
  <div class="dashboard-container">
    <div class="header">
      <h1>Dashboard</h1>
      <p class="subtitle">Overview sistem backup dan monitoring activity</p>
    </div>

    <div class="stats-section">
      <div class="stats-grid">
        <div class="stat-card">
          <div class="stat-content">
            <span class="stat-label">Total Gdrive</span>
            <span class="stat-value">{{ totalRemotes }}</span>
          </div>
        </div>

        <div class="stat-card">
          <div class="stat-content">
            <span class="stat-label">Total Job</span>
            <span class="stat-value">{{ totalJobs }}</span>
          </div>
        </div>
      </div>

      <div class="storage-card">
        <div class="storage-header">
          <span class="storage-label">Total Disk Space</span>
          <span class="storage-value">{{ totalStorageGB }}GB</span>
        </div>
        <div class="storage-body">
          <div class="chart-container">
            <div v-if="isLoadingData" class="status-message">
              <span class="loading-dot"></span> Loading...
            </div>
            <StorageChart
              v-else
              :series="storageChartSeries"
              :labels="storageLabels"
              :colors="storageColors"
            />
          </div>
        </div>
      </div>
    </div>

    <div class="section">
      <h2>Quick Action</h2>
      <p class="section-subtitle">Start backup or restore job instantly</p>
      
      <div class="action-cards">
        <button class="action-card" @click="openBackupModal">
          <div class="action-content">
            <h3>Create Backup</h3>
            <p>Start backup job</p>
          </div>
          <span class="action-arrow">→</span>
        </button>

        <button class="action-card" @click="openRestoreModal">
          <div class="action-content">
            <h3>Start Restore</h3>
            <p>Restore from Gdrive</p>
          </div>
          <span class="action-arrow">→</span>
        </button>
      </div>
    </div>

    <div class="section">
      <h2>Next Job</h2>
      
      <div v-if="isLoadingJobs" class="status-message">
        <span class="loading-dot"></span>
        Loading jobs...
      </div>

      <div v-else-if="nextJob" class="next-job-card">
        <div class="job-header">
          <span class="job-type-badge">{{ nextJob.sourcePath }}</span>
          <span class="job-time">{{ formatNextRun(nextJob.nextRun) }}</span>
        </div>
        <h3>{{ nextJob.name }}</h3>
        <div class="job-details">
          <div class="detail-item">
            <span class="detail-label">Gdrive</span>
            <span class="detail-value">{{ nextJob.remoteName }}</span>
          </div>
          <div class="detail-item">
            <span class="detail-label">Schedule</span>
            <span class="detail-value">{{ nextJob.scheduleCron }}</span>
          </div>
        </div>
      </div>

      <div v-else class="empty-state">
        <p>No scheduled jobs</p>
      </div>
    </div>

    <QuickBackupModal
      :is-visible="showBackupModal"
      @close="showBackupModal = false"
      @success="handleBackupSuccess"
    />

    <QuickRestoreModal
      :is-visible="showRestoreModal"
      @close="showRestoreModal = false"
      @success="handleRestoreSuccess"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue';
import monitoringService from '@/services/monitoringService';
import QuickBackupModal from '@/components/CreateBackup.vue';
import QuickRestoreModal from '@/components/CreateRestore.vue';
import StorageChart from '@/components/StorageChart.vue';
import jobService from '../services/jobService';

// State
const remotes = ref([]);
const allJobs = ref([]); // Untuk menampung semua job (Manual + Scheduled)
const scheduledJobs = ref([]); // Khusus untuk logika Next Job
const isLoadingData = ref(true);
const isLoadingJobs = ref(true);
const showBackupModal = ref(false);
const showRestoreModal = ref(false);
const driveColors = new Map();

// Computed
const totalRemotes = computed(() => remotes.value.length);

// Menghitung total job dari list allJobs, kecuali yang RESTORE
const totalJobs = computed(() => {
  return allJobs.value.filter(job => job.operation_mode !== 'RESTORE').length;
});

const totalStorageGB = computed(() => {
  const total = remotes.value.reduce((sum, remote) => {
    if (remote.status_connect === 'CONNECTED') {
      return sum + (remote.total_storage_gb || 0);
    }
    return sum;
  }, 0);
  return total.toFixed(2);
});

const storageChartSeries = computed(() => {
  const series = [];
  remotes.value
    .filter(remote => remote.status_connect === 'CONNECTED')
    .forEach(remote => {
      const used = remote.used_storage_gb || 0;
      const total = remote.total_storage_gb || 0;
      const free = total - used;
      series.push(used);
      if (free > 0) series.push(free);
    });
  return series;
});

const storageLabels = computed(() => {
  const labels = [];
  remotes.value
    .filter(remote => remote.status_connect === 'CONNECTED')
    .forEach(remote => {
      labels.push(remote.remote_name);
      const used = remote.used_storage_gb || 0;
      const total = remote.total_storage_gb || 0;
      const free = total - used;
      if (free > 0) labels.push(`${remote.remote_name} (Free)`);
    });
  return labels;
});

const storageColors = computed(() => {
  const colors = [];
  remotes.value
    .filter(remote => remote.status_connect === 'CONNECTED')
    .forEach((remote, index) => {
      const color = getColorForDrive(remote.remote_name, index);
      colors.push(color);
      const used = remote.used_storage_gb || 0;
      const total = remote.total_storage_gb || 0;
      const free = total - used;
      if (free > 0) {
        colors.push(hexToRgba(color, 0.3));
      }
    });
  return colors;
});

const nextJob = computed(() => {
  if (!Array.isArray(scheduledJobs.value) || scheduledJobs.value.length === 0) return null;
  
  const futureJobs = scheduledJobs.value
    .map(j => ({ ...j, nextRunDate: new Date(j.next_run) }))
    .filter(j => j.nextRunDate && j.nextRunDate > new Date())
    .sort((a, b) => a.nextRunDate - b.nextRunDate);
  
  return futureJobs.length > 0 ? {
    jobType: futureJobs[0].type,
    nextRun: futureJobs[0].nextRunDate,
    name: futureJobs[0].job_name,
    remoteName: futureJobs[0].gdrive_target,
    scheduleCron: futureJobs[0].next_run,
    sourcePath: futureJobs[0].source_path,
  } : null;
});

// Lifecycle
onMounted(async () => {
  await fetchData();
});

// Functions
const predefinedColors = [
  '#66A5AD', '#A0D468', '#F68484', '#B49FBC', '#FDBE34',
  '#7D9D9C', '#D97F7F', '#6A8EAE', '#A29B7F', '#CC9A78'
];

function hexToRgba(hex, alpha) {
  const r = parseInt(hex.slice(1, 3), 16);
  const g = parseInt(hex.slice(3, 5), 16);
  const b = parseInt(hex.slice(5, 7), 16);
  return `rgba(${r}, ${g}, ${b}, ${alpha})`;
}

function getColorForDrive(driveName, index) {
  if (!driveColors.has(driveName)) {
    driveColors.set(driveName, predefinedColors[index % predefinedColors.length]);
  }
  return driveColors.get(driveName);
}

async function fetchData() {
  isLoadingData.value = true;
  isLoadingJobs.value = true;
  
  try {
    // Panggil 3 endpoint paralel: Remote, All Jobs, Scheduled Jobs
    const [remotesData, allJobsData, scheduledJobsData] = await Promise.all([
      monitoringService.getRemoteStatus(),
      jobService.getAllJobs(),
      monitoringService.getScheduledJobs()
    ]);
    
    remotes.value = Array.isArray(remotesData) ? remotesData : [];
    allJobs.value = Array.isArray(allJobsData) ? allJobsData : [];
    scheduledJobs.value = Array.isArray(scheduledJobsData) ? scheduledJobsData : [];

  } catch (error) {
    console.error('Failed to fetch dashboard data:', error);
    remotes.value = [];
    allJobs.value = [];
    scheduledJobs.value = [];
  } finally {
    isLoadingData.value = false;
    isLoadingJobs.value = false;
  }
}

function openBackupModal() {
  showBackupModal.value = true;
}

function openRestoreModal() {
  showRestoreModal.value = true;
}

function handleBackupSuccess() {
  showBackupModal.value = false;
  fetchData();
}

function handleRestoreSuccess() {
  showRestoreModal.value = false;
}

function formatNextRun(timestamp) {
  if (!timestamp) return '-';
  
  try {
    const date = new Date(timestamp);
    const now = new Date();
    const diff = date - now;
    
    if (diff < 0) return 'Overdue';
    
    const hours = Math.floor(diff / 3600000);
    const minutes = Math.floor((diff % 3600000) / 60000);
    
    if (hours < 1) return `in ${minutes}m`;
    if (hours < 24) return `in ${hours}h ${minutes}m`;
    
    return date.toLocaleString('id-ID', {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  } catch (e) {
    return timestamp;
  }
}
</script>

<style scoped>
.dashboard-container {
  max-width: 1400px;
  margin: 0 auto;
  padding: 2rem 1.5rem;
}

/* Page Header */
.header {
  margin-bottom: 2.5rem;
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

/* Stats Section */
.stats-section {
  display: grid;
  grid-template-columns: 1fr 2fr;
  gap: 1rem;
  margin-bottom: 2rem;
}

/* Stats Grid */
.stats-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 1rem;
}

.stat-card {
  background: white;
  padding: 1.5rem;
  border-radius: 8px;
  border: 1px solid #e5e5e5;
}

.stat-content {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.stat-label {
  font-size: 0.8125rem;
  color: #666;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.stat-value {
  font-size: 2rem;
  font-weight: 600;
  color: #1a1a1a;
  line-height: 1;
}

/* Storage Card */
.storage-card {
  background: white;
  padding: 1.5rem;
  border-radius: 8px;
  border: 1px solid #e5e5e5;
}

.storage-header {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  margin-bottom: 1.5rem;
  text-align: center;
}

.storage-label {
  font-size: 0.8125rem;
  color: #666;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.storage-value {
  font-size: 2rem;
  font-weight: 600;
  color: #1a1a1a;
}

.storage-body {
  display: grid;
  grid-template-columns: 1fr;
  gap: 2rem;
  align-items: center;
}

.chart-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 300px;
  padding: 0;
}

/* Section */
.section {
  background: white;
  padding: 1.5rem;
  border-radius: 8px;
  border: 1px solid #e5e5e5;
  margin-bottom: 1.5rem;
}

.section h2 {
  font-size: 1.125rem;
  font-weight: 600;
  color: #1a1a1a;
  margin: 0 0 0.25rem 0;
}

.section-subtitle {
  font-size: 0.875rem;
  color: #666;
  margin: 0 0 1.5rem 0;
}

/* Action Cards */
.action-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 1rem;
}

.action-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1.25rem;
  border: 1px solid #e5e5e5;
  border-radius: 6px;
  background: white;
  cursor: pointer;
  transition: all 0.2s;
  text-align: left;
  width: 100%;
  font-family: inherit;
}

.action-card:hover {
  background: #f5f5f5;
  border-color: #1a1a1a;
}

.action-content h3 {
  font-size: 1rem;
  font-weight: 600;
  color: #1a1a1a;
  margin: 0 0 0.25rem 0;
}

.action-content p {
  margin: 0;
  font-size: 0.875rem;
  color: #666;
}

.action-arrow {
  font-size: 1.25rem;
  color: #666;
  transition: transform 0.2s;
  margin-left: 1rem;
}

.action-card:hover .action-arrow {
  transform: translateX(4px);
  color: #1a1a1a;
}

/* Next Job Card */
.next-job-card {
  background: #fafafa;
  padding: 1.25rem;
  border-radius: 6px;
  border-left: 3px solid #1a1a1a;
}

.job-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.875rem;
}

.job-type-badge {
  background: #1a1a1a;
  color: white;
  padding: 0.25rem 0.625rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.job-time {
  color: #666;
  font-weight: 500;
  font-size: 0.875rem;
}

.next-job-card h3 {
  font-size: 1.125rem;
  font-weight: 600;
  color: #1a1a1a;
  margin: 0 0 1rem 0;
}

.job-details {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.detail-label {
  font-size: 0.75rem;
  color: #666;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.detail-value {
  font-size: 0.9375rem;
  color: #1a1a1a;
  font-weight: 500;
  font-family: monospace;
}

/* Status Message */
.status-message {
  padding: 1rem;
  border-radius: 6px;
  font-size: 0.9375rem;
  background: #f8f8f8;
  color: #666;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  justify-content: center;
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

/* Empty State */
.empty-state {
  text-align: center;
  padding: 2rem 1rem;
}

.empty-state p {
  margin: 0;
  color: #999;
  font-size: 0.9375rem;
}

/* Responsive */
@media (max-width: 1024px) {
  .stats-section {
    grid-template-columns: 1fr;
  }

  .storage-body {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .dashboard-container {
    padding: 1.5rem 1rem;
  }
  
  h1 {
    font-size: 1.5rem;
  }
  
  .action-cards {
    grid-template-columns: 1fr;
  }
  
  .job-details {
    grid-template-columns: 1fr;
  }
}
</style>