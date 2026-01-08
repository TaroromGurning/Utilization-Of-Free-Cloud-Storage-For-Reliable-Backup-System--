<template>
  <div class="live-log-panel">
    <!-- Header -->
    <div class="panel-header">
      <h3>Recent Logs</h3>
      <button 
        @click="fetchLogs" 
        class="refresh-btn"
        :class="{ spinning: isRefreshing }"
        title="Refresh"
      >
        ↻
      </button>
    </div>

    <!-- Stats Bar -->
    <div class="stats-bar">
      <div class="stat success">
        <div class="stat-info">
          <span class="stat-value">{{ successCount }}</span>
          <span class="stat-label">Success</span>
        </div>
      </div>
      <div class="stat failed">
        <div class="stat-info">
          <span class="stat-value">{{ failedCount }}</span>
          <span class="stat-label">Failed</span>
        </div>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="isLoading" class="status-message">
      <span class="loading-dot"></span>
      <span>Loading...</span>
    </div>

    <!-- Error -->
    <div v-else-if="error" class="status-message error">
      <span>{{ error }}</span>
      <button @click="fetchLogs" class="retry-btn">Retry</button>
    </div>

    <!-- Logs List -->
    <div v-else class="logs-list">
      <div v-if="logs.length === 0" class="empty-state">
        <p>No logs yet</p>
      </div>

      <div 
        v-for="log in logs" 
        :key="log.ID"
        class="log-item"
        :class="getStatusClass(log.Status)"
      >
        <!-- Status Indicator -->
        <div class="log-status"></div>

        <!-- Log Info -->
        <div class="log-info">
          <div class="log-name">{{ getJobName(log) }}</div>
          <div class="log-meta">
            <span class="duration">{{ formatDuration(log.DurationSec) }}</span>
            <span class="separator">•</span>
            <span class="time">{{ formatTime(log.Timestamp) }}</span>
          </div>
        </div>

        <!-- Detail Button -->
        <button 
          @click="viewDetail(log)" 
          class="detail-btn"
          title="View detail"
        >
          →
        </button>
      </div>
    </div>

    <!-- Detail Modal -->
    <SimpleLogModal
      :is-visible="showModal"
      :log="selectedLog"
      @close="closeModal"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue';
import monitoringService from '@/services/monitoringService';
import SimpleLogModal from './LogModel.vue';

// ============================================================
// STATE
// ============================================================
const logs = ref([]);
const isLoading = ref(true);
const isRefreshing = ref(false);
const error = ref(null);
const showModal = ref(false);
const selectedLog = ref(null);

let refreshInterval = null;

// ============================================================
// COMPUTED
// ============================================================
const successCount = computed(() => {
  return logs.value.filter(l => 
    ['SUCCESS', 'COMPLETED'].includes((l.Status || '').toUpperCase())
  ).length;
});

const failedCount = computed(() => {
  return logs.value.filter(l => 
    (l.Status || '').toUpperCase().includes('FAIL') || 
    (l.Status || '').toUpperCase() === 'ERROR'
  ).length;
});

// ============================================================
// METHODS: FETCH DATA
// ============================================================
async function fetchLogs() {
  if (isRefreshing.value) return;
  
  isRefreshing.value = true;
  error.value = null;
  
  try {
    const data = await monitoringService.getLogs();
    
    // Sort by timestamp (newest first) dan ambil 20 terakhir
    logs.value = (Array.isArray(data) ? data : [])
      .sort((a, b) => new Date(b.Timestamp) - new Date(a.Timestamp))
      .slice(0, 20);
  } catch (err) {
    console.error('Failed to fetch logs:', err);
    error.value = 'Failed to load logs';
  } finally {
    isLoading.value = false;
    isRefreshing.value = false;
  }
}

// ============================================================
// METHODS: AUTO REFRESH
// ============================================================
function startAutoRefresh() {
  refreshInterval = setInterval(fetchLogs, 10000);
}

function stopAutoRefresh() {
  if (refreshInterval) {
    clearInterval(refreshInterval);
    refreshInterval = null;
  }
}

// ============================================================
// METHODS: MODAL
// ============================================================
function viewDetail(log) {
  selectedLog.value = log;
  showModal.value = true;
}

function closeModal() {
  showModal.value = false;
  selectedLog.value = null;
}

// ============================================================
// HELPERS: DATA EXTRACTION
// ============================================================

/**
 * ✅ Get Job Name dengan prioritas:
 * 1. Field baru di logs table (JobName/job_name)
 * 2. Relasi ScheduledJob (untuk backup job)
 * 3. ConfigSnapshot (untuk manual job lama)
 */
function getJobName(log) {
  if (!log) return 'Unknown';

  // 1. PRIORITAS: Cek dari field baru di logs table
  if (log.JobName) return log.JobName;
  if (log.job_name) return log.job_name;

  // 2. Fallback: Cek dari relasi ScheduledJob
  if (log.ScheduledJob?.JobName) return log.ScheduledJob.JobName;
  if (log.scheduled_job?.job_name) return log.scheduled_job.job_name;

  // 3. Fallback: Cek dari ConfigSnapshot
  if (log.ConfigSnapshot) {
    try {
      const config = typeof log.ConfigSnapshot === 'string' 
        ? JSON.parse(log.ConfigSnapshot) 
        : log.ConfigSnapshot;
      return config.job_name || config.JobName || 'Manual Job';
    } catch (e) {
      return 'Manual Job';
    }
  }

  return 'Unknown';
}

/**
 * Get Status Class untuk styling
 */
function getStatusClass(status) {
  if (!status) return 'pending';
  
  const s = status.toUpperCase();
  if (['SUCCESS', 'COMPLETED'].includes(s)) return 'success';
  if (s.includes('FAIL') || s === 'ERROR') return 'failed';
  if (s === 'RUNNING') return 'running';
  
  return 'pending';
}

// ============================================================
// HELPERS: FORMATTING
// ============================================================

/**
 * Format duration dalam format readable
 */
function formatDuration(seconds) {
  if (!seconds || seconds === 0) return '-';
  
  if (seconds < 60) return `${seconds}s`;
  
  const m = Math.floor(seconds / 60);
  const s = seconds % 60;
  
  return `${m}m ${s}s`;
}

/**
 * Format timestamp dalam format relative time
 */
function formatTime(timestamp) {
  if (!timestamp) return '-';
  
  try {
    const date = new Date(timestamp);
    const now = new Date();
    const diff = now - date;
    
    // Less than 1 minute
    if (diff < 60000) return 'Just now';
    
    // Less than 1 hour
    if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`;
    
    // Less than 1 day
    if (diff < 86400000) return `${Math.floor(diff / 3600000)}h ago`;
    
    // More than 1 day
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

// ============================================================
// LIFECYCLE
// ============================================================
onMounted(() => {
  fetchLogs();
  startAutoRefresh();
});

onUnmounted(() => {
  stopAutoRefresh();
});
</script>

<style scoped>
/* ============================================================ */
/* LAYOUT UTAMA */
/* ============================================================ */
.live-log-panel {
  background: white;
  border: 1px solid #e5e5e5;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

/* ============================================================ */
/* HEADER */
/* ============================================================ */
.panel-header {
  background: #fafafa;
  padding: 1rem 1.25rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid #e5e5e5;
}

.panel-header h3 {
  margin: 0;
  font-size: 0.9375rem;
  font-weight: 600;
  color: #1a1a1a;
}

.refresh-btn {
  background: transparent;
  border: 1px solid #e5e5e5;
  color: #666;
  width: 32px;
  height: 32px;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 1.25rem;
  display: flex;
  align-items: center;
  justify-content: center;
}

.refresh-btn:hover {
  background: #f5f5f5;
  border-color: #1a1a1a;
  color: #1a1a1a;
}

.refresh-btn.spinning {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* ============================================================ */
/* STATS BAR */
/* ============================================================ */
.stats-bar {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
  padding: 1rem 1.25rem;
  background: #fafafa;
  border-bottom: 1px solid #e5e5e5;
}

.stat {
  display: flex;
  align-items: center;
  padding: 0.75rem;
  background: white;
  border-radius: 6px;
  border-left: 3px solid;
}

.stat.success {
  border-left-color: #22c55e;
}

.stat.failed {
  border-left-color: #ef4444;
}

.stat-info {
  display: flex;
  flex-direction: column;
}

.stat-value {
  font-size: 1.5rem;
  font-weight: 600;
  color: #1a1a1a;
  line-height: 1;
}

.stat-label {
  font-size: 0.75rem;
  color: #666;
  font-weight: 500;
  margin-top: 0.25rem;
}

/* ============================================================ */
/* STATUS MESSAGES */
/* ============================================================ */
.status-message {
  padding: 1.5rem 1.25rem;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  font-size: 0.875rem;
  color: #666;
}

.status-message.error {
  background: #fef2f2;
  color: #dc2626;
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
  padding: 0.375rem 0.75rem;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.8125rem;
  transition: all 0.2s;
}

.retry-btn:hover {
  background: #dc2626;
  color: white;
}

/* ============================================================ */
/* LOGS LIST */
/* ============================================================ */
.logs-list {
  flex: 1;
  overflow-y: auto;
  padding: 0.75rem;
}

.logs-list::-webkit-scrollbar {
  width: 6px;
}

.logs-list::-webkit-scrollbar-track {
  background: #f8f8f8;
}

.logs-list::-webkit-scrollbar-thumb {
  background: #d4d4d4;
  border-radius: 3px;
}

.logs-list::-webkit-scrollbar-thumb:hover {
  background: #a3a3a3;
}

/* ============================================================ */
/* EMPTY STATE */
/* ============================================================ */
.empty-state {
  text-align: center;
  padding: 2rem 1rem;
}

.empty-state p {
  margin: 0;
  color: #999;
  font-size: 0.875rem;
}

/* ============================================================ */
/* LOG ITEM */
/* ============================================================ */
.log-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem;
  background: #fafafa;
  border-radius: 6px;
  margin-bottom: 0.5rem;
  border-left: 3px solid #d4d4d4;
  transition: all 0.2s;
}

.log-item:hover {
  background: #f5f5f5;
}

.log-item:last-child {
  margin-bottom: 0;
}

.log-item.success {
  border-left-color: #22c55e;
}

.log-item.failed {
  border-left-color: #ef4444;
}

.log-item.running {
  border-left-color: #f59e0b;
}

.log-item.pending {
  border-left-color: #6b7280;
}

/* ============================================================ */
/* LOG STATUS INDICATOR */
/* ============================================================ */
.log-status {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.log-item.success .log-status {
  background: #22c55e;
}

.log-item.failed .log-status {
  background: #ef4444;
}

.log-item.running .log-status {
  background: #f59e0b;
}

.log-item.pending .log-status {
  background: #6b7280;
}

/* ============================================================ */
/* LOG INFO */
/* ============================================================ */
.log-info {
  flex: 1;
  min-width: 0;
}

.log-name {
  font-weight: 500;
  font-size: 0.875rem;
  color: #1a1a1a;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-bottom: 0.25rem;
}

.log-meta {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.75rem;
  color: #666;
}

.separator {
  color: #d4d4d4;
}

/* ============================================================ */
/* DETAIL BUTTON */
/* ============================================================ */
.detail-btn {
  background: transparent;
  border: 1px solid #e5e5e5;
  color: #666;
  width: 28px;
  height: 28px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 1rem;
  transition: all 0.2s;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0;
}

.detail-btn:hover {
  background: #f5f5f5;
  border-color: #1a1a1a;
  color: #1a1a1a;
}

/* ============================================================ */
/* RESPONSIVE */
/* ============================================================ */
@media (max-width: 768px) {
  .panel-header {
    padding: 0.875rem 1rem;
  }

  .stats-bar {
    padding: 0.875rem 1rem;
    gap: 0.75rem;
  }

  .log-item {
    padding: 0.625rem;
    gap: 0.5rem;
  }

  .log-name {
    font-size: 0.8125rem;
  }

  .log-meta {
    font-size: 0.7rem;
  }
}
</style>