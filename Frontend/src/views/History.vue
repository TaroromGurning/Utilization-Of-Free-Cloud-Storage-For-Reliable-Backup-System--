<!-- ============================================ -->
<!-- FILE: src/views/History.vue - Updated -->
<!-- ============================================ -->
<template>
  <div class="history-page">
    <!-- Header -->
    <div class="page-header">
      <div class="header-left">
        <h1>Job History</h1>
        <p>Complete execution history with detailed logs</p>
      </div>
    </div>

    <!-- Filters -->
    <div class="filters-section">
      <!-- Search -->
      <div class="search-box">
        <span class="search-icon">🔍</span>
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Search by job name..."
          class="search-input"
        />
        <button v-if="searchQuery" @click="searchQuery = ''" class="clear-btn">
          ✕
        </button>
      </div>

      <!-- Filter by GDrive -->
      <select v-model="selectedRemote" class="filter-select">
        <option value="">All Remotes</option>
        <option v-for="remote in uniqueRemotes" :key="remote" :value="remote">
          {{ remote }}
        </option>
      </select>

      <!-- Filter by Status -->
      <select v-model="selectedStatus" class="filter-select">
        <option value="">All Status</option>
        <option value="SUCCESS">Success</option>
        <option value="FAILED">Failed</option>
      </select>

      <!-- Clear Filters -->
      <button
        v-if="hasActiveFilters"
        @click="clearFilters"
        class="clear-filters-btn"
      >
        Clear Filters
      </button>
    </div>

    <!-- Stats -->
    <div class="stats-row">
      <div class="stat-card">
        <span class="stat-label">Total Logs</span>
        <span class="stat-value">{{ filteredLogs.length }}</span>
      </div>
      <div class="stat-card success">
        <span class="stat-label">Success</span>
        <span class="stat-value">{{ statsCount.success }}</span>
      </div>
      <div class="stat-card failed">
        <span class="stat-label">Failed</span>
        <span class="stat-value">{{ statsCount.failed }}</span>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="isLoading" class="loading-state">
      <div class="spinner"></div>
      <p>Loading history...</p>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="error-state">
      <span class="error-icon">⚠️</span>
      <p>{{ error }}</p>
      <button @click="fetchLogs" class="retry-btn">Retry</button>
    </div>

    <!-- Empty State -->
    <div v-else-if="filteredLogs.length === 0" class="empty-state">
      <div class="empty-icon">📭</div>
      <h3>No Logs Found</h3>
      <p v-if="hasActiveFilters">
        Try adjusting your filters or search query
      </p>
      <p v-else>No execution history available yet</p>
    </div>

    <!-- History Table -->
    <div v-else class="table-container">
      <table class="history-table">
        <thead>
          <tr>
            <th style="width: 40px"></th>
            <th>Job Name</th>
            <th>Remote</th>
            <th>Source Path</th>
            <th>Status</th>
            <th>Duration</th>
            <th>Size</th>
            <th>Timestamp</th>
          </tr>
        </thead>
        <tbody>
          <template v-for="log in paginatedLogs" :key="log.ID">
            <!-- Main Row -->
            <tr
              @click="toggleRow(log.ID)"
              class="log-row"
              :class="[getStatusClass(log.Status), { expanded: expandedRows.has(log.ID) }]"
            >
              <td class="expand-cell">
                <button class="expand-btn">
                  <span v-if="expandedRows.has(log.ID)">▼</span>
                  <span v-else>▶</span>
                </button>
              </td>
              <td>
                <div class="job-name">{{ getJobName(log) }}</div>
              </td>
              <td>
                <span class="remote-badge">
                  {{ getRemoteName(log) }}
                </span>
              </td>
              <td>
                <div class="source-path">{{ log.SourcePath || log.source_path || '-' }}</div>
              </td>
              <td>
                <span class="status-badge" :class="getStatusBadgeClass(log.Status)">
                  {{ log.Status }}
                </span>
              </td>
              <td>
                <span class="duration">{{ formatDuration(log.DurationSec) }}</span>
              </td>
              <td>
                <span class="file-size">{{ formatBytes(log.TransferredBytes) }}</span>
              </td>
              <td>
                <div class="timestamp">
                  <div class="time">{{ formatTime(log.Timestamp) }}</div>
                  <div class="date">{{ formatDate(log.Timestamp) }}</div>
                </div>
              </td>
            </tr>

            <!-- Expanded Row (Message Detail) -->
            <tr v-if="expandedRows.has(log.ID)" class="detail-row">
              <td colspan="8">
                <div class="detail-content">
                  <div class="detail-section">
                    <h4>📝 Execution Message</h4>
                    <pre class="message-box">{{ log.Message || 'No message available' }}</pre>
                  </div>

                  <div class="detail-grid">
                    <div class="detail-item">
                      <span class="detail-label">Job ID</span>
                      <span class="detail-value">{{ log.JobID || 'Manual' }}</span>
                    </div>
                    <div class="detail-item">
                      <span class="detail-label">Log ID</span>
                      <span class="detail-value">#{{ log.ID }}</span>
                    </div>
                    <div class="detail-item">
                      <span class="detail-label">Transferred</span>
                      <span class="detail-value">{{ formatBytes(log.TransferredBytes) }}</span>
                    </div>
                    <div class="detail-item">
                      <span class="detail-label">Duration</span>
                      <span class="detail-value">{{ formatDuration(log.DurationSec) }}</span>
                    </div>
                  </div>

                  <div v-if="log.ConfigSnapshot" class="detail-section">
                    <h4>⚙️ Configuration</h4>
                    <pre class="config-box">{{ formatJSON(log.ConfigSnapshot) }}</pre>
                  </div>
                </div>
              </td>
            </tr>
          </template>
        </tbody>
      </table>

      <!-- Pagination -->
      <div class="pagination">
        <button
          @click="currentPage--"
          :disabled="currentPage === 1"
          class="page-btn"
        >
          ← Previous
        </button>

        <div class="page-info">
          <span>Page {{ currentPage }} of {{ totalPages }}</span>
          <span class="separator">•</span>
          <span>{{ paginatedLogs.length }} of {{ filteredLogs.length }} logs</span>
        </div>

        <button
          @click="currentPage++"
          :disabled="currentPage === totalPages"
          class="page-btn"
        >
          Next →
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue';
import monitoringService from '@/services/monitoringService';

// ============================================================
// STATE
// ============================================================
const logs = ref([]);
const isLoading = ref(true);
const isRefreshing = ref(false);
const error = ref(null);

// Filters
const searchQuery = ref('');
const selectedRemote = ref('');
const selectedStatus = ref('');

// Pagination
const currentPage = ref(1);
const itemsPerPage = 50;

// Expanded rows
const expandedRows = ref(new Set());

// ============================================================
// FETCH DATA
// ============================================================
async function fetchLogs() {
  isRefreshing.value = true;
  error.value = null;

  try {
    const data = await monitoringService.getLogs();
    logs.value = Array.isArray(data) ? data : [];
  } catch (err) {
    console.error('Failed to fetch logs:', err);
    error.value = 'Failed to load history. Please try again.';
  } finally {
    isLoading.value = false;
    isRefreshing.value = false;
  }
}

onMounted(fetchLogs);

// ============================================================
// COMPUTED: UNIQUE VALUES FOR FILTERS
// ============================================================
const uniqueRemotes = computed(() => {
  const remotes = new Set();
  logs.value.forEach(log => {
    const remote = getRemoteName(log);
    if (remote && remote !== 'N/A') remotes.add(remote);
  });
  return Array.from(remotes).sort();
});

// ============================================================
// COMPUTED: FILTERED LOGS
// ============================================================
const filteredLogs = computed(() => {
  let result = [...logs.value];

  // Filter by search (job name)
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase();
    result = result.filter(log => {
      const jobName = getJobName(log).toLowerCase();
      const sourcePath = (log.SourcePath || log.source_path || '').toLowerCase();
      return jobName.includes(query) || sourcePath.includes(query);
    });
  }

  // Filter by remote
  if (selectedRemote.value) {
    result = result.filter(log => getRemoteName(log) === selectedRemote.value);
  }

  // Filter by status
  if (selectedStatus.value) {
    result = result.filter(log => {
      const status = (log.Status || '').toUpperCase();
      
      if (selectedStatus.value === 'SUCCESS') {
        // Success includes: SUCCESS, COMPLETED
        return ['SUCCESS', 'COMPLETED'].includes(status);
      } else if (selectedStatus.value === 'FAILED') {
        // Failed includes: anything that's NOT success/completed
        return !['SUCCESS', 'COMPLETED'].includes(status);
      }
      
      return true;
    });
  }

  // Sort by timestamp (newest first)
  result.sort((a, b) => new Date(b.Timestamp) - new Date(a.Timestamp));

  return result;
});

// ============================================================
// COMPUTED: PAGINATION
// ============================================================
const totalPages = computed(() => {
  return Math.ceil(filteredLogs.value.length / itemsPerPage);
});

const paginatedLogs = computed(() => {
  const start = (currentPage.value - 1) * itemsPerPage;
  const end = start + itemsPerPage;
  return filteredLogs.value.slice(start, end);
});

// ============================================================
// COMPUTED: STATS
// ============================================================
const statsCount = computed(() => {
  const total = filteredLogs.value.length;
  const success = filteredLogs.value.filter(l =>
    ['SUCCESS', 'COMPLETED'].includes((l.Status || '').toUpperCase())
  ).length;
  const failed = filteredLogs.value.filter(l => {
    const status = (l.Status || '').toUpperCase();
    return status.includes('FAIL') || status === 'ERROR';
  }).length;

  return { total, success, failed };
});

const hasActiveFilters = computed(() => {
  return searchQuery.value || selectedRemote.value || selectedStatus.value;
});

// ============================================================
// METHODS: ROW EXPANSION
// ============================================================
function toggleRow(id) {
  if (expandedRows.value.has(id)) {
    expandedRows.value.delete(id);
  } else {
    expandedRows.value.add(id);
  }
}

function clearFilters() {
  searchQuery.value = '';
  selectedRemote.value = '';
  selectedStatus.value = '';
  currentPage.value = 1;
}

// ============================================================
// HELPERS: DATA EXTRACTION
// ============================================================
function getJobName(log) {
  if (log.JobName || log.job_name) return log.JobName || log.job_name;
  if (log.ScheduledJob?.JobName) return log.ScheduledJob.JobName;

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

function getRemoteName(log) {
  if (log.ScheduledJob?.RemoteName) return log.ScheduledJob.RemoteName;

  if (log.ConfigSnapshot) {
    try {
      const config = typeof log.ConfigSnapshot === 'string'
        ? JSON.parse(log.ConfigSnapshot)
        : log.ConfigSnapshot;
      return config.remote_name || config.RemoteName || 'N/A';
    } catch (e) {
      return 'N/A';
    }
  }

  return 'N/A';
}

// ============================================================
// HELPERS: STYLING
// ============================================================
function getStatusClass(status) {
  if (!status) return '';
  const s = status.toUpperCase();
  if (['SUCCESS', 'COMPLETED'].includes(s)) return 'status-success';
  if (s.includes('FAIL') || s === 'ERROR') return 'status-failed';
  if (s === 'RUNNING') return 'status-running';
  return '';
}

function getStatusBadgeClass(status) {
  if (!status) return 'badge-default';
  const s = status.toUpperCase();
  if (['SUCCESS', 'COMPLETED'].includes(s)) return 'badge-success';
  if (s.includes('FAIL') || s === 'ERROR') return 'badge-failed';
  if (s === 'RUNNING') return 'badge-running';
  return 'badge-default';
}

// ============================================================
// HELPERS: FORMATTING
// ============================================================
function formatDuration(seconds) {
  if (!seconds || seconds === 0) return '-';
  if (seconds < 60) return `${seconds}s`;

  const m = Math.floor(seconds / 60);
  const s = seconds % 60;
  return `${m}m ${s}s`;
}

function formatBytes(bytes) {
  if (!bytes || bytes === 0) return '-';

  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  const k = 1024;
  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + units[i];
}

function formatTime(timestamp) {
  if (!timestamp) return '-';
  try {
    return new Date(timestamp).toLocaleTimeString('id-ID', {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    });
  } catch (e) {
    return timestamp;
  }
}

function formatDate(timestamp) {
  if (!timestamp) return '-';
  try {
    return new Date(timestamp).toLocaleDateString('id-ID', {
      day: 'numeric',
      month: 'short',
      year: 'numeric'
    });
  } catch (e) {
    return timestamp;
  }
}

function formatJSON(json) {
  if (!json) return '';
  try {
    const obj = typeof json === 'string' ? JSON.parse(json) : json;
    return JSON.stringify(obj, null, 2);
  } catch (e) {
    return json;
  }
}
</script>

<style scoped>
.history-page {
  padding: 0;
  max-width: 100%;
}

/* ============================================ */
/* HEADER */
/* ============================================ */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 1.5rem;
  gap: 1rem;
}

.header-left h1 {
  font-size: 1.75rem;
  font-weight: 600;
  color: #1a1a1a;
  margin: 0 0 0.5rem 0;
}

.header-left p {
  color: #666;
  margin: 0;
  font-size: 0.9375rem;
}

.refresh-btn {
  background: white;
  border: 1px solid #e5e5e5;
  color: #1a1a1a;
  padding: 0.625rem 1.25rem;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  transition: all 0.2s;
}

.refresh-btn:hover:not(:disabled) {
  background: #f5f5f5;
  border-color: #1a1a1a;
}

.refresh-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.spinning {
  display: inline-block;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* ============================================ */
/* FILTERS - HORIZONTAL LAYOUT */
/* ============================================ */
.filters-section {
  display: flex;
  gap: 0.75rem;
  margin-bottom: 1.5rem;
  align-items: center;
}

.search-box {
  flex: 1;
  min-width: 250px;
  position: relative;
  display: flex;
  align-items: center;
}

.search-icon {
  position: absolute;
  left: 0.875rem;
  font-size: 1rem;
  color: #999;
}

.search-input {
  width: 100%;
  padding: 0.625rem 2.5rem 0.625rem 2.5rem;
  border: 1px solid #e5e5e5;
  border-radius: 6px;
  font-size: 0.9375rem;
  transition: all 0.2s;
}

.search-input:focus {
  outline: none;
  border-color: #1a1a1a;
}

.clear-btn {
  position: absolute;
  right: 0.5rem;
  background: #e5e5e5;
  border: none;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  cursor: pointer;
  font-size: 0.75rem;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.clear-btn:hover {
  background: #d4d4d4;
}

.filter-select {
  padding: 0.625rem 0.875rem;
  border: 1px solid #e5e5e5;
  border-radius: 6px;
  font-size: 0.9375rem;
  cursor: pointer;
  background: white;
  transition: all 0.2s;
  min-width: 150px;
}

.filter-select:focus {
  outline: none;
  border-color: #1a1a1a;
}

.clear-filters-btn {
  padding: 0.625rem 1.25rem;
  background: #fef2f2;
  border: 1px solid #fecaca;
  color: #dc2626;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 500;
  font-size: 0.9375rem;
  transition: all 0.2s;
  white-space: nowrap;
}

.clear-filters-btn:hover {
  background: #fee2e2;
  border-color: #fca5a5;
}

/* ============================================ */
/* STATS - 3 CARDS ONLY */
/* ============================================ */
.stats-row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.stat-card {
  background: white;
  border: 1px solid #e5e5e5;
  border-radius: 8px;
  padding: 1rem 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.stat-card.success {
  border-left: 3px solid #22c55e;
}

.stat-card.failed {
  border-left: 3px solid #ef4444;
}

.stat-label {
  font-size: 0.8125rem;
  color: #666;
  font-weight: 500;
}

.stat-value {
  font-size: 1.5rem;
  font-weight: 600;
  color: #1a1a1a;
}

/* ============================================ */
/* STATES */
/* ============================================ */
.loading-state,
.error-state,
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem 2rem;
  gap: 1rem;
  background: white;
  border: 1px solid #e5e5e5;
  border-radius: 8px;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 3px solid #f3f3f3;
  border-top: 3px solid #1a1a1a;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

.error-icon,
.empty-icon {
  font-size: 3rem;
}

.retry-btn {
  padding: 0.625rem 1.25rem;
  background: #1a1a1a;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 500;
  transition: all 0.2s;
}

.retry-btn:hover {
  background: #333;
}

/* ============================================ */
/* TABLE */
/* ============================================ */
.table-container {
  background: white;
  border: 1px solid #e5e5e5;
  border-radius: 8px;
  overflow: hidden;
}

.history-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.875rem;
}

.history-table thead {
  background: #fafafa;
  border-bottom: 1px solid #e5e5e5;
}

.history-table th {
  padding: 0.875rem 1rem;
  text-align: left;
  font-weight: bold;
  color: #000000;
  font-size: 0.8125rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.history-table tbody tr.log-row {
  border-bottom: 1px solid #f0f0f0;
  cursor: pointer;
  transition: all 0.15s;
}

.history-table tbody tr.log-row:hover {
  background: #fafafa;
}

.history-table tbody tr.log-row.expanded {
  background: #f5f5f5;
  border-bottom: none;
}

.history-table tbody tr.log-row.status-success {
  border-left: 3px solid #22c55e;
}

.history-table tbody tr.log-row.status-failed {
  border-left: 3px solid #ef4444;
}

.history-table tbody tr.log-row.status-running {
  border-left: 3px solid #f59e0b;
}

.history-table td {
  padding: 1rem;
  vertical-align: middle;
  font-size: 0.9375rem;
  color: #333;
}

.expand-cell {
  width: 40px;
  padding: 0.5rem;
}

.expand-btn {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 0.75rem;
  color: #666;
  padding: 0.25rem;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.expand-btn:hover {
  color: #1a1a1a;
}

.job-name {
  font-weight: 500;
  color: #1a1a1a;
}

.remote-badge {
  background: #eff6ff;
  color: #1e40af;
  padding: 0.25rem 0.625rem;
  border-radius: 4px;
  font-size: 0.8125rem;
  font-weight: 500;
}

.source-path {
  color: #666;
  font-family: monospace;
  font-size: 0.8125rem;
  max-width: 250px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.status-badge {
  padding: 0.375rem 0.75rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  display: inline-block;
}

.badge-success {
  background: #d1fae5;
  color: #065f46;
}

.badge-failed {
  background: #fee2e2;
  color: #991b1b;
}

.badge-running {
  background: #fef3c7;
  color: #92400e;
}

.badge-default {
  background: #f3f4f6;
  color: #374151;
}

.duration,
.file-size {
  font-family: monospace;
  color: #666;
  font-weight: 500;
}

.timestamp {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}

.time {
  font-weight: 500;
  color: #1a1a1a;
}

.date {
  font-size: 0.75rem;
  color: #999;
}

/* ============================================ */
/* DETAIL ROW */
/* ============================================ */
.detail-row {
  background: #fafafa;
  border-bottom: 1px solid #e5e5e5;
}

.detail-content {
  padding: 1.5rem 1rem;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.detail-section h4 {
  margin: 0 0 0.75rem 0;
  font-size: 0.875rem;
  font-weight: 600;
  color: #1a1a1a;
}

.message-box,
.config-box {
  background: #1f2937;
  color: #e5e7eb;
  padding: 1rem;
  border-radius: 6px;
  font-family: 'Consolas', monospace;
  font-size: 0.8125rem;
  line-height: 1.5;
  margin: 0;
  overflow-x: auto;
  white-space: pre-wrap;
  word-wrap: break-word;
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
  background: white;
  padding: 1rem;
  border-radius: 6px;
  border: 1px solid #e5e5e5;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.detail-label {
  font-size: 0.75rem;
  color: #666;
  font-weight: 600;
  text-transform: uppercase;
}

.detail-value {
  font-size: 0.9375rem;
  color: #1a1a1a;
  font-weight: 500;
}

/* ============================================ */
/* PAGINATION */
/* ============================================ */
.pagination {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem;
  border-top: 1px solid #e5e5e5;
}

.page-btn {
  background: white;
  border: 1px solid #e5e5e5;
  color: #1a1a1a;
  padding: 0.5rem 1rem;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 500;
  transition: all 0.2s;
}

.page-btn:hover:not(:disabled) {
  background: #f5f5f5;
  border-color: #1a1a1a;
}

.page-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.page-info {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  color: #666;
  font-size: 0.875rem;
}

.separator {
  color: #d4d4d4;
}

/* ============================================ */
/* RESPONSIVE */
/* ============================================ */
@media (max-width: 1024px) {
  .stats-row {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (max-width: 768px) {
  .filters-section {
    flex-wrap: wrap;
  }

  .stats-row {
    grid-template-columns: 1fr;
  }

  .table-container {
    overflow-x: auto;
  }

  .history-table {
    min-width: 1000px;
  }
}
</style>