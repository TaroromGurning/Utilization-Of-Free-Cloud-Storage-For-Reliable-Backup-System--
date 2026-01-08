<template>
  <div class="logs-view">
    <!-- Page Header -->
    <div class="page-header">
      <div>
        <h1>📝 Execution Logs</h1>
        <p>Riwayat semua Job (Auto, Manual, Restore) yang telah dieksekusi.</p>
      </div>
      <button @click="fetchData" class="refresh-btn" :class="{ spinning: isRefreshing }">
        🔄 Refresh
      </button>
    </div>

    <!-- Filter & Search Bar -->
    <div class="filter-bar">
      <div class="search-box">
        <span class="search-icon">🔍</span>
        <input 
          v-model="searchQuery"
          type="text" 
          placeholder="Search by job name or message..."
          class="search-input"
        />
        <button v-if="searchQuery" @click="searchQuery = ''" class="clear-btn">
          ✕
        </button>
      </div>

      <div class="filters">
        <select v-model="statusFilter" class="filter-select">
          <option value="">All Status</option>
          <option value="SUCCESS">Success</option>
          <option value="COMPLETED">Completed</option>
          <option value="RUNNING">Running</option>
          <option value="PENDING">Pending</option>
          <option value="FAILED">Failed</option>
        </select>

        <select v-model="limitFilter" class="filter-select">
          <option :value="20">Show 20</option>
          <option :value="50">Show 50</option>
          <option :value="100">Show 100</option>
          <option :value="0">Show All</option>
        </select>
      </div>
    </div>

    <!-- Stats Cards -->
    <div class="stats-cards">
      <div class="stat-card success">
        <div class="stat-icon">✅</div>
        <div class="stat-content">
          <span class="stat-label">Success</span>
          <span class="stat-value">{{ statsCount.success }}</span>
        </div>
      </div>

      <div class="stat-card failed">
        <div class="stat-icon">❌</div>
        <div class="stat-content">
          <span class="stat-label">Failed</span>
          <span class="stat-value">{{ statsCount.failed }}</span>
        </div>
      </div>

      <div class="stat-card running">
        <div class="stat-icon">⏳</div>
        <div class="stat-content">
          <span class="stat-label">Running</span>
          <span class="stat-value">{{ statsCount.running }}</span>
        </div>
      </div>

      <div class="stat-card total">
        <div class="stat-icon">📊</div>
        <div class="stat-content">
          <span class="stat-label">Total Logs</span>
          <span class="stat-value">{{ statsCount.total }}</span>
        </div>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="isLoading" class="loading-state">
      <div class="spinner"></div>
      <p>Memuat data logs...</p>
    </div>

    <!-- Error State -->
    <div v-else-if="errorMessage" class="error-box">
      <span class="error-icon">⚠️</span>
      <div>
        <strong>Error:</strong> {{ errorMessage }}
        <button @click="fetchData" class="retry-btn">🔄 Coba Lagi</button>
      </div>
    </div>

    <!-- Empty State -->
    <div v-else-if="filteredLogs.length === 0" class="empty-state">
      <div class="empty-icon">📋</div>
      <h3>{{ searchQuery || statusFilter ? 'No Results Found' : 'No Logs Yet' }}</h3>
      <p>
        {{ searchQuery || statusFilter 
          ? 'Try adjusting your filters or search query' 
          : 'Belum ada eksekusi Job yang tercatat.' 
        }}
      </p>
      <button v-if="searchQuery || statusFilter" @click="clearFilters" class="clear-filters-btn">
        Clear Filters
      </button>
    </div>

    <!-- Logs Table -->
    <div v-else class="table-container">
      <div class="table-info">
        <span>Showing {{ displayedLogs.length }} of {{ filteredLogs.length }} logs</span>
      </div>

      <table class="logs-table">
        <thead>
          <tr>
            <th>Job ID</th>
            <th>Job Name</th>
            <th>Operation</th>
            <th>Status</th>
            <th>Duration</th>
            <th>Timestamp</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr 
            v-for="log in displayedLogs" 
            :key="log.ID"
            :class="getLogRowClass(log.Status)"
          >
            <td>
              <span class="job-id-badge">
                {{ log.JobID || 'MANUAL' }}
              </span>
            </td>

            <td>
              <strong class="job-name">{{ getJobName(log) }}</strong>
            </td>

            <td>
              <span class="operation-badge">
                {{ log.OperationType || 'N/A' }}
              </span>
            </td>

            <td>
              <span class="status-badge" :class="getStatusClass(log.Status)">
                <span class="status-icon">{{ getStatusIcon(log.Status) }}</span>
                {{ log.Status }}
              </span>
            </td>

            <td>
              <span class="duration">
                {{ formatDuration(log.DurationSec) }}
              </span>
            </td>

            <td>
              <div class="timestamp-cell">
                <span class="time">{{ formatTime(log.Timestamp) }}</span>
                <span class="date">{{ formatDate(log.Timestamp) }}</span>
              </div>
            </td>

            <td>
              <button 
                @click="viewLogDetail(log)" 
                class="action-btn view-btn"
                title="View details"
              >
                👁️
              </button>
            </td>
          </tr>
        </tbody>
      </table>

      <!-- Pagination (if needed) -->
      <div v-if="filteredLogs.length > displayedLogs.length" class="load-more">
        <button @click="loadMore" class="load-more-btn">
          Load More Logs
        </button>
      </div>
    </div>

    <!-- Log Detail Modal -->
    <LogDetailModal
      :is-visible="showDetailModal"
      :log="selectedLog"
      @close="closeDetailModal"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue';
import monitoringService from '@/services/monitoringService';
import LogDetailModal from '@/components/LogDetail.vue';

const logs = ref([]);
const isLoading = ref(true);
const isRefreshing = ref(false);
const errorMessage = ref(null);

// Filter states
const searchQuery = ref('');
const statusFilter = ref('');
const limitFilter = ref(50);

// Modal state
const showDetailModal = ref(false);
const selectedLog = ref(null);

// Fetch data
async function fetchData() {
  if (isRefreshing.value) return;
  
  isRefreshing.value = true;
  isLoading.value = logs.value.length === 0; // Show spinner only on first load
  errorMessage.value = null;

  try {
    const data = await monitoringService.getLogs();
    logs.value = Array.isArray(data) ? data : [];
  } catch (error) {
    errorMessage.value = error.response?.data?.error || 'Gagal memuat data logs.';
    console.error('Fetch Logs Error:', error);
  } finally {
    isLoading.value = false;
    isRefreshing.value = false;
  }
}

onMounted(fetchData);

// Filtered logs based on search and status
const filteredLogs = computed(() => {
  let result = [...logs.value];

  // Filter by search query
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase();
    result = result.filter(log => {
      const jobName = getJobName(log).toLowerCase();
      const message = (log.Message || '').toLowerCase();
      const operationType = (log.OperationType || '').toLowerCase();
      
      return jobName.includes(query) || 
             message.includes(query) || 
             operationType.includes(query);
    });
  }

  // Filter by status
  if (statusFilter.value) {
    result = result.filter(log => 
      log.Status.toUpperCase() === statusFilter.value.toUpperCase()
    );
  }

  // Sort by timestamp (newest first)
  result.sort((a, b) => new Date(b.Timestamp) - new Date(a.Timestamp));

  return result;
});

// Displayed logs with limit
const displayedLogs = computed(() => {
  if (limitFilter.value === 0) {
    return filteredLogs.value;
  }
  return filteredLogs.value.slice(0, limitFilter.value);
});

// Stats count
const statsCount = computed(() => {
  const total = logs.value.length;
  const success = logs.value.filter(l => 
    ['SUCCESS', 'COMPLETED'].includes(l.Status.toUpperCase())
  ).length;
  const failed = logs.value.filter(l => 
    l.Status.toUpperCase().includes('FAIL') || l.Status.toUpperCase() === 'ERROR'
  ).length;
  const running = logs.value.filter(l => 
    l.Status.toUpperCase() === 'RUNNING'
  ).length;

  return { total, success, failed, running };
});

// Helper functions
function getJobName(log) {
  if (log.ScheduledJob?.JobName) {
    return log.ScheduledJob.JobName;
  }
  if (log.ConfigSnapshot) {
    try {
      const config = JSON.parse(log.ConfigSnapshot);
      return config.job_name || 'Manual Job';
    } catch (e) {
      return 'Manual Job';
    }
  }
  return 'Unknown Job';
}

function getLogRowClass(status) {
  const statusUpper = status.toUpperCase();
  if (['SUCCESS', 'COMPLETED'].includes(statusUpper)) return 'row-success';
  if (statusUpper.includes('FAIL') || statusUpper === 'ERROR') return 'row-failed';
  if (statusUpper === 'RUNNING') return 'row-running';
  return '';
}

function getStatusClass(status) {
  const statusUpper = status.toUpperCase();
  if (['SUCCESS', 'COMPLETED'].includes(statusUpper)) return 'status-success';
  if (statusUpper.includes('FAIL') || statusUpper === 'ERROR') return 'status-failed';
  if (statusUpper === 'RUNNING') return 'status-running';
  if (statusUpper === 'PENDING') return 'status-pending';
  return 'status-default';
}

function getStatusIcon(status) {
  const statusUpper = status.toUpperCase();
  if (['SUCCESS', 'COMPLETED'].includes(statusUpper)) return '✅';
  if (statusUpper.includes('FAIL') || statusUpper === 'ERROR') return '❌';
  if (statusUpper === 'RUNNING') return '⏳';
  if (statusUpper === 'PENDING') return '⏱️';
  return '📝';
}

function formatDuration(seconds) {
  if (!seconds || seconds === 0) return '-';
  
  if (seconds < 60) {
    return `${seconds}s`;
  }
  
  const minutes = Math.floor(seconds / 60);
  const secs = seconds % 60;
  
  return `${minutes}m ${secs}s`;
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

function clearFilters() {
  searchQuery.value = '';
  statusFilter.value = '';
}

function loadMore() {
  limitFilter.value += 50;
}

function viewLogDetail(log) {
  selectedLog.value = log;
  showDetailModal.value = true;
}

function closeDetailModal() {
  showDetailModal.value = false;
  selectedLog.value = null;
}
</script>

<style scoped>
.logs-view {
  padding: 2rem;
  max-width: 100%;
  margin: 0 auto;
}

/* Page Header */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 2rem;
  flex-wrap: wrap;
  gap: 1rem;
}

.page-header h1 {
  font-size: 2rem;
  font-weight: 700;
  color: #2c3e50;
  margin: 0 0 0.5rem 0;
}

.page-header p {
  color: #6c757d;
  margin: 0;
}

.refresh-btn {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  padding: 0.75rem 1.5rem;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 600;
  transition: all 0.3s;
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

.refresh-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(102, 126, 234, 0.4);
}

.refresh-btn.spinning {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* Filter Bar */
.filter-bar {
  display: flex;
  gap: 1rem;
  margin-bottom: 1.5rem;
  flex-wrap: wrap;
}

.search-box {
  flex: 1;
  min-width: 300px;
  position: relative;
  display: flex;
  align-items: center;
}

.search-icon {
  position: absolute;
  left: 1rem;
  font-size: 1.2rem;
  color: #6c757d;
}

.search-input {
  width: 100%;
  padding: 0.75rem 1rem 0.75rem 3rem;
  border: 2px solid #e9ecef;
  border-radius: 8px;
  font-size: 1rem;
  transition: all 0.2s;
}

.search-input:focus {
  outline: none;
  border-color: #667eea;
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

.clear-btn {
  position: absolute;
  right: 0.5rem;
  background: #e9ecef;
  border: none;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.clear-btn:hover {
  background: #dee2e6;
}

.filters {
  display: flex;
  gap: 0.75rem;
}

.filter-select {
  padding: 0.75rem 1rem;
  border: 2px solid #e9ecef;
  border-radius: 8px;
  font-size: 0.9rem;
  cursor: pointer;
  transition: all 0.2s;
  background: white;
}

.filter-select:focus {
  outline: none;
  border-color: #667eea;
}

/* Stats Cards */
.stats-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
  margin-bottom: 2rem;
}

.stat-card {
  background: white;
  padding: 1.25rem;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
  display: flex;
  align-items: center;
  gap: 1rem;
  border-left: 4px solid;
  transition: all 0.2s;
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.stat-card.success { border-left-color: #28a745; }
.stat-card.failed { border-left-color: #dc3545; }
.stat-card.running { border-left-color: #ffc107; }
.stat-card.total { border-left-color: #667eea; }

.stat-icon {
  font-size: 2rem;
  line-height: 1;
}

.stat-content {
  display: flex;
  flex-direction: column;
}

.stat-label {
  font-size: 0.85rem;
  color: #6c757d;
  font-weight: 500;
}

.stat-value {
  font-size: 1.75rem;
  font-weight: 700;
  color: #2c3e50;
}

/* Loading, Error, Empty States */
.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 4rem 2rem;
  gap: 1rem;
}

.spinner {
  width: 48px;
  height: 48px;
  border: 4px solid #f3f3f3;
  border-top: 4px solid #667eea;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

.error-box {
  background: #fee;
  border: 1px solid #fcc;
  border-radius: 8px;
  padding: 1.5rem;
  display: flex;
  align-items: center;
  gap: 1rem;
}

.error-icon {
  font-size: 2rem;
}

.retry-btn {
  background: #dc3545;
  color: white;
  border: none;
  padding: 0.5rem 1rem;
  border-radius: 6px;
  cursor: pointer;
  margin-left: 1rem;
  transition: all 0.2s;
}

.retry-btn:hover {
  background: #c82333;
}

.empty-state {
  text-align: center;
  padding: 4rem 2rem;
}

.empty-icon {
  font-size: 4rem;
  opacity: 0.3;
  margin-bottom: 1rem;
}

.empty-state h3 {
  font-size: 1.5rem;
  color: #2c3e50;
  margin-bottom: 0.5rem;
}

.empty-state p {
  color: #6c757d;
  margin-bottom: 2rem;
}

.clear-filters-btn {
  background: #667eea;
  color: white;
  border: none;
  padding: 0.75rem 1.5rem;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 600;
  transition: all 0.2s;
}

.clear-filters-btn:hover {
  background: #5568d3;
}

/* Table Container */
.table-container {
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
  overflow: hidden;
}

.table-info {
  padding: 1rem 1.25rem;
  background: #f8f9fa;
  border-bottom: 1px solid #e9ecef;
  font-size: 0.9rem;
  color: #6c757d;
  font-weight: 500;
}

.logs-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.9rem;
}

.logs-table thead {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.logs-table th {
  padding: 1rem 1.25rem;
  text-align: left;
  font-weight: 600;
  letter-spacing: 0.5px;
  text-transform: uppercase;
  font-size: 0.8rem;
}

.logs-table td {
  padding: 1rem 1.25rem;
  border-bottom: 1px solid #f0f0f0;
  vertical-align: middle;
}

.logs-table tbody tr {
  transition: all 0.2s;
}

.logs-table tbody tr:hover {
  background-color: #f8f9fa;
}

.logs-table tbody tr:last-child td {
  border-bottom: none;
}

/* Row colors based on status */
.row-success {
  background: rgba(40, 167, 69, 0.02);
}

.row-failed {
  background: rgba(220, 53, 69, 0.02);
}

.row-running {
  background: rgba(255, 193, 7, 0.02);
}

/* Table Cell Styles */
.job-id-badge {
  background: #e9ecef;
  padding: 4px 10px;
  border-radius: 6px;
  font-family: monospace;
  font-size: 0.85rem;
  font-weight: 600;
  color: #495057;
}

.job-name {
  color: #2c3e50;
  font-size: 0.95rem;
}

.operation-badge {
  background: #667eea;
  color: white;
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 0.8rem;
  font-weight: 600;
  text-transform: uppercase;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 6px 12px;
  border-radius: 20px;
  font-weight: 600;
  font-size: 0.8rem;
  text-transform: uppercase;
}

.status-success {
  background: #d4edda;
  color: #155724;
}

.status-failed {
  background: #f8d7da;
  color: #721c24;
}

.status-running {
  background: #fff3cd;
  color: #856404;
}

.status-pending {
  background: #e2e3e5;
  color: #383d41;
}

.duration {
  font-family: monospace;
  font-weight: 600;
  color: #495057;
}

.timestamp-cell {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.time {
  font-weight: 600;
  color: #2c3e50;
}

.date {
  font-size: 0.85rem;
  color: #6c757d;
}

.action-btn {
  background: #667eea;
  color: white;
  border: none;
  padding: 0.5rem 0.75rem;
  border-radius: 6px;
  cursor: pointer;
  font-size: 1.1rem;
  transition: all 0.2s;
}

.action-btn:hover {
  background: #5568d3;
  transform: scale(1.1);
}

/* Load More */
.load-more {
  padding: 1.5rem;
  text-align: center;
  background: #f8f9fa;
  border-top: 1px solid #e9ecef;
}

.load-more-btn {
  background: white;
  border: 2px solid #667eea;
  color: #667eea;
  padding: 0.75rem 2rem;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 600;
  transition: all 0.2s;
}

.load-more-btn:hover {
  background: #667eea;
  color: white;
}

/* Responsive */
@media (max-width: 1024px) {
  .logs-view {
    padding: 1rem;
  }
  
  .filter-bar {
    flex-direction: column;
  }
  
  .search-box {
    min-width: 100%;
  }
  
  .filters {
    width: 100%;
  }
  
  .filter-select {
    flex: 1;
  }
  
  .table-container {
    overflow-x: auto;
  }
}
</style>