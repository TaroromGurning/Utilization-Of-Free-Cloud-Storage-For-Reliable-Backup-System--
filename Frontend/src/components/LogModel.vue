<script setup>
const props = defineProps({
  isVisible: Boolean,
  log: Object
});

const emit = defineEmits(['close']);

function handleClose() {
  emit('close');
}

// ✅ Helper untuk mendapatkan Job Name
function getJobName(log) {
  if (!log) return 'Unknown Job';

  // 1. PRIORITAS: Cek dari field baru di logs table (untuk one-shot RESTORE)
  if (log.JobName) return log.JobName;
  if (log.job_name) return log.job_name;

  // 2. Fallback: Cek dari relasi ScheduledJob
  if (log.ScheduledJob?.JobName) return log.ScheduledJob.JobName;
  if (log.scheduled_job?.job_name) return log.scheduled_job.job_name;

  // 3. Fallback: Cek dari ConfigSnapshot (backup lama)
  const snapshot = log.ConfigSnapshot || log.config_snapshot;
  if (snapshot) {
    try {
      const config = typeof snapshot === 'string' ? JSON.parse(snapshot) : snapshot;
      return config.job_name || config.JobName || 'Manual Job';
    } catch (e) {
      return 'Manual Job';
    }
  }
  return 'Unknown Job';
}

// ✅ Helper untuk mendapatkan Source Path
function getSourcePath(log) {
  if (!log) return '-';

  // 1. PRIORITAS: Cek dari field baru di logs table (untuk one-shot RESTORE)
  if (log.SourcePath) return log.SourcePath;
  if (log.source_path) return log.source_path;

  // 2. Fallback: Cek dari relasi ScheduledJob
  if (log.ScheduledJob?.SourcePath) return log.ScheduledJob.SourcePath;
  if (log.scheduled_job?.source_path) return log.scheduled_job.source_path;

  // 3. Fallback: Cek dari ConfigSnapshot (backup lama)
  const snapshot = log.ConfigSnapshot || log.config_snapshot;
  if (snapshot) {
    try {
      const config = typeof snapshot === 'string' ? JSON.parse(snapshot) : snapshot;
      return config.source_path || config.SourcePath || config.source || '-';
    } catch (e) {
      console.error("Error parsing snapshot for source path:", e);
      return 'Config Error';
    }
  }

  return '-';
}

// ✅ Helper untuk mendapatkan Transferred Bytes
function getTransferredBytes(log) {
  if (!log) return 0;
  return log.TransferredBytes || log.transferred_bytes || 0;
}

// ✅ Helper untuk Status Class
function getStatusClass(status) {
  if (!status) return 'status-default';
  const s = status.toUpperCase();
  if (['SUCCESS', 'COMPLETED'].includes(s)) return 'status-success';
  if (s.includes('FAIL') || s === 'ERROR') return 'status-failed';
  if (s === 'RUNNING') return 'status-running';
  return 'status-default';
}

// ✅ Format Full Timestamp
function formatFullTimestamp(timestamp) {
  if (!timestamp) return '-';
  try {
    return new Date(timestamp).toLocaleString('id-ID', {
      dateStyle: 'medium',
      timeStyle: 'medium'
    });
  } catch (e) {
    return timestamp;
  }
}

// ✅ Format File Size
function formatFileSize(bytes) {
  const size = Number(bytes);
  if (!size || size === 0 || isNaN(size)) return '0 B';
  
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  const k = 1024;
  const i = Math.floor(Math.log(size) / Math.log(k));
  
  return parseFloat((size / Math.pow(k, i)).toFixed(2)) + ' ' + units[i];
}

// ✅ Format JSON
function formatJSON(jsonInput) {
  if (!jsonInput) return '';
  try {
    const obj = typeof jsonInput === 'string' ? JSON.parse(jsonInput) : jsonInput;
    return JSON.stringify(obj, null, 2);
  } catch (e) {
    return jsonInput;
  }
}
</script>

<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="isVisible && log" class="modal-overlay" @click.self="handleClose">
        <div class="modal-container">
          <div class="modal-header">
            <div class="header-left">
              <h3>Log Details</h3>
              <span class="log-id">ID: {{ log.ID }}</span>
            </div>
            <button @click="handleClose" class="close-btn">×</button>
          </div>

          <div class="modal-body">
            <div class="info-section">
              <h4>Job Information</h4>
              <div class="info-grid">
                <div class="info-item">
                  <span class="label">Job Name</span>
                  <span class="value">{{ getJobName(log) }}</span>
                </div>
                
                <div class="info-item">
                  <span class="label">Object (Source Path)</span>
                  <span class="value truncate" :title="getSourcePath(log)">{{ getSourcePath(log) }}</span>
                </div>

                <div class="info-item">
                  <span class="label">Status</span>
                  <span class="value">
                    <span class="status-badge" :class="getStatusClass(log.Status || log.status)">
                      {{ log.Status || log.status }}
                    </span>
                  </span>
                </div>
                <div class="info-item">
                  <span class="label">Duration</span>
                  <span class="value">{{ log.DurationSec || log.duration_sec || 0 }} seconds</span>
                </div>
                
                <div class="info-item">
                  <span class="label">Transferred Size</span>
                  <span class="value">{{ formatFileSize(getTransferredBytes(log)) }}</span>
                </div>

                <div class="info-item">
                  <span class="label">Timestamp</span>
                  <span class="value">{{ formatFullTimestamp(log.Timestamp || log.timestamp) }}</span>
                </div>
              </div>
            </div>

            <div class="info-section">
              <h4>Output Message</h4>
              <div class="output-box">
                <pre>{{ log.Message || log.message || 'No message available' }}</pre>
              </div>
            </div>

            <div v-if="log.ConfigSnapshot || log.config_snapshot" class="info-section">
              <h4>Configuration Snapshot</h4>
              <div class="config-box">
                <pre>{{ formatJSON(log.ConfigSnapshot || log.config_snapshot) }}</pre>
              </div>
            </div>
          </div>

          <div class="modal-footer">
            <button @click="handleClose" class="btn-close">Close</button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  padding: 20px;
}

.modal-container {
  background: #fff;
  border-radius: 8px;
  border: 1px solid #e5e5e5;
  width: 90%;
  max-width: 900px;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.25rem 1.5rem;
  background: #fafafa;
  border-bottom: 1px solid #e5e5e5;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.modal-header h3 {
  margin: 0;
  font-size: 1.125rem;
  font-weight: 600;
  color: #1a1a1a;
}

.log-id {
  background: #f0f0f0;
  padding: 0.25rem 0.625rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 500;
  color: #666;
}

.close-btn {
  background: transparent;
  border: 1px solid #e5e5e5;
  color: #666;
  width: 32px;
  height: 32px;
  border-radius: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
  font-size: 1.5rem;
  line-height: 1;
}

.close-btn:hover {
  background: #f5f5f5;
  border-color: #1a1a1a;
  color: #1a1a1a;
}

.modal-body {
  flex: 1;
  overflow-y: auto;
  padding: 1.5rem;
  background: #fafafa;
}

.modal-body::-webkit-scrollbar {
  width: 8px;
}

.modal-body::-webkit-scrollbar-track {
  background: #f0f0f0;
}

.modal-body::-webkit-scrollbar-thumb {
  background: #d4d4d4;
  border-radius: 4px;
}

.modal-body::-webkit-scrollbar-thumb:hover {
  background: #a3a3a3;
}

.info-section {
  background: white;
  border-radius: 6px;
  border: 1px solid #e5e5e5;
  padding: 1.25rem;
  margin-bottom: 1rem;
}

.info-section:last-child {
  margin-bottom: 0;
}

.info-section h4 {
  margin: 0 0 1rem 0;
  font-size: 0.875rem;
  font-weight: 600;
  color: #1a1a1a;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid #e5e5e5;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 1rem;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.info-item .label {
  font-size: 0.75rem;
  color: #666;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.info-item .value {
  font-size: 0.9375rem;
  color: #1a1a1a;
  font-weight: 500;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  padding: 0.25rem 0.625rem;
  border-radius: 4px;
  font-weight: 600;
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.status-success { background: #d1fae5; color: #065f46; }
.status-failed { background: #fee2e2; color: #991b1b; }
.status-running { background: #fef3c7; color: #92400e; }
.status-default { background: #e5e7eb; color: #374151; }

.output-box, .config-box {
  background: #1a1a1a;
  border-radius: 6px;
  padding: 1rem;
  overflow-x: auto;
}

.output-box pre, .config-box pre {
  margin: 0;
  color: #d4d4d4;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 0.8125rem;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-word;
}

.modal-footer {
  padding: 1rem 1.5rem;
  background: white;
  border-top: 1px solid #e5e5e5;
  display: flex;
  justify-content: flex-end;
}

.btn-close {
  background: #1a1a1a;
  color: white;
  border: none;
  padding: 0.625rem 1.5rem;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 500;
  font-size: 0.9375rem;
  transition: all 0.2s;
}

.btn-close:hover {
  background: #333;
}

.modal-enter-active, .modal-leave-active {
  transition: opacity 0.2s ease;
}

.modal-enter-from, .modal-leave-to {
  opacity: 0;
}

.modal-enter-active .modal-container, .modal-leave-active .modal-container {
  transition: transform 0.2s ease;
}

.modal-enter-from .modal-container, .modal-leave-to .modal-container {
  transform: scale(0.95);
}

.truncate {
  max-width: 220px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  cursor: help;
  display: inline-block;
}

.truncate:hover::after {
  content: attr(title);
  position: absolute;
  background: #333;
  color: white;
  padding: 8px 12px;
  border-radius: 4px;
  font-size: 0.8rem;
  z-index: 1000;
  white-space: normal;
  max-width: 400px;
  word-wrap: break-word;
  margin-top: 5px;
  margin-left: -50px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  font-weight: normal;
}

@media (max-width: 768px) {
  .modal-container {
    width: 95%;
    max-height: 85vh;
  }

  .modal-header {
    padding: 1rem 1.25rem;
  }

  .header-left {
    flex-wrap: wrap;
    gap: 0.75rem;
  }

  .modal-header h3 {
    font-size: 1rem;
  }

  .modal-body {
    padding: 1rem;
  }

  .info-grid {
    grid-template-columns: 1fr;
  }

  .info-section {
    padding: 1rem;
  }
}
</style>