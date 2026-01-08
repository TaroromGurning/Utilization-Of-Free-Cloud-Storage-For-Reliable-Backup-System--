<template>
  <transition name="fade">
    <div v-if="props.isVisible" class="modal-overlay" @click.self="close">
      <div class="modal-content">
        <!-- Header -->
        <div class="modal-header">
          <h2>Restore Configuration</h2>
          <button type="button" class="close-btn" @click="close">×</button>
        </div>

        <!-- Content -->
        <div class="restore-view">
          <!-- Step Indicator -->
          <div class="step-indicator">
            <div :class="['step', { active: currentStep >= 1, completed: currentStep > 1 }]">
              <div class="step-number">1</div>
              <div class="step-label">Remote</div>
            </div>
            <div class="step-line" :class="{ completed: currentStep > 1 }"></div>

            <div :class="['step', { active: currentStep >= 2, completed: currentStep > 2 }]">
              <div class="step-number">2</div>
              <div class="step-label">Select</div>
            </div>
            <div class="step-line" :class="{ completed: currentStep > 2 }"></div>

            <div :class="['step', { active: currentStep >= 3 }]">
              <div class="step-number">3</div>
              <div class="step-label">Destination</div>
            </div>
          </div>

          <!-- Step 1: Select Remote Drive -->
          <div v-show="currentStep === 1" class="step-content">
            <h3>Select Remote Storage</h3>
            <p class="form-description">Choose the cloud storage where your backup is located</p>

            <div v-if="loadingRemotes" class="state-loading">
              <div class="spinner"></div>
              <p>Loading remotes...</p>
            </div>

            <div v-else-if="remotesList.length === 0" class="state-empty">
              <p>No remotes found</p>
            </div>

            <div v-else class="form-group">
              <label for="restore-remote">Remote Name *</label>
              <select id="restore-remote" v-model="selectedRemote" required>
                <option value="" disabled>Choose a remote...</option>
                <option v-for="remote in remotesList" :key="remote.name" :value="remote.name">
                  {{ remote.name }}
                </option>
              </select>
            </div>

            <div class="step-actions">
              <button type="button" @click="close" class="btn-secondary">Cancel</button>
              <button 
                type="button"
                @click="goToStep(2)" 
                class="btn-primary"
                :disabled="!selectedRemote"
              >
                Next
              </button>
            </div>
          </div>

          <!-- Step 2: Select File/Folder -->
          <div v-show="currentStep === 2" class="step-content">
            <h3>Select File or Folder</h3>
            <p class="form-description">Choose the file/folder from {{ selectedRemote }} to restore</p>

            <div class="file-browser-wrapper">
              <FileBrowser
                v-if="selectedRemote"
                :remote-name="selectedRemote"
                :initial-path="'/'"
                @select-file="handleSelectFile"
              />
            </div>

            <div v-if="selectedFile" class="selected-display">
              <div class="selected-item">
                <span class="item-type">{{ selectedFile.is_dir ? 'Folder' : 'File' }}</span>
                <span class="item-name">{{ selectedFile.name }}</span>
              </div>
              <div class="selected-meta">
                <small>{{ selectedFile.path }}</small>
              </div>
            </div>

            <div class="step-actions">
              <button type="button" @click="goToStep(1)" class="btn-secondary">Back</button>
              <button 
                type="button"
                @click="goToStep(3)" 
                class="btn-primary"
                :disabled="!selectedFile"
              >
                Next
              </button>
            </div>
          </div>

          <!-- Step 3: Set Destination Path -->
          <div v-show="currentStep === 3" class="step-content">
            <h3>Set Restore Destination</h3>
            <p class="form-description">Specify where to restore the file on your server</p>

            <!-- Review -->
            <div class="review-section">
              <div class="review-item">
                <span class="review-label">Remote:</span>
                <span class="review-value">{{ selectedRemote }}</span>
              </div>
              <div class="review-item">
                <span class="review-label">Source:</span>
                <span class="review-value">{{ selectedFile?.name }}</span>
              </div>
              <div class="review-item">
                <span class="review-label">Path:</span>
                <span class="review-value">{{ selectedRemote }}:{{ selectedFile?.path }}</span>
              </div>
            </div>

            <form @submit.prevent="handleRestoreSubmit" class="config-form">
              <div class="form-group">
                <label for="restore-dest">Destination Path *</label>
                <input 
                  type="text" 
                  id="restore-dest" 
                  v-model="destinationPath" 
                  required 
                  placeholder="/home/user/restore/"
                />
                <small class="hint">Example: /home/user/restore or /opt/backups</small>
              </div>

              <transition name="fade">
                <div v-if="errorMessage" class="alert alert-error">
                  {{ errorMessage }}
                </div>
              </transition>

              <transition name="fade">
                <div v-if="successMessage" class="alert alert-success">
                  {{ successMessage }}
                </div>
              </transition>

              <div class="step-actions">
                <button type="button" @click="goToStep(2)" class="btn-secondary" :disabled="isLoading">
                  Back
                </button>
                <button 
                  type="submit" 
                  :disabled="isLoading || !destinationPath" 
                  class="btn-submit"
                >
                  <span v-if="isLoading">Starting Restore...</span>
                  <span v-else>Start Restore</span>
                </button>
              </div>
            </form>
          </div>
        </div>
      </div>
    </div>
  </transition>
</template>

<script setup>
import { ref, watch } from 'vue'
import FileBrowser from '@/components/GDriveBrowser.vue'
import jobService from '@/services/jobService'
import driveService from '@/services/driveService'

const props = defineProps({
  isVisible: { type: Boolean, default: false }
})
const emit = defineEmits(['close', 'success'])

// Step management
const currentStep = ref(1)

// Step 1: Remote
const loadingRemotes = ref(false)
const remotesList = ref([])
const selectedRemote = ref('')

// Step 2: File
const selectedFile = ref(null)

// Step 3: Destination
const destinationPath = ref('')

// API state
const isLoading = ref(false)
const errorMessage = ref(null)
const successMessage = ref(null)

// ============================================
// Methods
// ============================================

// Load remotes
async function loadRemotes() {
  loadingRemotes.value = true
  try {
    const response = await driveService.listRemotes()
    remotesList.value = Array.isArray(response) ? response : (response.remotes || [])
  } catch (error) {
    console.error('Failed to load remotes:', error)
    remotesList.value = []
  } finally {
    loadingRemotes.value = false
  }
}

// Handle file selection from browser
function handleSelectFile(file) {
  selectedFile.value = file
}

// Navigate to step
function goToStep(step) {
  if (Math.abs(step - currentStep.value) === 1 || step < currentStep.value) {
    currentStep.value = step
  }
}

// Submit restore
async function handleRestoreSubmit() {
  if (!selectedRemote.value || !selectedFile.value || !destinationPath.value) {
    errorMessage.value = 'All fields must be filled'
    return
  }

  isLoading.value = true
  errorMessage.value = null
  successMessage.value = null

  try {
    const response = await jobService.createRestoreJob({
      remote_name: selectedRemote.value,
      source_path: selectedFile.value.path,
      destination_path: destinationPath.value,
      job_name: `Restore ${selectedFile.value.name}`
    })

    successMessage.value = response.message || 'Restore job started successfully!'
    emit('success')

    setTimeout(() => close(), 1500)
  } catch (error) {
    console.error('Restore error:', error)
    errorMessage.value = error.response?.data?.error || 'Failed to start restore'
  } finally {
    isLoading.value = false
  }
}

// Close modal
function close() {
  emit('close')
  resetForm()
}

// Reset form
function resetForm() {
  currentStep.value = 1
  selectedRemote.value = ''
  selectedFile.value = null
  destinationPath.value = ''
  errorMessage.value = null
  successMessage.value = null
}

// Load remotes when modal opens
watch(() => props.isVisible, (newVal) => {
  if (newVal) {
    loadRemotes()
  }
})
</script>

<style scoped>
/* Modal Overlay */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 999;
  padding: 20px;
}

.modal-content {
  background: #fff;
  border-radius: 8px;
  border: 1px solid #e5e5e5;
  width: 90%;
  max-width: 700px;
  max-height: 90vh;
  overflow-y: auto;
  position: relative;
}

/* Header */
.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.25rem;
  background: #f9f9f9;
  border-bottom: 1px solid #e5e5e5;
  position: sticky;
  top: 0;
  z-index: 10;
}

.modal-header h2 {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
  color: #1a1a1a;
}

.close-btn {
  background: transparent;
  border: 1px solid #e5e5e5;
  color: #666;
  width: 32px;
  height: 32px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 1.5rem;
  line-height: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s;
}

.close-btn:hover {
  background: #f0f0f0;
  border-color: #ccc;
  color: #1a1a1a;
}

/* Restore View */
.restore-view {
  padding: 1.25rem;
}

/* Step Content */
.step-content {
  animation: fadeIn 0.25s ease-in;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.step-content h3 {
  margin: 0 0 0.25rem 0;
  font-size: 1.05rem;
  font-weight: 600;
  color: #1a1a1a;
}

.form-description {
  color: #666;
  margin: 0 0 1.25rem 0;
  font-size: 0.8rem;
}

/* Step Indicator */
.step-indicator {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1.75rem;
  gap: 0.25rem;
}

.step {
  display: flex;
  flex-direction: column;
  align-items: center;
  flex: 1;
}

.step-number {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: #f0f0f0;
  border: 2px solid #e5e5e5;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  color: #666;
  font-size: 0.85rem;
  transition: all 0.2s;
}

.step.active .step-number {
  background: #0066cc;
  border-color: #0066cc;
  color: white;
}

.step.completed .step-number {
  background: #10b981;
  border-color: #10b981;
  color: white;
}

.step-label {
  font-size: 0.7rem;
  font-weight: 500;
  color: #666;
  margin-top: 0.3rem;
  text-align: center;
}

.step-line {
  flex: 1;
  height: 2px;
  background: #e5e5e5;
  transition: all 0.2s;
}

.step-line.completed {
  background: #10b981;
}

/* File Browser */
.file-browser-wrapper {
  margin-bottom: 1rem;
  border: 1px solid #e5e5e5;
  border-radius: 6px;
  overflow: hidden;
}

/* Selected Display */
.selected-display {
  background: #f0f8ff;
  border: 1px solid #d4e8f5;
  border-radius: 4px;
  padding: 0.75rem;
  margin-bottom: 1.25rem;
}

.selected-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.3rem;
}

.item-type {
  display: inline-block;
  padding: 0.2rem 0.4rem;
  background: #0066cc;
  color: white;
  border-radius: 3px;
  font-size: 0.65rem;
  font-weight: 600;
  min-width: 40px;
  text-align: center;
}

.item-name {
  font-weight: 500;
  color: #1a1a1a;
  font-size: 0.85rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  flex: 1;
}

.selected-meta {
  font-size: 0.75rem;
  color: #666;
}

/* Review Section */
.review-section {
  background: #f9f9f9;
  border: 1px solid #e5e5e5;
  border-radius: 4px;
  padding: 0.75rem;
  margin-bottom: 1.25rem;
  font-size: 0.8rem;
}

.review-item {
  display: flex;
  justify-content: space-between;
  padding: 0.4rem 0;
  border-bottom: 1px solid #e5e5e5;
}

.review-item:last-child {
  border-bottom: none;
}

.review-label {
  color: #666;
  font-weight: 500;
  min-width: 70px;
}

.review-value {
  color: #0066cc;
  font-weight: 500;
  text-align: right;
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* State Messages */
.state-loading,
.state-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 1.5rem;
  gap: 0.75rem;
}

.state-empty {
  color: #999;
  font-size: 0.85rem;
}

.spinner {
  width: 32px;
  height: 32px;
  border: 2px solid #e5e5e5;
  border-top-color: #0066cc;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Form */
.config-form {
  width: 100%;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.4rem;
  font-weight: 500;
  color: #1a1a1a;
  font-size: 0.8rem;
}

.form-group input[type="text"],
.form-group select {
  width: 100%;
  padding: 0.5rem 0.75rem;
  border: 1px solid #e5e5e5;
  border-radius: 4px;
  font-size: 0.8rem;
  transition: all 0.15s;
  box-sizing: border-box;
}

.form-group input:focus,
.form-group select:focus {
  outline: none;
  border-color: #0066cc;
  box-shadow: 0 0 0 2px rgba(0, 102, 204, 0.1);
}

.hint {
  display: block;
  margin-top: 0.3rem;
  font-size: 0.7rem;
  color: #999;
}

/* Alerts */
.alert {
  padding: 0.75rem;
  border-radius: 4px;
  margin-bottom: 1rem;
  font-size: 0.8rem;
  border-left: 3px solid;
}

.alert-success {
  background: #d1fae5;
  color: #065f46;
  border-left-color: #10b981;
}

.alert-error {
  background: #fee2e2;
  color: #991b1b;
  border-left-color: #ef4444;
}

/* Step Actions */
.step-actions {
  display: flex;
  justify-content: space-between;
  gap: 0.75rem;
  margin-top: 1.25rem;
  padding-top: 1rem;
  border-top: 1px solid #e5e5e5;
}

.btn-secondary,
.btn-primary,
.btn-submit {
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-weight: 500;
  font-size: 0.8rem;
  transition: all 0.15s;
  min-width: fit-content;
}

.btn-secondary {
  background: #f5f5f5;
  color: #666;
  border: 1px solid #e5e5e5;
}

.btn-secondary:hover:not(:disabled) {
  background: #efefef;
  border-color: #ccc;
}

.btn-secondary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-primary {
  background: #f5f5f5;
  color: #1a1a1a;
  border: 1px solid #e5e5e5;
  flex: 1;
}

.btn-primary:hover:not(:disabled) {
  background: #e5e5e5;
  border-color: #ccc;
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-submit {
  background: #0066cc;
  color: white;
  flex: 1;
}

.btn-submit:hover:not(:disabled) {
  background: #0052a3;
}

.btn-submit:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Transitions */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* Responsive */
@media (max-width: 640px) {
  .modal-content {
    max-height: 85vh;
  }

  .modal-header {
    padding: 0.875rem 1rem;
  }

  .modal-header h2 {
    font-size: 0.95rem;
  }

  .restore-view {
    padding: 1rem;
  }

  .step-indicator {
    margin-bottom: 1.5rem;
  }

  .step-number {
    width: 28px;
    height: 28px;
    font-size: 0.75rem;
  }

  .step-label {
    font-size: 0.65rem;
  }
}
</style>