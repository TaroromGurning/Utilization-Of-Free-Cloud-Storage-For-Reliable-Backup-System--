// src/stores/setup.js atau src/stores/app.js

import { defineStore } from "pinia";
import { ref } from "vue";

export const useAppStore = defineStore('app', () => {
    // Status setup: 'LOADING', 'SETUP_NEEDED', 'SETUP_COMPLETE'
    const setupStatus = ref('LOADING'); 

    // Fungsi untuk memanggil API backend
    async function checkSetupStatus() {
        if (setupStatus.value !== 'LOADING') {
            return; // Hindari panggilan berulang jika sudah diketahui statusnya
        }
        
        try {
            const response = await fetch('/api/v1/setup/status');
            const data = await response.json();
            
            // Asumsi backend mengembalikan { "is_admin_registered": true/false }
            if (data.is_admin_registered === false) {
                setupStatus.value = 'SETUP_NEEDED';
            } else {
                setupStatus.value = 'SETUP_COMPLETE';
            }
        } catch (error) {
            console.error('Gagal memeriksa status setup:', error);
            // Fallback aman: jika gagal, asumsikan perlu setup
            setupStatus.value = 'SETUP_NEEDED'; 
        }
    }

    function setSetupComplete() {
        setupStatus.value = 'SETUP_COMPLETE';
    }

    return { setupStatus, checkSetupStatus, setSetupComplete };
});