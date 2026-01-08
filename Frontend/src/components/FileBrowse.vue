<template>
  <div v-if="isVisible" class="modal-overlay" @click.self="close">
    <div class="modal-content">
      <div class="modal-header">
        <h2>Browse Files - {{ remoteName }}</h2>
        <button class="close-btn" @click="close">×</button>
      </div>
      
      <div class="modal-body">
        <FileBrowser 
          :remoteName="remoteName"
          :initialPath="initialPath"
          @select-file="handleSelectFile"
          @navigate="handleNavigate"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import FileBrowser from '@/components/GDriveBrowser.vue'

const props = defineProps({
  isVisible: {
    type: Boolean,
    default: false
  },
  remoteName: {
    type: String,
    required: true
  },
  initialPath: {
    type: String,
    default: '/'
  }
})

const emit = defineEmits(['close', 'select-file', 'navigate'])

function close() {
  emit('close')
}

function handleSelectFile(fileData) {
  emit('select-file', fileData)
}

function handleNavigate(navData) {
  emit('navigate', navData)
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 20px;
}

.modal-content {
  background: #fff;
  border-radius: 8px;
  border: 1px solid #e5e5e5;
  width: 90%;
  max-width: 800px;
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

.modal-header h2 {
  margin: 0;
  font-size: 1.125rem;
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
  transition: all 0.2s;
}

.close-btn:hover {
  background: #f5f5f5;
  border-color: #1a1a1a;
  color: #1a1a1a;
}

.modal-body {
  padding: 1.5rem;
  overflow: auto;
  flex: 1;
}

@media (max-width: 768px) {
  .modal-content {
    width: 95%;
    max-height: 85vh;
  }
  
  .modal-body {
    padding: 1rem;
  }
}
</style>