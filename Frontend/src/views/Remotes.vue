<template>
 <div class="remotes-view">
  <div class="header">
   <div>
    <h1>Gdrive Monitoring</h1>
    <p class="subtitle">Manage and monitor your Gdrive status</p>
   </div>
   <button @click="showTutorial = true" class="btn-tutorial">
    <span class="icon">?</span>
    Panduan
   </button>
  </div>
  
  <div v-if="isLoading" class="status-message">
   <span class="loading-dot"></span>
   Memuat data...
  </div>
  
  <div v-if="errorMessage" class="status-message error">
   {{ errorMessage }}
  </div>

  <div v-if="!isLoading && remotes.length > 0" class="table-container">
   <table class="remotes-table">
    <thead>
     <tr>
      <th>Gdrive</th>
      <th>Email</th>
      <th>Status</th>
      <th>Storage</th>
      <th>Job Runs</th>
      <th>Last Check</th>
     </tr>
    </thead>
    <tbody>
     <RemoteRow
      v-for="remote in remotes"
      :key="remote.remote_name"
      :remote="remote"
     />
    </tbody>
   </table>
  </div>

  <div v-if="!isLoading && remotes.length === 0" class="empty-state">
   <p>Belum ada remote yang terdaftar</p>
   <button @click="showTutorial = true" class="btn-empty">
    Lihat panduan menambah remote
   </button>
  </div>

  <!-- Tutorial Modal -->
  <Teleport to="body">
   <div v-if="showTutorial" class="modal-overlay" @click="showTutorial = false">
    <div class="modal-content" @click.stop>
     <div class="modal-header">
      <h2>Cara Menambah Remote Google Drive</h2>
      <button @click="showTutorial = false" class="btn-close">×</button>
     </div>

     <div class="modal-body">
      <!-- Mode Selection -->
      <div class="mode-selector">
       <button 
        @click="selectedMode = 'headless'" 
        :class="['mode-btn', { active: selectedMode === 'headless' }]"
       >
        <span class="mode-icon">🖥️</span>
        <div>
         <div class="mode-title">Headless Server</div>
         <div class="mode-desc">Server tanpa GUI</div>
        </div>
       </button>
       <button 
        @click="selectedMode = 'gui'" 
        :class="['mode-btn', { active: selectedMode === 'gui' }]"
       >
        <span class="mode-icon">🖱️</span>
        <div>
         <div class="mode-title">Desktop/GUI</div>
         <div class="mode-desc">Komputer dengan tampilan</div>
        </div>
       </button>
      </div>

      <!-- Headless Mode Steps -->
      <div v-if="selectedMode === 'headless'" class="steps">
       <div class="step">
        <div class="step-number">1</div>
        <div class="step-content">
         <h3>SSH ke server</h3>
         <div class="command-box">
          <code>ssh user@your-server-ip</code>
          <button @click="copyToClipboard('ssh user@your-server-ip')" class="btn-copy">Copy</button>
         </div>
        </div>
       </div>

       <div class="step">
        <div class="step-number">2</div>
        <div class="step-content">
         <h3>Jalankan rclone config</h3>
         <div class="command-box">
          <code>rclone config</code>
          <button @click="copyToClipboard('rclone config')" class="btn-copy">Copy</button>
         </div>
        </div>
       </div>

       <div class="step">
        <div class="step-number">3</div>
        <div class="step-content">
         <h3>Pilih New remote</h3>
         <p>Ketik <strong>n</strong> lalu Enter</p>
        </div>
       </div>

       <div class="step">
        <div class="step-number">4</div>
        <div class="step-content">
         <h3>Masukkan nama remote</h3>
         <p>Contoh: <strong>mydrive</strong> atau <strong>backup</strong></p>
         <div class="note">Hanya huruf kecil, angka, dan dash (-)</div>
        </div>
       </div>

       <div class="step">
        <div class="step-number">5</div>
        <div class="step-content">
         <h3>Pilih storage type</h3>
         <p>Ketik <strong>drive</strong> untuk Google Drive</p>
        </div>
       </div>

       <div class="step">
        <div class="step-number">6</div>
        <div class="step-content">
         <h3>Lewati Client ID & Secret</h3>
         <p>Tekan <strong>Enter</strong> (kosongkan) untuk keduanya</p>
        </div>
       </div>

       <div class="step">
        <div class="step-number">7</div>
        <div class="step-content">
         <h3>Pilih scope</h3>
         <p>Ketik <strong>1</strong> untuk full access</p>
        </div>
       </div>

       <div class="step">
        <div class="step-number">8</div>
        <div class="step-content">
         <h3>Lewati pertanyaan lain</h3>
         <p>Tekan <strong>Enter</strong> untuk:</p>
         <ul>
          <li>root_folder_id</li>
          <li>service_account_file</li>
         </ul>
        </div>
       </div>

       <div class="step">
        <div class="step-number">9</div>
        <div class="step-content">
         <h3>Edit advanced config?</h3>
         <p>Ketik <strong>n</strong> (No)</p>
        </div>
       </div>

       <div class="step highlight">
        <div class="step-number">10</div>
        <div class="step-content">
         <h3>Use auto config?</h3>
         <p><strong>Ketik n (No)</strong> - Ini penting untuk headless!</p>
        </div>
       </div>

       <div class="step">
        <div class="step-number">11</div>
        <div class="step-content">
         <h3>Buka link di browser lokal</h3>
         <p>Copy link yang muncul, buka di browser komputer Anda</p>
         <p>Login Google → Copy authorization code</p>
        </div>
       </div>

       <div class="step">
        <div class="step-number">12</div>
        <div class="step-content">
         <h3>Paste code ke terminal</h3>
         <p>Paste code yang sudah di-copy</p>
        </div>
       </div>

       <div class="step">
        <div class="step-number">13</div>
        <div class="step-content">
         <h3>Team drive?</h3>
         <p>Ketik <strong>n</strong> (No)</p>
        </div>
       </div>

       <div class="step">
        <div class="step-number">14</div>
        <div class="step-content">
         <h3>Konfirmasi</h3>
         <p>Ketik <strong>y</strong> (Yes) untuk konfirmasi</p>
        </div>
       </div>

       <div class="success-box">
        <span class="success-icon">✓</span>
        <div>
         <strong>Selesai!</strong>
         <p>Ketik <strong>q</strong> untuk keluar. Refresh halaman untuk melihat remote baru.</p>
        </div>
       </div>
      </div>

      <!-- GUI Mode Steps -->
      <div v-if="selectedMode === 'gui'" class="steps">
       <div class="step">
        <div class="step-number">1</div>
        <div class="step-content">
         <h3>Buka terminal</h3>
         <p><strong>Windows:</strong> Command Prompt / PowerShell</p>
         <p><strong>Mac/Linux:</strong> Terminal</p>
        </div>
       </div>

       <div class="step">
        <div class="step-number">2</div>
        <div class="step-content">
         <h3>Jalankan rclone config</h3>
         <div class="command-box">
          <code>rclone config</code>
          <button @click="copyToClipboard('rclone config')" class="btn-copy">Copy</button>
         </div>
        </div>
       </div>

       <div class="step">
        <div class="step-number">3</div>
        <div class="step-content">
         <h3>Pilih New remote</h3>
         <p>Ketik <strong>n</strong> lalu Enter</p>
        </div>
       </div>

       <div class="step">
        <div class="step-number">4</div>
        <div class="step-content">
         <h3>Masukkan nama remote</h3>
         <p>Contoh: <strong>mydrive</strong> atau <strong>backup</strong></p>
        </div>
       </div>

       <div class="step">
        <div class="step-number">5</div>
        <div class="step-content">
         <h3>Pilih storage type</h3>
         <p>Ketik <strong>drive</strong> untuk Google Drive</p>
        </div>
       </div>

       <div class="step">
        <div class="step-number">6</div>
        <div class="step-content">
         <h3>Lewati Client ID & Secret</h3>
         <p>Tekan <strong>Enter</strong> untuk keduanya</p>
        </div>
       </div>

       <div class="step">
        <div class="step-number">7</div>
        <div class="step-content">
         <h3>Pilih scope</h3>
         <p>Ketik <strong>1</strong> untuk full access</p>
        </div>
       </div>

       <div class="step">
        <div class="step-number">8</div>
        <div class="step-content">
         <h3>Lewati pertanyaan lain</h3>
         <p>Tekan <strong>Enter</strong> untuk root_folder_id & service_account_file</p>
        </div>
       </div>

       <div class="step">
        <div class="step-number">9</div>
        <div class="step-content">
         <h3>Edit advanced config?</h3>
         <p>Ketik <strong>n</strong> (No)</p>
        </div>
       </div>

       <div class="step highlight">
        <div class="step-number">10</div>
        <div class="step-content">
         <h3>Use auto config?</h3>
         <p><strong>Ketik y (Yes)</strong> - Browser akan terbuka otomatis</p>
        </div>
       </div>

       <div class="step">
        <div class="step-number">11</div>
        <div class="step-content">
         <h3>Login di browser</h3>
         <p>Browser otomatis terbuka → Login Google → Klik Allow</p>
         <p>Proses otomatis selesai di terminal</p>
        </div>
       </div>

       <div class="step">
        <div class="step-number">12</div>
        <div class="step-content">
         <h3>Team drive?</h3>
         <p>Ketik <strong>n</strong> (No)</p>
        </div>
       </div>

       <div class="step">
        <div class="step-number">13</div>
        <div class="step-content">
         <h3>Konfirmasi</h3>
         <p>Ketik <strong>y</strong> (Yes)</p>
        </div>
       </div>

       <div class="success-box">
        <span class="success-icon">✓</span>
        <div>
         <strong>Selesai!</strong>
         <p>Ketik <strong>q</strong> untuk keluar. Refresh halaman untuk melihat remote baru.</p>
        </div>
       </div>
      </div>

      <!-- Install Info -->
      <div class="install-info">
       <h4>Belum punya rclone?</h4>
       <div class="install-commands">
        <div class="install-item">
         <strong>Linux/Mac:</strong>
         <div class="command-box">
          <code>curl https://rclone.org/install.sh | sudo bash</code>
          <button @click="copyToClipboard('curl https://rclone.org/install.sh | sudo bash')" class="btn-copy">Copy</button>
         </div>
        </div>
        <div class="install-item">
         <strong>Windows:</strong>
         <a href="https://rclone.org/downloads/" target="_blank" class="link">Download dari rclone.org</a>
        </div>
       </div>
      </div>
     </div>
    </div>
   </div>
  </Teleport>
 </div>
</template>

<script setup>
import { ref, onMounted } from 'vue' 
import monitoringService from '@/services/monitoringService'
import RemoteRow from '@/components/RemoteRow.vue'

const remotes = ref([])
const isLoading = ref(true)
const errorMessage = ref(null)
const showTutorial = ref(false)
const selectedMode = ref('headless')

onMounted(async () => {
 try {
  const data = await monitoringService.getRemoteStatus()
  remotes.value = data
 } catch (error) {
  errorMessage.value = 'Gagal memuat data. Silakan login ulang.'
 } finally {
  isLoading.value = false
 }
})

const copyToClipboard = async (text) => {
 try {
  await navigator.clipboard.writeText(text)
  // Optional: Show toast notification
 } catch (err) {
  console.error('Failed to copy:', err)
 }
}
</script>

<style scoped>
.remotes-view {
 max-width: 1200px;
 margin: 0 auto;
 padding: 2rem 1.5rem;
}

.header {
 margin-bottom: 2.5rem;
 display: flex;
 justify-content: space-between;
 align-items: flex-start;
}

h1 {
 font-size: 1.75rem;
 font-weight: 600;
 color: #1a1a1a;
 margin: 0 0 0.5rem 0;
 letter-spacing: -0.02em;
}

.subtitle {
 font-size: 0.95rem;
 color: #666;
 margin: 0;
 font-weight: 400;
}

.btn-tutorial {
 background: #f0f0f0;
 color: #333;
 border: 1px solid #e0e0e0;
 padding: 0.625rem 1rem;
 border-radius: 6px;
 font-size: 0.875rem;
 font-weight: 500;
 cursor: pointer;
 transition: all 0.2s;
 display: flex;
 align-items: center;
 gap: 0.5rem;
}

.btn-tutorial:hover {
 background: #e5e5e5;
 border-color: #d0d0d0;
}

.btn-tutorial .icon {
 width: 1.125rem;
 height: 1.125rem;
 background: #666;
 color: #fff;
 border-radius: 50%;
 display: flex;
 align-items: center;
 justify-content: center;
 font-size: 0.75rem;
 font-weight: bold;
}

.table-container {
 background: #fff;
 border: 1px solid #e5e5e5;
 border-radius: 8px;
 overflow: hidden;
}

.remotes-table {
 width: 100%;
 border-collapse: collapse;
}

.remotes-table th {
 background: #fafafa;
 padding: 0.875rem 1rem;
 text-align: left;
 font-size: 0.8125rem;
 font-weight: bold;
 color: #000000;
 text-transform: uppercase;
 letter-spacing: 0.05em;
 border-bottom: 1px solid #e5e5e5;
}

.remotes-table td {
 padding: 1rem;
 border-bottom: 1px solid #f0f0f0;
 font-size: 0.9375rem;
 color: #333;
}

.remotes-table tbody tr:last-child td {
 border-bottom: none;
}

.remotes-table tbody tr:hover {
 background: #fafafa;
}

.status-message {
 padding: 1rem;
 border-radius: 6px;
 font-size: 0.9375rem;
 background: #f8f8f8;
 color: #666;
 display: flex;
 align-items: center;
 gap: 0.75rem;
}

.status-message.error {
 background: #fef2f2;
 color: #dc2626;
 border: 1px solid #fee2e2;
}

.loading-dot {
 width: 8px;
 height: 8px;
 background: #666;
 border-radius: 50%;
 animation: pulse 1.5s ease-in-out infinite;
}

@keyframes pulse {
 0%, 100% { opacity: 1; }
 50% { opacity: 0.3; }
}

.empty-state {
 text-align: center;
 padding: 3rem 1rem;
 color: #999;
 font-size: 0.9375rem;
}

.btn-empty {
 margin-top: 1rem;
 background: #f0f0f0;
 color: #333;
 border: 1px solid #e0e0e0;
 padding: 0.625rem 1rem;
 border-radius: 6px;
 font-size: 0.875rem;
 cursor: pointer;
 transition: all 0.2s;
}

.btn-empty:hover {
 background: #e5e5e5;
}

/* Modal Styles */
.modal-overlay {
 position: fixed;
 top: 0;
 left: 0;
 right: 0;
 bottom: 0;
 background: rgba(0, 0, 0, 0.5);
 display: flex;
 align-items: center;
 justify-content: center;
 z-index: 1000;
 padding: 1rem;
}

.modal-content {
 background: #fff;
 border-radius: 8px;
 max-width: 700px;
 width: 100%;
 max-height: 90vh;
 overflow: hidden;
 display: flex;
 flex-direction: column;
}

.modal-header {
 padding: 1.5rem;
 border-bottom: 1px solid #e5e5e5;
 display: flex;
 justify-content: space-between;
 align-items: center;
}

.modal-header h2 {
 font-size: 1.25rem;
 font-weight: 600;
 color: #1a1a1a;
 margin: 0;
}

.btn-close {
 background: none;
 border: none;
 font-size: 2rem;
 color: #999;
 cursor: pointer;
 line-height: 1;
 padding: 0;
 width: 2rem;
 height: 2rem;
 display: flex;
 align-items: center;
 justify-content: center;
 transition: color 0.2s;
}

.btn-close:hover {
 color: #333;
}

.modal-body {
 padding: 1.5rem;
 overflow-y: auto;
}

/* Mode Selector */
.mode-selector {
 display: grid;
 grid-template-columns: 1fr 1fr;
 gap: 0.75rem;
 margin-bottom: 2rem;
}

.mode-btn {
 background: #fff;
 border: 2px solid #e5e5e5;
 padding: 1rem;
 border-radius: 6px;
 cursor: pointer;
 display: flex;
 align-items: center;
 gap: 0.75rem;
 transition: all 0.2s;
 text-align: left;
}

.mode-btn:hover {
 border-color: #ccc;
}

.mode-btn.active {
 border-color: #000;
 background: #fafafa;
}

.mode-icon {
 font-size: 1.5rem;
}

.mode-title {
 font-size: 0.9375rem;
 font-weight: 600;
 color: #1a1a1a;
 margin-bottom: 0.25rem;
}

.mode-desc {
 font-size: 0.8125rem;
 color: #666;
}

/* Steps */
.steps {
 display: flex;
 flex-direction: column;
 gap: 1rem;
}

.step {
 display: flex;
 gap: 1rem;
 padding: 1rem;
 background: #fafafa;
 border-radius: 6px;
 border-left: 3px solid #e5e5e5;
}

.step.highlight {
 background: #fff9e6;
 border-left-color: #ffc107;
}

.step-number {
 flex-shrink: 0;
 width: 2rem;
 height: 2rem;
 background: #000;
 color: #fff;
 border-radius: 50%;
 display: flex;
 align-items: center;
 justify-content: center;
 font-weight: 600;
 font-size: 0.875rem;
}

.step-content h3 {
 font-size: 0.9375rem;
 font-weight: 600;
 color: #1a1a1a;
 margin: 0 0 0.5rem 0;
}

.step-content p {
 font-size: 0.875rem;
 color: #666;
 margin: 0.25rem 0;
 line-height: 1.5;
}

.step-content ul {
 margin: 0.5rem 0 0 1.25rem;
 padding: 0;
}

.step-content li {
 font-size: 0.875rem;
 color: #666;
 margin: 0.25rem 0;
}

.note {
 font-size: 0.8125rem;
 color: #666;
 background: #fff;
 padding: 0.5rem 0.75rem;
 border-radius: 4px;
 margin-top: 0.5rem;
 border: 1px solid #e5e5e5;
}

.command-box {
 background: #1a1a1a;
 color: #4ade80;
 padding: 0.75rem;
 border-radius: 4px;
 font-family: 'Courier New', monospace;
 font-size: 0.8125rem;
 margin-top: 0.5rem;
 display: flex;
 justify-content: space-between;
 align-items: center;
 gap: 0.75rem;
}

.command-box code {
 flex: 1;
 word-break: break-all;
}

.btn-copy {
 background: #333;
 color: #fff;
 border: none;
 padding: 0.375rem 0.75rem;
 border-radius: 4px;
 font-size: 0.75rem;
 cursor: pointer;
 white-space: nowrap;
 transition: background 0.2s;
}

.btn-copy:hover {
 background: #555;
}

.success-box {
 background: #f0fdf4;
 border: 1px solid #86efac;
 border-radius: 6px;
 padding: 1rem;
 display: flex;
 gap: 1rem;
 margin-top: 1rem;
}

.success-icon {
 flex-shrink: 0;
 width: 2rem;
 height: 2rem;
 background: #22c55e;
 color: #fff;
 border-radius: 50%;
 display: flex;
 align-items: center;
 justify-content: center;
 font-weight: bold;
}

.success-box strong {
 display: block;
 color: #166534;
 margin-bottom: 0.25rem;
 font-size: 0.9375rem;
}

.success-box p {
 color: #166534;
 font-size: 0.875rem;
 margin: 0;
}

/* Install Info */
.install-info {
 background: #f8f9fa;
 border-radius: 6px;
 padding: 1rem;
 margin-top: 2rem;
}

.install-info h4 {
 font-size: 0.9375rem;
 font-weight: 600;
 color: #1a1a1a;
 margin: 0 0 1rem 0;
}

.install-commands {
 display: flex;
 flex-direction: column;
 gap: 1rem;
}

.install-item strong {
 display: block;
 font-size: 0.875rem;
 color: #333;
 margin-bottom: 0.5rem;
}

.link {
 color: #0066cc;
 text-decoration: none;
 font-size: 0.875rem;
}

.link:hover {
 text-decoration: underline;
}

/* Responsive */
@media (max-width: 768px) {
 .remotes-view {
  padding: 1.5rem 1rem;
 }
 
 .header {
  flex-direction: column;
  gap: 1rem;
 }

 .btn-tutorial {
  width: 100%;
  justify-content: center;
 }
 
 h1 {
  font-size: 1.5rem;
 }
 
 .table-container {
  overflow-x: auto;
 }
 
 .remotes-table {
  min-width: 600px;
 }

 .mode-selector {
  grid-template-columns: 1fr;
 }

 .modal-content {
  max-height: 95vh;
 }
}
</style>