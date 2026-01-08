<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="isVisible && log" class="modal-overlay" @click.self="handleClose">
        <div class="modal-container">
          <div class="modal-header">
            <div class="header-left">
              <h3>Log Details</h3>
              <span class="log-id">ID: {{ log.id || log.ID }}</span>
            </div>
            <button @click="handleClose" class="close-btn">×</button>
          </div>

          <div class="modal-body">
            <div class="info-section">
              <h4>Job Information</h4>
              <div class="info-grid">
                <div class="info-item">
                  <span class="label">Object</span>
                  <span class="value">{{ getSourceObject(log) }}</span>
                </div>
                <div class="info-item">
                  <span class="label">Job Name</span>
                  <span class="value">{{ getJobName(log) }}</span>
                </div>
                <div class="info-item">
                  <span class="label">Status</span>
                  <span class="value">
                    <span class="status-badge" :class="getStatusClass(log.status || log.Status)">
                      {{ log.status || log.Status }}
                    </span>
                  </span>
                </div>
                <div class="info-item">
                  <span class="label">Duration</span>
                  <span class="value">{{ log.duration_sec || log.DurationSec || 0 }} seconds</span>
                </div>
                
                <div class="info-item">
                  <span class="label">Transferred Size</span>
                  <span class="value">{{ formatFileSize(log.transferred_bytes || log.TransferredBytes) }}</span>
                </div>

                <div class="info-item">
                  <span class="label">Timestamp</span>
                  <span class="value">{{ formatFullTimestamp(log.timestamp || log.Timestamp) }}</span>
                </div>
              </div>
            </div>

            <div class="info-section">
              <h4>Output Message</h4>
              <div class="output-box">
                <pre>{{ log.message || log.Message || 'No message available' }}</pre>
              </div>
            </div>

            <div v-if="log.config_snapshot || log.ConfigSnapshot" class="info-section">
              <h4>Configuration Snapshot</h4>
              <div class="config-box">
                <pre>{{ formatJSON(log.config_snapshot || log.ConfigSnapshot) }}</pre>
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

<script setup>
import { computed } from 'vue';

const props = defineProps({
  isVisible: Boolean,
  log: Object
});

const emit = defineEmits(['close']);

function handleClose() {
  emit('close');
}

// ✅ Helper untuk mendapatkan source path
function getSourceObject(log) {
  // 1. PRIORITAS UTAMA: Cek dari field baru di logs table
  if (log.source_path) return log.source_path;
  if (log.SourcePath) return log.SourcePath;

  // 2. Fallback: Cek dari relasi ScheduledJob
  if (log.scheduled_job?.source_path) return log.scheduled_job.source_path;
  if (log.ScheduledJob?.SourcePath) return log.ScheduledJob.SourcePath;

  // 3. Fallback: Cek dari ConfigSnapshot (backup lama)
  const snapshot = log.config_snapshot || log.ConfigSnapshot;
  if (snapshot) {
    try {
      const config = typeof snapshot === 'string' ? JSON.parse(snapshot) : snapshot;
      return config.source_path || config.SourcePath || '-';
    } catch (e) {
      return '-';
    }
  }
  return '-';
}

// ✅ Helper untuk mendapatkan job name
function getJobName(log) {
  // 1. PRIORITAS UTAMA: Cek dari field baru di logs table
  if (log.job_name) return log.job_name;
  if (log.JobName) return log.JobName;

  // 2. Fallback: Cek dari relasi ScheduledJob
  if (log.scheduled_job?.job_name) return log.scheduled_job.job_name;
  if (log.ScheduledJob?.JobName) return log.ScheduledJob.JobName;

  // 3. Fallback: Cek dari ConfigSnapshot (backup lama)
  const snapshot = log.config_snapshot || log.ConfigSnapshot;
  if (snapshot) {
    try {
      const config = typeof snapshot === 'string' ? JSON.parse(snapshot) : snapshot;
      return config.job_name || config.JobName || 'Unknown Job';
    } catch (e) {
      return 'Unknown Job';
    }
  }
  return 'Unknown Job';
}

function getStatusClass(status) {
  if (!status) return 'status-default';
  const s = status.toUpperCase();
  if (['SUCCESS', 'COMPLETED'].includes(s)) return 'status-success';
  if (s.includes('FAIL') || s === 'ERROR') return 'status-failed';
  if (s === 'RUNNING') return 'status-running';
  return 'status-default';
}

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

function formatFileSize(bytes) {
  // Konversi ke number untuk memastikan keamanan
  const size = Number(bytes);
  if (!size || size === 0 || isNaN(size)) return '0 B';
  
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  const k = 1024;
  const i = Math.floor(Math.log(size) / Math.log(k));
  
  return parseFloat((size / Math.pow(k, i)).toFixed(2)) + ' ' + units[i];
}

function formatJSON(jsonInput) {
  if (!jsonInput) return '';
  try {
    // Jika input sudah object, stringify langsung. Jika string, parse dulu.
    const obj = typeof jsonInput === 'string' ? JSON.parse(jsonInput) : jsonInput;
    return JSON.stringify(obj, null, 2);
  } catch (e) {
    return jsonInput;
  }
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(2px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-container {
  background-color: white;
  border-radius: 12px;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
  max-width: 800px;
  width: 90%;
  max-height: 85vh;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid #f3f4f6;
  background: #fff;
  position: sticky;
  top: 0;
  z-index: 10;
}

.header-left {
  display: flex;
  align-items: baseline;
  gap: 1rem;
}

.header-left h3 {
  margin: 0;
  font-size: 1.25rem;
  color: #111827;
  font-weight: 600;
}

.log-id {
  font-size: 0.875rem;
  color: #6b7280;
  font-family: monospace;
  background: #f3f4f6;
  padding: 2px 6px;
  border-radius: 4px;
}

.close-btn {
  background: none;
  border: none;
  font-size: 1.5rem;
  color: #9ca3af;
  cursor: pointer;
  transition: color 0.2s;
  line-height: 1;
}

.close-btn:hover {
  color: #ef4444;
}

.modal-body {
  padding: 1.5rem;
  overflow-y: auto;
}

.info-section {
  margin-bottom: 2rem;
}

.info-section h4 {
  margin: 0 0 1rem 0;
  font-size: 0.95rem;
  color: #374151;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1.5rem;
  background: #f9fafb;
  padding: 1.25rem;
  border-radius: 8px;
  border: 1px solid #e5e7eb;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.label {
  font-size: 0.75rem;
  color: #6b7280;
  font-weight: 600;
  text-transform: uppercase;
}

.value {
  font-size: 0.9375rem;
  color: #111827;
  font-weight: 500;
  word-break: break-all;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  padding: 0.25rem 0.75rem;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 600;
}

.status-success { background-color: #ecfdf5; color: #047857; }
.status-failed { background-color: #fef2f2; color: #b91c1c; }
.status-running { background-color: #eff6ff; color: #1d4ed8; }
.status-default { background-color: #f3f4f6; color: #374151; }

.output-box,
.config-box {
  background-color: #1f2937;
  border-radius: 8px;
  padding: 1rem;
  overflow-x: auto;
  border: 1px solid #374151;
}

.output-box pre,
.config-box pre {
  margin: 0;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 0.85rem;
  color: #e5e7eb;
  line-height: 1.6;
  white-space: pre-wrap;
  word-wrap: break-word;
}

.modal-footer {
  padding: 1rem 1.5rem;
  border-top: 1px solid #f3f4f6;
  display: flex;
  justify-content: flex-end;
  background: #fff;
}

.btn-close {
  padding: 0.625rem 1.25rem;
  background-color: white;
  border: 1px solid #d1d5db;
  color: #374151;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 500;
  transition: all 0.2s;
}

.btn-close:hover {
  background-color: #f3f4f6;
  border-color: #9ca3af;
}

.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.2s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

@media (max-width: 640px) {
  .modal-container {
    width: 100%;
    height: 100%;
    max-height: 100%;
    border-radius: 0;
  }
  
  .info-grid {
    grid-template-columns: 1fr;
    gap: 1rem;
  }
}
</style>