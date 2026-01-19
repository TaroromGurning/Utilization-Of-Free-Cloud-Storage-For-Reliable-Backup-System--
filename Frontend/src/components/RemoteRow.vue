<template>
    <tr class="remote-row" :class="getRowClass(remote.status_connect)">
        <!-- Remote Name Column - Clickable untuk browse -->
        <td>
            <button 
                @click="openFileBrowser" 
                class="remote-name-btn"
                :title="`Browse files in ${remote.remote_name}`"
            >
                {{ remote.remote_name }}
            </button>
            <div v-if="remote.system_message" class="warning">
                {{ remote.system_message }}
            </div>
        </td>
        
        <!-- Email Column with Truncation and Tooltip -->
        <td>
            <span v-if="remote.email" class="email-text" :title="remote.email">
                {{ truncatedEmail }}
            </span>
            <span v-else class="email-placeholder">No email</span>
        </td>
        
        <!-- Status Column -->
        <td>
            <span class="status-badge" :class="getStatusClass(remote.status_connect)">
                {{ remote.status_connect || 'PENDING' }}
            </span>
        </td>

        <!-- Storage Column -->
        <td>
            <div class="storage-info">
                <span class="storage-text">{{ usedFormatted }} / {{ totalFormatted }} GB</span>
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
        
        <!-- Active Jobs Column -->
        <td class="center-text">{{ remote.active_job_count }}</td>
        
        <!-- Last Checked Column -->
        <td>
            <span class="timestamp">{{ formatLastChecked(remote.last_checked_at) }}</span>
        </td>
    </tr>

    <!-- File Browser Modal -->
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

// State untuk File Browser Modal
const showFileBrowser = ref(false);

// Function untuk open file browser
function openFileBrowser() {
    showFileBrowser.value = true;
}

// Function untuk handle file selection
function handleFileSelect(fileData) {
    console.log('File selected:', fileData);
    // Bisa tambah logic lain jika diperlukan
}

// Computed properties
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

// Truncate email for display
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

// Helper function for row class based on status
function getRowClass(status) {
    if (!status) return '';
    const s = status.toUpperCase();
    if (s === 'CONNECTED') return 'row-connected';
    if (s === 'DISCONNECTED') return 'row-disconnected';
    return '';
}

// Helper function for status class
function getStatusClass(status) {
    if (!status) return 'badge-pending';
    const s = status.toUpperCase();
    if (s === 'CONNECTED') return 'badge-connected';
    if (s === 'DISCONNECTED') return 'badge-disconnected';
    if (s === 'PENDING') return 'badge-pending';
    return 'badge-default';
}

// Format date function
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
/* ============================================ */
/* TABLE ROW */
/* ============================================ */
.remote-row {
    border-bottom: 1px solid #f0f0f0;
    transition: all 0.15s;
    cursor: default;
}

.remote-row:hover {
    background: #fafafa;
}

.remote-row.row-connected {
    border-left: 3px solid #22c55e;
}

.remote-row.row-disconnected {
    border-left: 3px solid #ef4444;
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
/* REMOTE NAME BUTTON - Minimal */
/* ============================================ */
.remote-name-btn {
    background: none;
    border: none;
    padding: 0;
    cursor: pointer;
    font-size: 0.9375rem;
    font-weight: 500;
    color: #3b82f6;
    text-decoration: underline;
    text-decoration-style: dotted;
    text-underline-offset: 3px;
    transition: color 0.2s;
}

.remote-name-btn:hover {
    color: #2563eb;
}

/* ============================================ */
/* EMAIL CELL - Minimal */
/* ============================================ */
.email-text {
    color: #666;
    font-family: monospace;
    font-size: 0.875rem;
}

.email-placeholder {
    color: #999;
    font-style: italic;
    font-size: 0.875rem;
}

/* ============================================ */
/* STATUS BADGE - Same as ManualJobRow */
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

.badge-connected {
    background: #d1fae5;
    color: #065f46;
}

.badge-disconnected {
    background: #fee2e2;
    color: #991b1b;
}

.badge-pending {
    background: #fef3c7;
    color: #92400e;
}

.badge-default {
    background: #f3f4f6;
    color: #374151;
}

/* ============================================ */
/* WARNING MESSAGE */
/* ============================================ */
.warning {
    font-size: 0.75rem;
    color: #ef4444;
    margin-top: 0.375rem;
    font-weight: 500;
}

/* ============================================ */
/* STORAGE INFO */
/* ============================================ */
.storage-info {
    display: flex;
    flex-direction: column;
    gap: 0.375rem;
}

.storage-text {
    font-size: 0.8125rem;
    font-weight: 500;
    color: #1a1a1a;
}

.storage-row {
    display: flex;
    align-items: center;
    gap: 0.5rem;
}

.storage-bar {
    width: 150px;
    height: 6px;
    background-color: #f0f0f0;
    border-radius: 3px;
    overflow: hidden;
}

.bar-fill {
    height: 100%;
    background: linear-gradient(90deg, #22c55e 0%, #16a34a 100%);
    transition: width 0.3s ease;
}

.percentage {
    font-size: 0.75rem;
    color: #666;
    font-weight: 600;
    min-width: 2.5rem;
    font-family: monospace;
}

/* ============================================ */
/* JOB COUNT - Minimal */
/* ============================================ */
.center-text {
    text-align: center;
    font-weight: 600;
    color: #1a1a1a;
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
/* RESPONSIVE */
/* ============================================ */
@media (max-width: 1024px) {
    td {
        padding: 0.75rem;
        font-size: 0.8125rem;
    }

    .remote-name-btn {
        font-size: 0.8125rem;
        padding: 0.25rem 0.5rem;
    }

    .email-cell {
        max-width: 180px;
    }

    .storage-bar {
        width: 120px;
    }
}

@media (max-width: 768px) {
    td {
        padding: 0.625rem 0.5rem;
    }

    .remote-name-btn {
        font-size: 0.875rem;
    }

    .email-text {
        font-size: 0.75rem;
    }

    .storage-text {
        font-size: 0.75rem;
    }

    .storage-bar {
        width: 100px;
        height: 4px;
    }

    .status-badge {
        padding: 0.25rem 0.5rem;
        font-size: 0.7rem;
    }

    .timestamp {
        font-size: 0.75rem;
    }
}

@media (max-width: 480px) {
    .storage-bar {
        width: 80px;
    }
}
</style>