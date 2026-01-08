<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="isVisible" class="modal-overlay" @click.self="handleClose">
        <div class="modal-container">
          <!-- HEADER -->
          <div class="modal-header">
            <div class="header-left">
              <h3>Script Preview</h3>
              <span class="job-id">Job ID: {{ jobId }}</span>
            </div>
            <button @click="handleClose" class="close-btn" aria-label="Close">
              <span>×</span>
            </button>
          </div>

          <!-- TOOLBAR -->
          <div class="modal-toolbar">
            <div class="toolbar-left">
              <span class="language-badge">{{ detectedLanguage }}</span>
              <span class="line-count">{{ lineCount }} lines</span>
            </div>
            <div class="toolbar-right">
              <button @click="downloadScript" class="toolbar-btn">
                Download
              </button>
            </div>
          </div>

          <!-- CODE CONTENT -->
          <div class="modal-body">
            <div class="code-container">
              <!-- Line Numbers -->
              <div class="line-numbers">
                <div v-for="n in lineCount" :key="n" class="line-number">
                  {{ n }}
                </div>
              </div>

              <!-- Code with Syntax Highlighting -->
              <pre class="code-content"><code 
                ref="codeElement" 
                :class="`language-${detectedLanguage}`"
                v-html="highlightedCode"
              ></code></pre>
            </div>
          </div>

          <!-- FOOTER -->
          <div class="modal-footer">
            <div class="footer-info">
              <span>Scroll to view more code</span>
            </div>
            <button @click="handleClose" class="btn-close">Close</button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { ref, computed, watch } from 'vue';
import hljs from 'highlight.js';

// Import theme CSS
import 'highlight.js/styles/atom-one-dark.css';

const props = defineProps({
  isVisible: {
    type: Boolean,
    required: true
  },
  jobId: {
    type: [String, Number],
    default: null
  },
  scriptContent: {
    type: String,
    default: ''
  }
});

const emit = defineEmits(['close']);

const codeElement = ref(null);
const isCopied = ref(false);

// Auto-detect language from script content
const detectedLanguage = computed(() => {
  const content = props.scriptContent.toLowerCase();
  
  if (content.includes('#!/bin/bash') || content.includes('#!/bin/sh')) {
    return 'bash';
  } else if (content.includes('<?php')) {
    return 'php';
  } else if (content.includes('import ') || content.includes('def ')) {
    return 'python';
  } else if (content.includes('function') || content.includes('const ') || content.includes('let ')) {
    return 'javascript';
  } else if (content.includes('mysqldump') || content.includes('rclone')) {
    return 'bash';
  }
  
  return 'bash'; // default
});

// Count lines
const lineCount = computed(() => {
  if (!props.scriptContent) return 0;
  return props.scriptContent.split('\n').length;
});

// Highlighted code
const highlightedCode = computed(() => {
  if (!props.scriptContent) return '';
  
  try {
    const result = hljs.highlight(props.scriptContent, {
      language: detectedLanguage.value
    });
    return result.value;
  } catch (error) {
    console.error('Highlight error:', error);
    return props.scriptContent
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;');
  }
});

// Watch for visibility changes
watch(() => props.isVisible, (newVal) => {
  if (newVal) {
    document.body.style.overflow = 'hidden';
  } else {
    document.body.style.overflow = '';
    isCopied.value = false;
  }
});

// Close modal
function handleClose() {
  emit('close');
}


// Download script
function downloadScript() {
  const blob = new Blob([props.scriptContent], { type: 'text/plain' });
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  
  link.href = url;
  link.download = `job_${props.jobId}_script.sh`;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
}

// Handle ESC key
function handleKeydown(e) {
  if (e.key === 'Escape' && props.isVisible) {
    handleClose();
  }
}

// Add/remove event listener
watch(() => props.isVisible, (newVal) => {
  if (newVal) {
    window.addEventListener('keydown', handleKeydown);
  } else {
    window.removeEventListener('keydown', handleKeydown);
  }
});
</script>

<style scoped>
/* Modal Overlay */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  padding: 20px;
}

/* Modal Container */
.modal-container {
  background: #1e1e1e;
  border-radius: 8px;
  border: 1px solid #3e3e42;
  width: 90%;
  max-width: 1000px;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* Header */
.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.25rem 1.5rem;
  background: #252526;
  border-bottom: 1px solid #3e3e42;
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
  color: #d4d4d4;
}

.job-id {
  background: #3e3e42;
  padding: 0.25rem 0.625rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 500;
  color: #858585;
}

.close-btn {
  background: transparent;
  border: 1px solid #3e3e42;
  color: #858585;
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
  background: #3e3e42;
  color: #d4d4d4;
}

/* Toolbar */
.modal-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 1.5rem;
  background: #252526;
  border-bottom: 1px solid #3e3e42;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.language-badge {
  background: #1a1a1a;
  color: #858585;
  padding: 0.25rem 0.625rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.line-count {
  color: #858585;
  font-size: 0.8125rem;
}

.toolbar-right {
  display: flex;
  gap: 0.5rem;
}

.toolbar-btn {
  background: transparent;
  border: 1px solid #3e3e42;
  color: #858585;
  padding: 0.375rem 0.875rem;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.8125rem;
  transition: all 0.2s;
  font-weight: 500;
}

.toolbar-btn:hover {
  background: #3e3e42;
  color: #d4d4d4;
}

.toolbar-btn.copied {
  background: #22c55e;
  border-color: #22c55e;
  color: white;
}

/* Body (Code Container) */
.modal-body {
  flex: 1;
  overflow: auto;
  background: #1e1e1e;
}

.code-container {
  display: flex;
  min-height: 100%;
}

/* Line Numbers */
.line-numbers {
  background: #1e1e1e;
  color: #858585;
  padding: 1rem 0.75rem;
  text-align: right;
  user-select: none;
  border-right: 1px solid #3e3e42;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.6;
  min-width: 50px;
}

.line-number {
  height: 20.8px;
}

/* Code Content */
.code-content {
  flex: 1;
  margin: 0;
  padding: 1rem 1.25rem;
  background: transparent;
  color: #d4d4d4;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.6;
  overflow-x: auto;
  white-space: pre;
  tab-size: 4;
}

.code-content code {
  display: block;
  color: inherit;
}

/* Custom scrollbar */
.modal-body::-webkit-scrollbar {
  width: 12px;
  height: 12px;
}

.modal-body::-webkit-scrollbar-track {
  background: #1e1e1e;
}

.modal-body::-webkit-scrollbar-thumb {
  background: #424242;
  border-radius: 6px;
}

.modal-body::-webkit-scrollbar-thumb:hover {
  background: #4e4e4e;
}

/* Footer */
.modal-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.5rem;
  background: #252526;
  border-top: 1px solid #3e3e42;
}

.footer-info {
  color: #858585;
  font-size: 0.8125rem;
}

.btn-close {
  background: #1a1a1a;
  border: 1px solid #3e3e42;
  color: #d4d4d4;
  padding: 0.5rem 1.25rem;
  border-radius: 4px;
  cursor: pointer;
  font-weight: 500;
  font-size: 0.875rem;
  transition: all 0.2s;
}

.btn-close:hover {
  background: #3e3e42;
  color: #fff;
}

/* Transitions */
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.2s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-active .modal-container,
.modal-leave-active .modal-container {
  transition: transform 0.2s ease;
}

.modal-enter-from .modal-container,
.modal-leave-to .modal-container {
  transform: scale(0.95);
}

/* Responsive */
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
  
  .modal-toolbar {
    flex-direction: column;
    gap: 0.75rem;
    align-items: stretch;
  }
  
  .toolbar-right {
    justify-content: stretch;
  }
  
  .toolbar-btn {
    flex: 1;
    justify-content: center;
  }
  
  .line-numbers {
    min-width: 40px;
    padding: 1rem 0.5rem;
  }
  
  .code-content {
    font-size: 12px;
    padding: 1rem 0.75rem;
  }
}
</style>