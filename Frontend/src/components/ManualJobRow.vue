<template>
  <tr>
    <td class="truncate" :title="job.job_name"><strong>{{ job.job_name }}</strong></td>
    <td class="truncate" :title="job.source_path">{{ job.source_path }}</td>
    <td class="truncate" :title="job.gdrive_target">{{ job.gdrive_target }}</td>
    <td>{{ job.last_run || 'N/A' }}</td>

    <td>
      <span class="status" :class="getStatusClass(job.status)">
        {{ job.status }}
      </span>
    </td>

    <td>{{ job.next_run || 'N/A' }}</td>
    
    <td class="job-actions-col"> 
      <div class="actions">
        <!-- Run/Trigger Button -->
        <button 
          @click="$emit('trigger', job.id)" 
          class="action-btn trigger" 
          title="Run Job Now"
        >
          ▶
        </button>

        <!-- Edit Button - HIDDEN untuk RESTORE -->
        <button 
          v-if="job.operation_mode != 'RESTORE'"
          @click="$emit('edit', job.id)" 
          class="action-btn edit" 
          title="Edit Job"
        >
          ✎
        </button>

        <!-- View Script Button - HIDDEN untuk RESTORE -->
        <button 
          v-if="job.operation_mode != 'RESTORE'"
          @click="$emit('view-script', job.id)" 
          class="action-btn view" 
          title="View Script"
        >
          (i)
        </button>

        <!-- Delete Button -->
        <button 
          @click="handleDelete" 
          class="action-btn delete" 
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
  const jobType = props.job.operation_mode === 'RESTORE' ? 'restore' : 'job manual';
  if (!confirm(`Apakah Anda yakin ingin menghapus ${jobType} "${props.job.job_name}"?\nAksi ini tidak dapat dibatalkan.`)) {
    return;
  }
  emit('delete', props.job.id, props.job.job_name);
}

function getStatusClass(status) {
  if (!status) return 'pending';
  const statusLower = status.toLowerCase();
  
  if (statusLower === 'completed' || statusLower === 'success') return 'completed';
  if (statusLower === 'running') return 'running';
  if (statusLower === 'pending') return 'pending';
  if (statusLower.includes('fail')) return 'failed';
  
  return statusLower;
}
</script>

<style scoped>
/* --- STATUS BADGE --- */
.status {
  display: inline-block;
  padding: 3px 8px;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  color: white;
}

.status.pending {
  background-color: #7f8c8d;
}

.status.running {
  background-color: #3498db;
}

.status.completed,
.status.success {
  background-color: #27ae60;
}

.status.failed,
.status.fail_source_check, 
.status.fail_pre_script, 
.status.fail_rclone,
.status.fail_post_script {
  background-color: #e74c3c;
}

/* --- TABLE CELL --- */
td {
  padding: 1rem;
  vertical-align: middle;
  border-bottom: 1px solid #f0f0f0;
  font-size: 0.9375rem;
  color: #333;
}

strong {
  color: #1a1a1a;
}

/* --- TRUNCATE DENGAN ELLIPSIS --- */
.truncate {
  max-width: 220px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  cursor: help;
}

/* --- TOOLTIP SAAT HOVER --- */
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

/* --- ACTIONS COLUMN --- */
.job-actions-col {
  width: auto;
  text-align: center;
}

.actions {
  display: flex;
  justify-content: center;
  gap: 8px;
}

/* Action Buttons */
.action-btn {
  border: none;
  background: none;
  cursor: pointer;
  padding: 6px;
  font-size: 1.15rem;
  transition: transform 0.15s ease, opacity 0.2s ease;
}

.action-btn:hover {
  opacity: 0.85;
  transform: scale(1.15);
}

.action-btn:active {
  transform: scale(0.95);
}

/* Button Colors */
.action-btn.trigger {
  color: #27ae60;
}

.action-btn.edit {
  color: #f39c12;
}

.action-btn.view {
  color: #3498db;
}

.action-btn.delete {
  color: #e74c3c;
}

@media (max-width: 768px) {
  .actions {
    gap: 4px;
  }
  
  .action-btn {
    font-size: 1rem;
    padding: 4px;
  }

  .truncate {
    max-width: 100px;
  }

  .truncate:hover::after {
    max-width: 200px;
  }
}
</style>