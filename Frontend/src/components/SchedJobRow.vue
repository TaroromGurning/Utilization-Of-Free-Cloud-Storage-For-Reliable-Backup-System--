<template>
  <tr class="job-row" :class="getRowClass(job.status)">
    <td>
      <div class="job-name">
        <strong>{{ job.job_name }}</strong>
      </div>
    </td>
    
    <td>
      <div class="source-path" :title="job.source_path">
        {{ job.source_path }}
      </div>
    </td>
    
    <td>
      <span class="remote-badge">
        {{ job.gdrive_target }}
      </span>
    </td>
    
    <td>
      <span class="timestamp">{{ job.last_run || 'N/A' }}</span>
    </td>

    <td>
      <span class="status-badge" :class="getStatusBadgeClass(job.status)">
        {{ job.status || 'PENDING' }}
      </span>
    </td>

    <td>
      <span class="timestamp">{{ job.next_run || 'N/A' }}</span>
    </td>
    
    <td class="actions-cell">
      <div class="actions">
        <!-- Run/Trigger Button -->
        <button 
          @click="$emit('trigger', job.id)" 
          class="action-btn trigger-btn" 
          title="Run Job Now"
        >
          ▶
        </button>

        <!-- Edit Button -->
        <button 
          @click="$emit('edit', job.id)" 
          class="action-btn edit-btn" 
          title="Edit Job"
        >
          ✎
        </button>

        <!-- View Script Button -->
        <button 
          @click="$emit('view-script', job.id)" 
          class="action-btn view-btn" 
          title="View Script"
        >
          (i)
        </button>

        <!-- Delete Button -->
        <button 
          @click="handleDelete" 
          class="action-btn delete-btn" 
          title="Delete Job"
        >
          ✕
        </button>
      </div>
    </td>
  </tr>
</template>

<script setup>
const props = defineProps({
  job: {
    type: Object,
    required: true
  }
});

const emit = defineEmits(['trigger', 'view-script', 'edit', 'delete']);

function handleDelete() {
  if (!confirm(`PERINGATAN! Menghapus Job Terjadwal '${props.job.job_name}' juga akan menghapus jadwal CRON.\nApakah Anda yakin ingin menghapus job ini?`)) {
    return;
  }
  emit('delete', props.job.id, props.job.job_name);
}

function getRowClass(status) {
  if (!status) return '';
  const s = status.toUpperCase();
  if (['COMPLETED', 'SUCCESS'].includes(s)) return 'row-success';
  if (s.includes('FAIL') || s === 'ERROR') return 'row-failed';
  if (s === 'RUNNING') return 'row-running';
  return '';
}

function getStatusBadgeClass(status) {
  if (!status) return 'badge-pending';
  const s = status.toUpperCase();
  if (['COMPLETED', 'SUCCESS'].includes(s)) return 'badge-success';
  if (s.includes('FAIL') || s === 'ERROR') return 'badge-failed';
  if (s === 'RUNNING') return 'badge-running';
  if (s === 'PENDING') return 'badge-pending';
  return 'badge-default';
}
</script>

<style scoped>
/* ============================================ */
/* TABLE ROW */
/* ============================================ */
.job-row {
  border-bottom: 1px solid #f0f0f0;
  transition: all 0.15s;
  cursor: default;
}

.job-row:hover {
  background: #fafafa;
}

.job-row.row-success {
  border-left: 3px solid #22c55e;
}

.job-row.row-failed {
  border-left: 3px solid #ef4444;
}

.job-row.row-running {
  border-left: 3px solid #f59e0b;
}

/* ============================================ */
/* TABLE CELLS */
/* ============================================ */
td {
  padding: 1rem;
  vertical-align: middle;
  font-size: 0.9375rem;
  color: #333;
  border-bottom: 1px solid #f0f0f0;
}

/* ============================================ */
/* JOB NAME */
/* ============================================ */
.job-name strong {
  font-weight: 500;
  color: #1a1a1a;
}

/* ============================================ */
/* SOURCE PATH */
/* ============================================ */
.source-path {
  color: #666;
  font-family: monospace;
  font-size: 0.8125rem;
  max-width: 250px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* ============================================ */
/* REMOTE BADGE */
/* ============================================ */
.remote-badge {
  background: #eff6ff;
  color: #1e40af;
  padding: 0.25rem 0.625rem;
  border-radius: 4px;
  font-size: 0.8125rem;
  font-weight: 500;
  display: inline-block;
}

/* ============================================ */
/* TIMESTAMP */
/* ============================================ */
.timestamp {
  color: #666;
  font-size: 0.875rem;
  font-family: monospace;
}

/* ============================================ */
/* STATUS BADGE */
/* ============================================ */
.status-badge {
  padding: 0.375rem 0.75rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  display: inline-block;
  white-space: nowrap;
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

.badge-pending {
  background: #f3f4f6;
  color: #6b7280;
}

.badge-default {
  background: #f3f4f6;
  color: #374151;
}

/* ============================================ */
/* ACTIONS CELL */
/* ============================================ */
.actions-cell {
  width: auto;
  text-align: center;
  padding: 0.5rem;
}

.actions {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 0.5rem;
}

/* ============================================ */
/* ACTION BUTTONS */
/* ============================================ */
.action-btn {
  background: white;
  border: 1px solid #e5e5e5;
  border-radius: 4px;
  padding: 0.375rem 0.5rem;
  cursor: pointer;
  font-size: 0.875rem;
  transition: all 0.2s;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 32px;
  height: 32px;
  line-height: 1;
}

.action-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.action-btn:active {
  transform: translateY(0);
}

/* Button Variants */
.trigger-btn {
  color: #22c55e;
  border-color: #22c55e;
}

.trigger-btn:hover {
  background: #f0fdf4;
  border-color: #16a34a;
  color: #16a34a;
}

.edit-btn {
  color: #f59e0b;
  border-color: #f59e0b;
}

.edit-btn:hover {
  background: #fffbeb;
  border-color: #d97706;
  color: #d97706;
}

.view-btn {
  color: #3b82f6;
  border-color: #3b82f6;
}

.view-btn:hover {
  background: #eff6ff;
  border-color: #2563eb;
  color: #2563eb;
}

.delete-btn {
  color: #ef4444;
  border-color: #ef4444;
}

.delete-btn:hover {
  background: #fef2f2;
  border-color: #dc2626;
  color: #dc2626;
}

/* ============================================ */
/* RESPONSIVE */
/* ============================================ */
@media (max-width: 1024px) {
  td {
    padding: 0.75rem;
    font-size: 0.8125rem;
  }

  .source-path {
    max-width: 180px;
  }

  .actions {
    gap: 0.375rem;
  }

  .action-btn {
    padding: 0.25rem 0.375rem;
    min-width: 28px;
    height: 28px;
    font-size: 0.8125rem;
  }
}

@media (max-width: 768px) {
  td {
    padding: 0.625rem 0.5rem;
  }

  .source-path {
    max-width: 120px;
    font-size: 0.75rem;
  }

  .remote-badge {
    padding: 0.25rem 0.5rem;
    font-size: 0.75rem;
  }

  .status-badge {
    padding: 0.25rem 0.5rem;
    font-size: 0.7rem;
  }

  .timestamp {
    font-size: 0.75rem;
  }

  .actions {
    gap: 0.25rem;
    flex-wrap: wrap;
  }

  .action-btn {
    padding: 0.25rem;
    min-width: 26px;
    height: 26px;
    font-size: 0.75rem;
  }
}

@media (max-width: 480px) {
  .job-name {
    font-size: 0.8125rem;
  }

  .source-path {
    max-width: 100px;
  }

  .actions {
    flex-direction: column;
    gap: 0.25rem;
  }

  .action-btn {
    width: 100%;
    justify-content: center;
  }
}
</style>