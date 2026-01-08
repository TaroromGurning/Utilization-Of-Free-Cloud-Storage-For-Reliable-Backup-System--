<template>
    <tr>
        <!-- ✅ UPDATED: Remote Name Column - Clickable untuk browse -->
        <td>
            <div class="remote-name-wrapper">
                <button 
                    @click="openFileBrowser" 
                    class="remote-name-btn"
                    :title="`Browse files in ${remote.remote_name}`"
                >
                    <span class="drive-icon"></span>
                    <strong>{{ remote.remote_name }}</strong>
                </button>
            </div>
            <div v-if="remote.system_message" class="warning">
                {{ remote.system_message }}
            </div>
        </td>
        
        <!-- ✅ UNCHANGED: Email Column with Truncation and Tooltip -->
        <td>
            <div v-if="remote.email" class="email-cell" :title="remote.email">
                <span class="email-icon">📧</span>
                <span class="email-text">{{ truncatedEmail }}</span>
            </div>
            <div v-else class="email-cell">
                <span class="email-placeholder">No email</span>
            </div>
        </td>
        
        <!-- ✅ UNCHANGED: Status Column -->
        <td>
            <span class="status" :class="getStatusClass(remote.status_connect)">
                {{ remote.status_connect }}
            </span>
        </td>

        <!-- ✅ UNCHANGED: Storage Column -->
        <td>
            <div class="storage-info">
                <span>{{ usedFormatted }} / {{ totalFormatted }} GB</span>
                <div class="storage-row">
                    <div class="storage-bar">
                        <div 
                            class="bar-fill" 
                            :style="{ width: usagePercentage + '%' }"
                        ></div>
                    </div>
                    <span class="percentage">{{ usagePercentage.toFixed(0) }}%</span>
                </div>
            </div>
        </td>
        
        <!-- ✅ UNCHANGED: Active Jobs Column -->
        <td class="center">{{ remote.active_job_count }}</td>
        
        <!-- ✅ UNCHANGED: Last Checked Column -->
        <td class="text-muted">{{ formatLastChecked(remote.last_checked_at) }}</td>
    </tr>

    <!-- ✅ NEW: File Browser Modal -->
    <FileBrowserModal
        :isVisible="showFileBrowser"
        :remoteName="remote.remote_name"
        @close="showFileBrowser = false"
        @select-file="handleFileSelect"
    />
</template>

<script setup>
import { ref, computed } from 'vue';
import FileBrowserModal from '@/components/FileBrowse.vue';

const props = defineProps({
    remote: {
        type: Object,
        required: true
    }
});

// ✅ NEW: State untuk File Browser Modal
const showFileBrowser = ref(false);

// ✅ NEW: Function untuk open file browser
function openFileBrowser() {
    showFileBrowser.value = true;
}

// ✅ NEW: Function untuk handle file selection (optional)
function handleFileSelect(fileData) {
    console.log('File selected:', fileData);
    // Bisa tambah logic lain jika diperlukan
}

// ✅ UNCHANGED: Existing computed properties
const used = computed(() => props.remote.used_storage_gb || 0);
const total = computed(() => props.remote.total_storage_gb || 0);

const usagePercentage = computed(() => {
    if (!total.value || total.value === 0) return 0;
    return (used.value / total.value) * 100;
});

const usedFormatted = computed(() => {
    if (used.value > 0 && used.value < 0.01) return '<0.01';
    return used.value.toFixed(2);
});

const totalFormatted = computed(() => total.value.toFixed(2));

// ✅ UNCHANGED: Truncate email for display
const truncatedEmail = computed(() => {
    const email = props.remote.email || '';
    const maxLength = 25;
    
    if (email.length <= maxLength) {
        return email;
    }
    
    const [username, domain] = email.split('@');
    
    if (!domain) return email;
    
    if (username.length > 15) {
        return `${username.substring(0, 12)}...@${domain}`;
    }
    
    if (domain.length > 15) {
        return `${username}@${domain.substring(0, 12)}...`;
    }
    
    return email;
});

// ✅ UNCHANGED: Helper function for status class
function getStatusClass(status) {
    if (!status) return 'pending';
    return status.toLowerCase();
}

// ✅ UNCHANGED: Format date function
function formatLastChecked(isoDate) {
    if (!isoDate) return 'N/A';
    try {
        const date = new Date(isoDate);
        return date.toLocaleDateString('en-GB', { 
            year: 'numeric', 
            month: '2-digit', 
            day: '2-digit', 
            hour: '2-digit', 
            minute: '2-digit' 
        });
    } catch (e) {
        return isoDate;
    }
}
</script>

<style scoped>
/* ✅ NEW: Remote Name Button Styles */
.remote-name-btn {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    background: none;
    border: none;
    padding: 0.5rem 0.75rem;
    cursor: pointer;
    border-radius: 6px;
    transition: all 0.2s ease;
    font-size: 0.90rem;
    margin: -0.5rem 0;
}


.remote-name-btn:active {
    transform: translateX(2px);
}

.drive-icon {
    font-size: 1.25rem;
    flex-shrink: 0;
}

.remote-name-btn strong {
    color: #3498db;
    font-weight: 600;
    text-decoration: underline;
    text-decoration-style: dotted;
    text-underline-offset: 3px;
    transition: color 0.2s ease;
}

.remote-name-btn:hover strong {
    color: #2980b9;
}

/* ✅ UPDATED: Simplified remote name wrapper */
.remote-name-wrapper {
    display: inline-block;
}

/* ✅ REMOVED: Old .remote-icon styles (diganti dengan button) */

/* ✅ UNCHANGED: Email cell styles */
.email-cell {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.85rem;
    max-width: 200px;
    cursor: help;
}

.email-icon {
    font-size: 0.9rem;
    flex-shrink: 0;
}

.email-text {
    color: #2c3e50;
    font-weight: 500;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    transition: color 0.2s ease;
}

.email-cell:hover .email-text {
    color: #3498db;
}

.email-placeholder {
    color: #95a5a6;
    font-style: italic;
    font-size: 0.8rem;
}

/* ✅ UNCHANGED: Status styles */
.status {
    display: inline-block;
    padding: 3px 8px;
    border-radius: 4px;
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: uppercase;
    color: white;
}

.status.connected {
    background-color: #27ae60;
}

.status.disconnected {
    background-color: #e74c3c;
}

.status.pending {
    background-color: #f39c12;
}

/* ✅ UNCHANGED: Warning styles */
.warning {
    font-size: 0.75rem;
    color: #ff0000;
    margin-top: 0.25rem;
}

/* ✅ UNCHANGED: Storage info styles */
.storage-info {
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
}

.storage-info > span:first-child {
    font-size: 0.73rem;
    font-weight: 500;
}

.storage-row {
    display: flex;
    align-items: center;
    gap: 0.5rem;
}

.storage-bar {
    width: 150px;
    height: 4px;
    background-color: #ecf0f1;
    border-radius: 2px;
    overflow: hidden;
}

.bar-fill {
    height: 100%;
    background-color: #3498db;
    transition: width 0.3s ease;
}

.percentage {
    font-size: 0.75rem;
    color: #7f8c8d;
    font-weight: 500;
    min-width: 2rem;
}

/* ✅ UNCHANGED: Table cell styles */
.center {
    text-align: center;
}

.text-muted {
    color: #95a5a6;
    font-size: 0.85rem;
}

td {
    padding: 0.75rem;
    vertical-align: middle;
}

/* ✅ NEW: Responsive styles */
@media (max-width: 768px) {
    .remote-name-btn {
        font-size: 0.85rem;
        padding: 0.4rem 0.6rem;
    }
    
    .drive-icon {
        font-size: 1rem;
    }
    
    .email-cell {
        max-width: 150px;
    }
}
</style>