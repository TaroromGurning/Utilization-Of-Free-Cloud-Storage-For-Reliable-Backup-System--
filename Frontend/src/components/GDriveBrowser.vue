<template>
  <div class="file-browser">
    <!-- Header -->
    <div class="browser-header">
      <div class="header-title">
        <h3>{{ remoteName }}</h3>
      </div>
      <button 
        v-if="currentPath !== '/'" 
        @click="goBack" 
        class="btn-back"
      >
        Back
      </button>
    </div>

    <!-- Breadcrumb -->
    <div class="breadcrumb">
      <button 
        @click="navigateTo('/')" 
        :class="{ active: currentPath === '/' }"
        class="breadcrumb-item"
      >
        Root
      </button>
      <template v-for="(segment, index) in pathSegments" :key="index">
        <span class="separator">/</span>
        <button 
          @click="navigateTo(buildPath(index))"
          :class="{ active: buildPath(index) === currentPath }"
          class="breadcrumb-item"
        >
          {{ segment }}
        </button>
      </template>
    </div>

    <!-- Selected Item Info -->
    <div v-if="selectedItem" class="selected-info">
      <div class="selected-item">
        <span class="item-type">{{ selectedItem.is_dir ? 'Folder' : 'File' }}</span>
        <span class="item-name">{{ selectedItem.name }}</span>
        <span class="item-size" v-if="!selectedItem.is_dir">{{ formatFileSize(selectedItem.size) }}</span>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="isLoading" class="state-container loading">
      <div class="spinner"></div>
      <p>Loading...</p>
    </div>

    <!-- Error -->
    <div v-else-if="error" class="state-container error">
      <p class="error-text">{{ error }}</p>
      <button @click="loadFiles" class="btn-retry">Retry</button>
    </div>

    <!-- Empty -->
    <div v-else-if="files.length === 0" class="state-container empty">
      <p>No items found</p>
    </div>

    <!-- Files List -->
    <div v-else class="files-list">
      <div 
        v-for="file in files" 
        :key="file.path"
        :class="['file-row', { 
          'is-dir': file.is_dir,
          'is-selected': selectedItem?.path === file.path
        }]"
      >
        <!-- Folder -->
        <template v-if="file.is_dir">
          <div 
            class="file-content"
            @click="selectItem(file)"
            @dblclick="navigateTo(file.path)"
          >
            <span class="file-type-icon">📂</span>
            <div class="file-meta">
              <div class="file-name">{{ file.name }}</div>
              <div class="file-time">{{ formatDate(file.mod_time) }}</div>
            </div>
          </div>
        </template>

        <!-- File -->
        <template v-else>
          <div 
            class="file-content"
            @click="selectItem(file)"
          >
            <span class="file-type-icon">📄</span>
            <div class="file-meta">
              <div class="file-name">{{ file.name }}</div>
              <div class="file-time">{{ formatFileSize(file.size) }} • {{ formatDate(file.mod_time) }}</div>
            </div>
          </div>
        </template>
      </div>

      <!-- Summary -->
      <div class="files-summary">
        <span>{{ files.length }} items</span>
        <span>•</span>
        <span>{{ formatFileSize(totalSize) }}</span>
      </div>
    </div>

    <!-- Help Text -->
    <div class="help-text">
      <small>Click to select • Double-click folder to open</small>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue';
import driveService from '@/services/driveService';

// Props
const props = defineProps({
  remoteName: {
    type: String,
    default: 'Remote'
  },
  initialPath: {
    type: String,
    default: '/'
  }
});

// Emits
const emit = defineEmits(['select-file', 'navigate']);

// State
const currentPath = ref('/');
const files = ref([]);
const isLoading = ref(false);
const error = ref(null);
const totalSize = ref(0);
const selectedItem = ref(null);

// Computed
const pathSegments = computed(() => {
  if (currentPath.value === '/') return [];
  return currentPath.value
    .split('/')
    .filter(s => s !== '')
    .map(s => s);
});

// Load files
async function loadFiles() {
  isLoading.value = true;
  error.value = null;
  files.value = [];
  totalSize.value = 0;
  selectedItem.value = null;

  try {
    const response = await driveService.browseFiles(props.remoteName, currentPath.value);
    
    files.value = response.files || [];
    totalSize.value = response.total_size || 0;
    
    emit('navigate', {
      remote: props.remoteName,
      path: currentPath.value,
      filesCount: files.value.length
    });

  } catch (err) {
    error.value = err.response?.data?.error || err.message || 'Failed to load';
  } finally {
    isLoading.value = false;
  }
}

// Select item
function selectItem(file) {
  selectedItem.value = file;
  
  emit('select-file', {
    name: file.name,
    path: file.path,
    size: file.size,
    is_dir: file.is_dir,
    remote: props.remoteName
  });
}

// Navigate
function navigateTo(path) {
  currentPath.value = path;
  loadFiles();
}

// Go back
function goBack() {
  const parts = currentPath.value
    .split('/')
    .filter(s => s !== '')
    .slice(0, -1);
  
  const parentPath = parts.length === 0 ? '/' : '/' + parts.join('/');
  navigateTo(parentPath);
}

// Build path
function buildPath(index) {
  const parts = pathSegments.value.slice(0, index + 1);
  return '/' + parts.join('/');
}

// Format file size
function formatFileSize(bytes) {
  if (!bytes || bytes === 0) return '0 B';
  
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

// Format date
function formatDate(dateString) {
  if (!dateString) return '-';
  
  try {
    const date = new Date(dateString);
    return date.toLocaleDateString('id-ID', {
      year: '2-digit',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  } catch (e) {
    return dateString;
  }
}

// Lifecycle
onMounted(() => {
  currentPath.value = props.initialPath;
  loadFiles();
});
</script>

<style scoped>
.file-browser {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  background: #fff;
  border: 1px solid #e5e5e5;
  border-radius: 6px;
  padding: 1rem;
  max-height: 600px;
}

/* Header */
.browser-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid #e5e5e5;
}

.header-title h3 {
  margin: 0;
  font-size: 0.95rem;
  font-weight: 600;
  color: #1a1a1a;
}

.btn-back {
  padding: 0.35rem 0.75rem;
  font-size: 0.8rem;
  border: 1px solid #d5d5d5;
  background: #f5f5f5;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.15s;
  color: #666;
  font-weight: 500;
}

.btn-back:hover {
  background: #efefef;
  border-color: #ccc;
}

/* Breadcrumb */
.breadcrumb {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  font-size: 0.8rem;
  overflow-x: auto;
  padding: 0.25rem 0;
}

.breadcrumb-item {
  background: none;
  border: none;
  color: #0066cc;
  cursor: pointer;
  padding: 0.2rem 0.4rem;
  border-radius: 3px;
  transition: all 0.15s;
  white-space: nowrap;
  font-size: 0.8rem;
}

.breadcrumb-item:hover {
  background: #f0f0f0;
}

.breadcrumb-item.active {
  color: #333;
  font-weight: 600;
  background: #f5f5f5;
}

.separator {
  color: #bbb;
  margin: 0 0.1rem;
}

/* Selected Info */
.selected-info {
  padding: 0.6rem;
  background: #f0f8ff;
  border: 1px solid #d4e8f5;
  border-radius: 4px;
  font-size: 0.8rem;
}

.selected-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.item-type {
  display: inline-block;
  padding: 0.2rem 0.4rem;
  background: #0066cc;
  color: white;
  border-radius: 3px;
  font-size: 0.7rem;
  font-weight: 600;
  min-width: 45px;
  text-align: center;
}

.item-name {
  flex: 1;
  font-weight: 500;
  color: #1a1a1a;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.item-size {
  color: #666;
  font-size: 0.75rem;
  flex-shrink: 0;
}

/* State Containers */
.state-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  flex: 1;
}

.state-container p {
  margin: 0.5rem 0 0 0;
  color: #666;
  font-size: 0.85rem;
}

.loading {
  gap: 0.75rem;
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

.error {
  gap: 0.75rem;
}

.error-text {
  color: #c41e3a;
  text-align: center;
  margin: 0;
}

.btn-retry {
  padding: 0.4rem 0.8rem;
  font-size: 0.8rem;
  background: #c41e3a;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-weight: 500;
  transition: all 0.15s;
}

.btn-retry:hover {
  background: #a01730;
}

.empty {
  color: #999;
}

/* Files List */
.files-list {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  overflow-y: auto;
  min-height: 200px;
}

.files-list::-webkit-scrollbar {
  width: 5px;
}

.files-list::-webkit-scrollbar-track {
  background: transparent;
}

.files-list::-webkit-scrollbar-thumb {
  background: #d0d0d0;
  border-radius: 3px;
}

.files-list::-webkit-scrollbar-thumb:hover {
  background: #b0b0b0;
}

/* File Row */
.file-row {
  padding: 0.6rem;
  border: 1px solid #e5e5e5;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.15s;
  background: #fff;
}

.file-row:hover {
  background: #fafafa;
  border-color: #d0d0d0;
}

.file-row.is-selected {
  background: #f0f8ff;
  border-color: #0066cc;
}

.file-content {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  min-width: 0;
}

.file-type-icon {
  font-size: 1.1rem;
  flex-shrink: 0;
}

.file-meta {
  flex: 1;
  min-width: 0;
}

.file-name {
  font-size: 0.85rem;
  font-weight: 500;
  color: #1a1a1a;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.file-time {
  font-size: 0.75rem;
  color: #999;
  margin-top: 0.1rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Summary */
.files-summary {
  padding-top: 0.6rem;
  border-top: 1px solid #e5e5e5;
  font-size: 0.75rem;
  color: #999;
  display: flex;
  gap: 0.3rem;
}

/* Help Text */
.help-text {
  padding: 0.5rem 0;
  font-size: 0.75rem;
  color: #999;
  text-align: center;
}

/* Responsive */
@media (max-width: 640px) {
  .file-browser {
    padding: 0.75rem;
    max-height: 400px;
  }

  .file-name {
    font-size: 0.8rem;
  }

  .file-time {
    font-size: 0.7rem;
  }
}
</style>