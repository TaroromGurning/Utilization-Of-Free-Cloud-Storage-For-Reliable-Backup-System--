<template>
 <div class="script-editor">
  <!-- Pre-Script Section -->
  <div class="script-section">
   <div class="section-header">
    <h4>📝 Pre-Script</h4>
    <p class="section-desc">Script yang dijalankan sebelum backup</p>
   </div>
  
   <div class="editor-box">
    <textarea
     v-model="preScript"
     @input="updatePreScript"
     class="script-textarea"
     placeholder="#!/bin/bash
# Example: Database dump
mysqldump -u user -p password database > /tmp/backup.sql
gzip /tmp/backup.sql"
     spellcheck="false"
    ></textarea>
   
    <!-- Preview dengan Highlighting -->
    <pre class="script-preview"><code v-html="highlightCode(preScript)"></code></pre>
   </div>
  </div>

  <!-- Rclone Command Preview -->
  <div class="script-section">
   <div class="section-header">
    <h4>⚙️ Rclone Command</h4>
    <p class="section-desc">Command yang akan dijalankan</p>
   </div>
  
   <div class="command-box">
    <pre><code v-html="generateRcloneCommand()"></code></pre>
   </div>
  </div>

  <!-- Post-Script Section -->
  <div class="script-section">
   <div class="section-header">
    <h4>📝 Post-Script</h4>
    <p class="section-desc">Script yang dijalankan setelah backup sukses</p>
   </div>
  
   <div class="editor-box">
    <textarea
     v-model="postScript"
     @input="updatePostScript"
     class="script-textarea"
     placeholder="#!/bin/bash
# Cleanup
rm /tmp/backup.sql"
     spellcheck="false"
    ></textarea>
   
    <pre class="script-preview"><code v-html="highlightCode(postScript)"></code></pre>
   </div>
  </div>

  <!-- Execution Flow -->
  <div class="execution-flow">
   <h4>📊 Execution Flow</h4>
   <div class="flow">
    <div class="flow-item" :class="{ active: preScript }">
     <span class="flow-num">1</span>
     <span class="flow-label">Pre-Script</span>
    </div>
    <span class="flow-arrow">→</span>
    <div class="flow-item">
     <span class="flow-num">2</span>
     <span class="flow-label">Rclone Upload</span>
    </div>
    <span class="flow-arrow">→</span>
    <div class="flow-item" :class="{ active: postScript }">
     <span class="flow-num">3</span>
     <span class="flow-label">Post-Script</span>
    </div>
   </div>
  </div>
 </div>
</template>

<script setup>
import { ref } from 'vue'

const props = defineProps({
 preScriptContent: String,
 postScriptContent: String,
 sourcePath: String,
 remoteName: String,
 destinationPath: String,
 rcloneMode: {
  type: String,
  default: 'copy'
 }
})

const emit = defineEmits(['update:preScript', 'update:postScript'])

const preScript = ref(props.preScriptContent || '')
const postScript = ref(props.postScriptContent || '')

// ============================================
// Syntax Highlighting & Rclone Generation
// ============================================

// Highlight Code (Simple Bash Highlighting)
function highlightCode(code) {
 if (!code) return ''

 let html = escapeHtml(code)

 // Daftar keyword Bash yang aman
 const keywords = [
  "echo", "mkdir", "rm", "mv", "cp", "cd", "chmod", "chown",
  "cat", "grep", "sed", "awk", "if", "then", "else", "fi",
  "for", "do", "done", "while", "function"
 ];
  
 // Buat Regex dari daftar keyword
 const keywordRegex = new RegExp(`\\b(${keywords.join("|")})\\b`, "g");

 // 1. Comments (Termasuk Shebang: #!)
 html = html.replace(/(#.*?)$/gm, '<mark class="comment">$1</mark>')

 // 2. Strings (Harus dilakukan sebelum keyword agar tidak rusak)
 html = html.replace(/(["'])(.*?)\1/g, '<mark class="string">$1$2$1</mark>')

 // 3. Keywords
 html = html.replace(keywordRegex, '<mark class="keyword">$&</mark>')

 // 4. Variables
 html = html.replace(/\$([A-Za-z_][A-Za-z0-9_]*)/g, '<mark class="variable">$$1</mark>')

 // 5. Flags
 html = html.replace(/(-[a-zA-Z])/g, '<mark class="flag">$1</mark>')

 return html
}

// Generate Rclone Command (Dynamic Preview)
function generateRcloneCommand() {
 const { sourcePath, remoteName, destinationPath, rcloneMode } = props;
  
 if (!sourcePath || !remoteName || !destinationPath) {
  return '<span class="error">⚠️ Lengkapi Source Path, Remote Name, dan Destination Path</span>'
 }

 // Gunakan quotes di path untuk menangani spasi
 const source = `"${sourcePath}"`;
 const destination = `"${remoteName}:${destinationPath}"`;
  
 const mode = rcloneMode.toLowerCase() || 'copy'
 const cmd = `rclone ${mode} ${source} ${destination} --checksum`

 let html = escapeHtml(cmd)
 
  // Highlighting Rclone Command
 html = html.replace(/\brclone\b/, '<span class="keyword">rclone</span>')
 html = html.replace(new RegExp(`\\b${mode}\\b`), `<span class="command">${mode}</span>`)
 html = html.replace(/--checksum/g, '<span class="flag">--checksum</span>') // Highlight flag
 html = html.replace(/("[^"]*")/g, '<span class="string">$1</span>') // Highlight strings

 return html
}

function escapeHtml(text) {
 const div = document.createElement('div')
 div.textContent = text
 return div.innerHTML
}

function updatePreScript() {
 emit('update:preScript', preScript.value)
}

function updatePostScript() {
 emit('update:postScript', postScript.value)
}
</script>

<style scoped>
/*
==================================================
STYLES FOR SCRIPT EDITOR COMPONENT
==================================================
*/

/* Variables used in the component */
.script-editor {
 display: flex;
 flex-direction: column;
 gap: 1.5rem;
 margin-bottom: 1.5rem;
}

/* Section */
.script-section {
 background: #fff;
 border: 1px solid #e5e5e5;
 border-radius: 8px;
 padding: 1.25rem;
 overflow: hidden;
}

.section-header {
 margin-bottom: 1rem;
}

.section-header h4 {
 margin: 0 0 0.25rem 0;
 font-size: 0.95rem;
 color: #1a1a1a;
 font-weight: 600;
}

.section-desc {
 margin: 0;
 font-size: 0.8rem;
 color: #999;
}

/* Editor Box (Simulasi Floating/Transparent Textarea) */
.editor-box {
 position: relative;
 background: #f5f5f5;
 border-radius: 6px;
 overflow: hidden;
 min-height: 120px;
}

/* Textarea - Textnya TRANSPARAN, hanya caret yang terlihat */
.script-textarea {
 position: absolute;
 top: 0;
 left: 0;
 width: 100%;
 height: 100%;
 padding: 0.75rem;
 border: none;
 background: transparent;
 color: transparent; /* ⭐ KRITIS: Membuat teks transparan */
 caret-color: #1a1a1a; /* Warna kursor */
 font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
 font-size: 0.8rem;
 line-height: 1.6;
 resize: vertical;
 z-index: 2;
 box-sizing: border-box;
}

.script-textarea:focus {
 outline: none;
}

/* Preview - Menampilkan syntax highlighted code */
.script-preview {
 display: block;
 padding: 0.75rem;
 margin: 0;
 background: #f5f5f5;
 font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
 font-size: 0.8rem;
 line-height: 1.6;
 color: #1a1a1a;
 white-space: pre-wrap;
 word-wrap: break-word;
 pointer-events: none; /* ⭐ KRITIS: Agar klik tembus ke textarea di bawahnya */
 min-height: 120px;
 box-sizing: border-box;
}

.script-preview code {
 display: block;
}

/* Syntax Highlighting Colors */
mark.keyword {
 color: #d32f2f;
 font-weight: 600;
 background: none;
}

mark.command {
 color: #f57c00;
 font-weight: 600;
 background: none;
}

mark.comment {
 color: #999;
 font-style: italic;
 background: none;
}

mark.string {
 color: #388e3c;
 background: none;
}

mark.variable {
 color: #0288d1;
 font-weight: 500;
 background: none;
}

mark.flag {
 color: #6f42c1;
 background: none;
}

span.error {
 color: #d32f2f;
}

/* Command Box */
.command-box {
 background: #1a1a1a;
 border-radius: 6px;
 padding: 1rem;
 overflow-x: auto;
 font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
 font-size: 0.8rem;
 line-height: 1.6;
}

.command-box pre {
 margin: 0;
 color: #22c55e;
 white-space: pre-wrap;
 word-wrap: break-word;
}

.command-box code {
 display: block;
}

/* Execution Flow */
.execution-flow {
 background: #f0f8ff;
 border: 1px solid #b3d9ff;
 border-radius: 8px;
 padding: 1.25rem;
}

.execution-flow h4 {
 margin: 0 0 1rem 0;
 font-size: 0.95rem;
 color: #1a1a1a;
 font-weight: 600;
}

.flow {
 display: flex;
 align-items: center;
 justify-content: space-between;
 gap: 0.75rem;
 flex-wrap: wrap;
}

.flow-item {
 display: flex;
 flex-direction: column;
 align-items: center;
 gap: 0.375rem;
 padding: 0.75rem 1rem;
 background: #fff;
 border: 1px solid #e5e5e5;
 border-radius: 6px;
 flex: 1;
 min-width: 90px;
 text-align: center;
}

.flow-item.active {
 background: #e8f5e9;
 border-color: #81c784;
}

.flow-num {
 font-size: 1.25rem;
 font-weight: 600;
 color: #0288d1;
}

.flow-label {
 font-size: 0.75rem;
 color: #666;
 font-weight: 500;
}

.flow-arrow {
 color: #0288d1;
 font-size: 1.25rem;
 font-weight: bold;
 flex: 0 0 auto;
}

/* Responsive */
@media (max-width: 768px) {
 .script-section {
  padding: 1rem;
 }

 .script-textarea,
 .script-preview {
  font-size: 0.75rem;
  padding: 0.5rem;
  min-height: 100px;
 }

 .flow {
  gap: 0.5rem;
 }

 .flow-item {
  min-width: 70px;
  padding: 0.5rem 0.75rem;
  font-size: 0.75rem;
 }

 .flow-arrow {
  writing-mode: vertical-rl;
  transform: rotate(180deg);
  font-size: 1rem;
 }
}
</style>