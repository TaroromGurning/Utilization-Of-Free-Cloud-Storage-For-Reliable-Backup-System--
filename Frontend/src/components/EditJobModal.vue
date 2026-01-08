<template>
  <div class="config-view">
    <div v-if="isVisible" class="modal-overlay" @click.self="close">
      <div class="modal-content">
        <div class="modal-header">
          <h2>Edit Backup Configuration</h2>
          <button class="close-btn" @click="close">×</button>
        </div>  

        <form @submit.prevent="handleBackupSubmit" class="config-form">

          <div class="form-group">
            <label for="backup-jobName">Job Name *</label>
            <input type="text" id="backup-jobName" v-model="backupForm.job_name" required placeholder="e.g., Daily Database Backup" />
          </div>

          <div class="form-group">
            <label for="backup-source">Source Path (Lokal) *</label>
            <input type="text" id="backup-source" v-model="backupForm.source_path" required placeholder="/tmp/backup_file.zip" />
          </div>

          <div class="form-group">
            <label for="backup-mode">Backup Mode *</label>
            <select id="backup-mode" v-model="backupForm.rclone_mode" required>
              <option value="copy">Copy</option>
              <option value="sync">Sync</option>
            </select>
          </div>

          <div class="form-group">
            <label for="backup-remote">Drive Name *</label>
            <select id="backup-remote" v-model="backupForm.remote_name" required>
              <option value="" disabled>Select Drive</option>
              <option v-for="remote in remoteList" :key="remote.name" :value="remote.name">
                {{ remote.name }}
              </option>
            </select>
          </div>

          <div class="form-group">
            <label for="backup-dest">Destination Path (Cloud) *</label>
            <input type="text" id="backup-dest" v-model="backupForm.destination_path" required placeholder="/backups/database/" />
          </div>

          <div class="form-group" :class="{ 'disabled-group': backupForm.rclone_mode === 'sync' }">
            <label for="create-retention">Max Retention</label>
            <div class="input-with-hint">
              <input 
                type="number" 
                id="create-retention" 
                v-model.number="backupForm.max_retention" 
                min="1" 
                max="100"
                class="form-input"
                placeholder="10"
                :disabled="backupForm.rclone_mode === 'sync'"
              />
            </div>
            <small class="hint" v-if="backupForm.rclone_mode === 'copy'">
              <br><strong>Default: 10</strong>
            </small>
            <small class="hint warning" v-else>
              Fitur ini dinonaktifkan pada mode Sync.
            </small>
          </div>

          <div class="schedule-section">
            <div class="section-header">
              <h3>Schedule Configuration</h3>
              <label class="toggle-switch">
                <input type="checkbox" v-model="isScheduled" @change="handleScheduleToggle" />
                <span class="toggle-label">{{ isScheduled ? 'Scheduled Job' : 'Manual Job' }}</span>
              </label>
            </div>

            <transition name="slide-fade">
              <div v-if="isScheduled" class="schedule-options">
                <div class="schedule-type-selector">
                  <button type="button" v-for="type in scheduleTypes" :key="type.value"
                    @click="selectScheduleType(type.value)"
                    :class="['type-btn', { active: scheduleType === type.value }]">
                    <span class="type-label">{{ type.label }}</span>
                  </button>
                </div>

                <div v-if="scheduleType === 'hourly'" class="schedule-config">
                  <label>Every</label>
                  <div class="input-group">
                    <input type="number" v-model.number="scheduleConfig.hours" min="1" max="23" class="time-input" />
                    <span class="input-suffix">hour(s)</span>
                  </div>
                </div>

                <div v-if="scheduleType === 'daily'" class="schedule-config">
                  <label>Every day at</label>
                  <div class="input-group">
                    <input type="time" v-model="scheduleConfig.time" class="time-input" />
                  </div>
                </div>

                <div v-if="scheduleType === 'weekly'" class="schedule-config">
                  <label>Every</label>
                  <div class="weekdays-selector">
                    <button type="button" v-for="day in weekdays" :key="day.value" @click="toggleWeekday(day.value)"
                      :class="['weekday-btn', { active: scheduleConfig.weekdays.includes(day.value) }]">
                      {{ day.short }}
                    </button>
                  </div>
                  <label>at</label>
                  <div class="input-group">
                    <input type="time" v-model="scheduleConfig.time" class="time-input" />
                  </div>
                </div>

                <div v-if="scheduleType === 'monthly'" class="schedule-config">
                  <label>On day</label>
                  <div class="input-group">
                    <input type="number" v-model.number="scheduleConfig.dayOfMonth" min="1" max="31" class="time-input" />
                    <span class="input-suffix">of every month</span>
                  </div>
                  <label>at</label>
                  <div class="input-group">
                    <input type="time" v-model="scheduleConfig.time" class="time-input" />
                  </div>
                </div>

                <div v-if="scheduleType === 'custom'" class="schedule-config">
                  <label>Custom Cron Expression</label>
                  <input type="text" v-model="scheduleConfig.customCron" placeholder="*/5 * * * *" class="cron-input" />
                  <small class="hint">Format: minute hour day month weekday <a href="https://crontab.guru" target="_blank">Need help?</a></small>
                </div>

                <div class="cron-preview">
                  <span class="preview-label">Cron Expression:</span>
                  <code class="preview-code">{{ generatedCron || '-' }}</code>
                  <span class="preview-description">{{ cronDescription }}</span>
                </div>
              </div>
            </transition>
          </div>

          <div class="form-group">
            <label for="backup-pre">Pre-Script (Executed BEFORE Rclone)</label>
            <textarea id="backup-pre" v-model="backupForm.pre_script" rows="3" placeholder="#!/bin/bash
# Example: Database dump
mysqldump -u user -p password database > /tmp/backup.sql
gzip /tmp/backup.sql"></textarea>
          </div>

          <div class="form-group">
            <label for="backup-post">Post-Script (Executed AFTER successful upload)</label>
            <textarea id="backup-post" v-model="backupForm.post_script" rows="3" placeholder="#!/bin/bash
# Example: Cleanup
rm /tmp/backup.sql.gz"></textarea>
          </div>

          <div class="form-actions">
            <button type="button" @click="close" class="btn-secondary">Cancel</button>
            <button type="submit" :disabled="isLoading" class="btn-submit">
              <span v-if="isLoading">Updating...</span>
              <span v-else>Update Job</span>
            </button>
          </div>
        </form>

        <transition name="fade">
          <div v-if="message" class="message success">{{ message }}</div>
        </transition>
        <transition name="fade">
          <div v-if="errorMessage" class="message error">{{ errorMessage }}</div>
        </transition>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import jobService from '@/services/jobService'
import driveService from '@/services/driveService'

const props = defineProps({ 
  isVisible: Boolean,
  jobData: Object
})
const emit = defineEmits(['close','success'])

// ✅ Computed untuk mendapatkan job ID dari nested structure
const jobId = computed(() => {
  if (!props.jobData) return null
  // Backend returns: { success: true, data: { id, ... } }
  const actualData = props.jobData.data || props.jobData
  return actualData.id
})

const jobIdToSubmit = computed(() => {
    // Cek format data yang dikirim Parent: { success: true, data: { id: X } }
    const data = props.jobData && props.jobData.data ? props.jobData.data : props.jobData;
    
    // Prioritaskan ID dari prop (jika ada), jika tidak, ambil dari data object
    const id = props.jobId || (data ? data.id : null);
    
    // Kembalikan sebagai Number, memastikan ID valid
    return Number(id) || null;
});

const isLoading = ref(false)
const message = ref(null)
const errorMessage = ref(null)

const isScheduled = ref(false)
const scheduleType = ref('daily')
const scheduleConfig = ref({ 
  hours: 1, 
  time: '00:00', 
  weekdays: [], 
  dayOfMonth: 1, 
  customCron: '' 
})

const scheduleTypes = [
  { value: 'hourly', label: 'Hourly' },
  { value: 'daily', label: 'Daily' },
  { value: 'weekly', label: 'Weekly' },
  { value: 'monthly', label: 'Monthly' },
  { value: 'custom', label: 'Custom' }
]

const weekdays = [
  { value: 0, short: 'Sun' },
  { value: 1, short: 'Mon' },
  { value: 2, short: 'Tue' },
  { value: 3, short: 'Wed' },
  { value: 4, short: 'Thu' },
  { value: 5, short: 'Fri' },
  { value: 6, short: 'Sat' }
]

const backupForm = ref({
  job_name: '',
  rclone_mode: 'copy',
  source_path: '',
  remote_name: '',
  destination_path: '',
  schedule_cron: '',
  pre_script: '',
  post_script: '',
  max_retention: 10
})

const remoteList = ref([])

async function fetchRemoteList() {
  try {
    const res = await driveService.listRemotes()
    remoteList.value = Array.isArray(res) ? res : res.remotes || []
  } catch (err) {
    console.error('Failed to fetch remotes:', err)
  }
}

// ✅ Function untuk parse cron expression yang lebih robust
function parseCronExpression(cron) {
  if (!cron || cron.trim() === '') {
    isScheduled.value = false
    scheduleType.value = 'daily'
    scheduleConfig.value = { 
      hours: 1, 
      time: '00:00', 
      weekdays: [], 
      dayOfMonth: 1, 
      customCron: '' 
    }
    return
  }
  
  isScheduled.value = true
  const parts = cron.trim().split(/\s+/)
  
  if (parts.length !== 5) {
    scheduleType.value = 'custom'
    scheduleConfig.value.customCron = cron
    return
  }
  
  const [minute, hour, dayOfMonth, month, dayOfWeek] = parts
  
  // Hourly: 0 */N * * *
  if (hour.startsWith('*/') && dayOfMonth === '*' && month === '*' && dayOfWeek === '*') {
    scheduleType.value = 'hourly'
    scheduleConfig.value.hours = parseInt(hour.replace('*/', '')) || 1
    return
  }
  
  // Daily: M H * * *
  if (dayOfMonth === '*' && month === '*' && dayOfWeek === '*' && !hour.includes('*') && !minute.includes('*')) {
    scheduleType.value = 'daily'
    const h = hour.padStart(2, '0')
    const m = minute.padStart(2, '0')
    scheduleConfig.value.time = `${h}:${m}`
    return
  }
  
  // Weekly: M H * * D,D,D
  if (dayOfMonth === '*' && month === '*' && dayOfWeek !== '*' && !hour.includes('*') && !minute.includes('*')) {
    scheduleType.value = 'weekly'
    const h = hour.padStart(2, '0')
    const m = minute.padStart(2, '0')
    scheduleConfig.value.time = `${h}:${m}`
    scheduleConfig.value.weekdays = dayOfWeek.split(',').map(d => parseInt(d.trim())).filter(d => !isNaN(d))
    return
  }
  
  // Monthly: M H D * *
  if (month === '*' && dayOfWeek === '*' && dayOfMonth !== '*' && !hour.includes('*') && !minute.includes('*')) {
    scheduleType.value = 'monthly'
    const h = hour.padStart(2, '0')
    const m = minute.padStart(2, '0')
    scheduleConfig.value.time = `${h}:${m}`
    scheduleConfig.value.dayOfMonth = parseInt(dayOfMonth) || 1
    return
  }
  
  // Fallback to custom
  scheduleType.value = 'custom'
  scheduleConfig.value.customCron = cron
}

// ✅ Function untuk populate form dengan data yang ada
function populateForm(jobData) {
  if (!jobData) {
    console.warn('No job data provided to populateForm')
    return
  }
  
  console.log('=== POPULATE FORM ===')
  console.log('Raw jobData received:', jobData)
  
  // ✅ Handle nested data structure dari backend
  // Backend returns: { success: true, data: { id, job_name, ... } }
  const actualData = jobData.data || jobData
  
  console.log('Actual data to populate:', actualData)
  console.log('Available fields:', Object.keys(actualData))
  
  // ✅ Populate basic fields sesuai dengan response backend
  backupForm.value = {
    job_name: actualData.job_name || '',
    rclone_mode: actualData.rclone_mode || 'copy',
    source_path: actualData.source_path || '',
    remote_name: actualData.remote_name || '',
    destination_path: actualData.destination_path || '',
    schedule_cron: actualData.schedule_cron || '',
    pre_script: actualData.pre_script || '',
    post_script: actualData.post_script || '',
    max_retention: actualData.max_retention || 10
  }
  
  console.log('✅ Form values populated:', {
    job_name: backupForm.value.job_name,
    source_path: backupForm.value.source_path,
    remote_name: backupForm.value.remote_name,
    destination_path: backupForm.value.destination_path,
    schedule_cron: backupForm.value.schedule_cron,
    max_retention: backupForm.value.max_retention,
    rclone_mode: backupForm.value.rclone_mode
  })
  
  // ✅ Parse dan set schedule configuration
  parseCronExpression(actualData.schedule_cron || '')
  
  console.log('✅ Schedule parsed:', {
    isScheduled: isScheduled.value,
    scheduleType: scheduleType.value,
    scheduleConfig: scheduleConfig.value
  })
}

// ✅ Computed untuk generate cron expression
const generatedCron = computed(() => {
  if (!isScheduled.value) return ''
  
  const cfg = scheduleConfig.value
  
  switch(scheduleType.value) {
    case 'hourly': 
      return `0 */${cfg.hours || 1} * * *`
    
    case 'daily': {
      const [h, m] = (cfg.time || '00:00').split(':')
      return `${m} ${h} * * *`
    }
    
    case 'weekly': {
      const [h, m] = (cfg.time || '00:00').split(':')
      const days = cfg.weekdays.sort((a, b) => a - b).join(',')
      return days ? `${m} ${h} * * ${days}` : ''
    }
    
    case 'monthly': {
      const [h, m] = (cfg.time || '00:00').split(':')
      return `${m} ${h} ${cfg.dayOfMonth || 1} * *`
    }
    
    case 'custom': 
      return cfg.customCron || ''
    
    default: 
      return ''
  }
})

const cronDescription = computed(() => {
  if (!generatedCron.value) return 'No schedule configured'
  
  const cfg = scheduleConfig.value
  
  switch(scheduleType.value) {
    case 'hourly': 
      return `Every ${cfg.hours} hour${cfg.hours > 1 ? 's' : ''}`
    
    case 'daily': 
      return `Every day at ${cfg.time}`
    
    case 'weekly': {
      const dayNames = cfg.weekdays
        .map(d => weekdays.find(w => w.value === d)?.short)
        .filter(Boolean)
        .join(', ')
      return `Every ${dayNames || 'no days'} at ${cfg.time}`
    }
    
    case 'monthly': 
      return `On day ${cfg.dayOfMonth} of every month at ${cfg.time}`
    
    case 'custom': 
      return `Custom cron: ${cfg.customCron}`
    
    default: 
      return 'No schedule configured'
  }
})

// ✅ Watch untuk sync cron expression
watch(generatedCron, (newCron) => {
  if (isScheduled.value) {
    backupForm.value.schedule_cron = newCron
  }
})

// ✅ Watch untuk mode sync
watch(() => backupForm.value.rclone_mode, (newMode) => {
  if (newMode === 'sync') {
    backupForm.value.max_retention = null
  }
})

function handleScheduleToggle() {
  if (!isScheduled.value) {
    backupForm.value.schedule_cron = ''
  } else {
    backupForm.value.schedule_cron = generatedCron.value
  }
}

function selectScheduleType(type) { 
  scheduleType.value = type 
}

function toggleWeekday(day) { 
  const idx = scheduleConfig.value.weekdays.indexOf(day)
  if (idx > -1) {
    scheduleConfig.value.weekdays.splice(idx, 1)
  } else {
    scheduleConfig.value.weekdays.push(day)
  }
}

async function handleBackupSubmit() {
  isLoading.value = true
  message.value = null
  errorMessage.value = null
  
  try {
    if (!isScheduled.value) {
      backupForm.value.schedule_cron = ''
    } else {
      backupForm.value.schedule_cron = generatedCron.value
    }

    console.log('Submitting update:', backupForm.value)
    
    const res = await jobService.updateJob(jobId.value, backupForm.value)
    message.value = res.message || 'Job updated successfully!'
    emit('success')
    setTimeout(() => emit('close'), 1500)
  } catch (err) {
    console.error('Update error:', err)
    errorMessage.value = err.response?.data?.error || 'Failed to update backup job.'
  } finally { 
    isLoading.value = false 
  }
}

function close() { 
  emit('close') 
}

// ✅ Watch untuk ketika modal dibuka
watch(() => props.isVisible, (newVal) => {
  if (newVal && props.jobData) {
    console.log('=== MODAL OPENED ===')
    console.log('Modal opened with job data:', props.jobData)
    console.log('Job data type:', typeof props.jobData)
    console.log('Job data keys:', Object.keys(props.jobData))
    console.log('Raw job data:', JSON.stringify(props.jobData, null, 2))
    fetchRemoteList()
    populateForm(props.jobData)
  }
})

// ✅ Watch untuk ketika jobData berubah
watch(() => props.jobData, (newData) => {
  if (props.isVisible && newData) {
    console.log('=== JOB DATA CHANGED ===')
    console.log('Job data changed:', newData)
    console.log('Job data keys:', Object.keys(newData))
    console.log('Raw job data:', JSON.stringify(newData, null, 2))
    populateForm(newData)
  }
}, { deep: true })

// ✅ Initial load
onMounted(() => {
  fetchRemoteList()
  if (props.isVisible && props.jobData) {
    populateForm(props.jobData)
  }
})
</script>

<style scoped>
.config-view { max-width: 900px; margin: 0 auto; }
.modal-overlay { position: fixed; inset: 0; background: rgba(0, 0, 0, 0.5); backdrop-filter: blur(4px); display: flex; align-items: center; justify-content: center; z-index: 1000; padding: 20px; }
.modal-content { background: #fff; border-radius: 8px; border: 1px solid #e5e5e5; width: 90%; max-width: 800px; max-height: 90vh; overflow-y: auto; position: relative; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 1.25rem 1.5rem; background: #fafafa; border-bottom: 1px solid #e5e5e5; }
.modal-header h2 { margin: 0; font-size: 1.125rem; font-weight: 600; color: #1a1a1a; }
.close-btn { background: transparent; border: 1px solid #e5e5e5; color: #666; width: 32px; height: 32px; border-radius: 4px; cursor: pointer; font-size: 1.5rem; line-height: 1; display: flex; align-items: center; justify-content: center; transition: all 0.2s; }
.close-btn:hover { background: #f5f5f5; border-color: #1a1a1a; color: #1a1a1a; }

.config-form { padding: 1.5rem; }
.form-description { color: #666; margin: 0 0 1.5rem 0; font-size: 0.9375rem; }
.form-group { margin-bottom: 1.25rem; }
.form-group label { display: block; margin-bottom: 0.5rem; font-weight: 500; color: #1a1a1a; font-size: 0.875rem; }
.form-group input, .form-group select, .form-group textarea { width: 100%; padding: 0.625rem 0.875rem; border: 1px solid #e5e5e5; border-radius: 6px; font-size: 0.9375rem; transition: all 0.2s; box-sizing: border-box; }
.form-group input:focus, .form-group select:focus, .form-group textarea:focus { outline: none; border-color: #1a1a1a; }

.disabled-group {
  opacity: 0.6;
  pointer-events: none;
}
.disabled-group input {
  background-color: #f5f5f5;
  color: #999;
}
.hint.warning {
  color: #d97706;
  font-weight: 500;
}

.hint { display: block; margin-top: 0.375rem; font-size: 0.8125rem; color: #666; }
.schedule-section { background: #fafafa; padding: 1rem; border-radius: 1px; margin-bottom: 1rem; border: 1px solid #000000; }
.section-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.25rem; }
.section-header h3 { margin: 0; font-size: 0.9375rem; font-weight: 600; color: #1a1a1a; }
.toggle-switch { display: flex; align-items: center; gap: 0.75rem; cursor: pointer; user-select: none; }
.toggle-switch input[type="checkbox"] { position: relative; width: 44px; height: 24px; appearance: none; background: #d4d4d4; border-radius: 12px; cursor: pointer; transition: all 0.2s; }
.toggle-switch input[type="checkbox"]:checked { background: #1a1a1a; }
.toggle-switch input[type="checkbox"]::before { content: ''; position: absolute; width: 20px; height: 20px; border-radius: 50%; background: white; top: 2px; left: 2px; transition: all 0.2s; }
.toggle-switch input[type="checkbox"]:checked::before { left: 22px; }
.toggle-label { font-weight: 500; color: #1a1a1a; font-size: 0.875rem; }
.schedule-options { margin-top: 1.25rem; }
.schedule-type-selector { display: grid; grid-template-columns: repeat(auto-fit, minmax(80px, 1fr)); gap: 0.5rem; margin-bottom: 1.25rem; }
.type-btn { padding: 0.625rem; background: white; border: 1px solid #e5e5e5; border-radius: 6px; cursor: pointer; transition: all 0.2s; font-weight: 500; font-size: 0.875rem; color: #666; }
.type-btn:hover { border-color: #1a1a1a; color: #1a1a1a; }
.type-btn.active { background: #1a1a1a; border-color: #1a1a1a; color: white; }
.schedule-config { background: white; padding: 1.25rem; border-radius: 6px; margin-bottom: 1rem; border: 1px solid #e5e5e5; }
.schedule-config label { display: block; margin-bottom: 0.5rem; font-weight: 500; color: #1a1a1a; font-size: 0.875rem; }
.input-group { display: flex; align-items: center; gap: 0.75rem; margin-bottom: 1rem; }
.time-input { flex: 0 0 auto; padding: 0.625rem 0.875rem; border: 1px solid #e5e5e5; border-radius: 6px; font-size: 0.9375rem; font-weight: 500; min-width: 100px; }
.input-suffix { color: #666; font-weight: 500; font-size: 0.875rem; }
.weekdays-selector { display: grid; grid-template-columns: repeat(7, 1fr); gap: 0.5rem; margin-bottom: 1rem; }
.weekday-btn { padding: 0.625rem 0.375rem; background: white; border: 1px solid #e5e5e5; border-radius: 6px; cursor: pointer; font-weight: 500; font-size: 0.875rem; color: #666; transition: all 0.2s; }
.weekday-btn:hover { border-color: #1a1a1a; color: #1a1a1a; }
.weekday-btn.active { background: #1a1a1a; border-color: #1a1a1a; color: white; }
.cron-input { width: 100%; padding: 0.625rem 0.875rem; border: 1px solid #e5e5e5; border-radius: 6px; font-family: 'Consolas', monospace; font-size: 0.875rem; }
.cron-preview { background: #1a1a1a; padding: 1rem; border-radius: 6px; display: flex; align-items: center; gap: 0.75rem; flex-wrap: wrap; }
.preview-label { color: #999; font-weight: 500; font-size: 0.8125rem; }
.preview-code { background: #2a2a2a; color: #22c55e; padding: 0.375rem 0.75rem; border-radius: 4px; font-family: 'Consolas', monospace; font-size: 0.875rem; font-weight: 500; }
.preview-description { color: #d4d4d4; font-size: 0.875rem; }
.form-actions { display: flex; justify-content: flex-end; gap: 0.75rem; margin-top: 1.5rem; padding-top: 1.25rem; border-top: 1px solid #e5e5e5; }
.btn-submit { background: #1a1a1a; color: white; padding: 0.625rem 1.5rem; border: none; border-radius: 6px; cursor: pointer; font-weight: 500; font-size: 0.9375rem; transition: all 0.2s; }
.btn-submit:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-secondary { background: white; color: #666; border: 1px solid #e5e5e5; padding: 0.625rem 1.125rem; border-radius: 6px; cursor: pointer; font-weight: 500; font-size: 0.9375rem; }
.message.success { background: #d1fae5; color: #065f46; border-left: 3px solid #22c55e; padding: 0.875rem 1rem; border-radius: 6px; margin-top: 1rem; }
.message.error { background: #fee2e2; color: #991b1b; border-left: 3px solid #ef4444; padding: 0.875rem 1rem; border-radius: 6px; margin-top: 1rem; }

.slide-fade-enter-active { transition: all 0.3s ease-out; }
.slide-fade-leave-active { transition: all 0.2s cubic-bezier(1.0, 0.5, 0.8, 1.0); }
.slide-fade-enter-from { transform: translateY(-10px); opacity: 0; }
.slide-fade-leave-to { transform: translateY(-10px); opacity: 0; }
.fade-enter-active, .fade-leave-active { transition: opacity 0.3s; }
.fade-enter-from, .fade-leave-to { opacity: 0; }

@media (max-width: 768px) { 
  .modal-content { width: 95%; } 
  .config-form { padding: 1rem; } 
  .schedule-type-selector { grid-template-columns: repeat(auto-fit, minmax(70px, 1fr)); } 
  .form-actions { flex-direction: column; } 
  .btn-submit, .btn-secondary { width: 100%; } 
}
</style>