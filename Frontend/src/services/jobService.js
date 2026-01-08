import apiClient from './api' // Import instance Axios (sudah ada interceptor JWT)

export default {
  /**
   * Mengirim konfigurasi Job Backup baru (Auto atau Manual) ke backend.
   * Endpoint: POST /api/v1/jobs/new
   * @param {object} backupConfig - DTO (JSON) dari form Backup Config.
   */
  async createBackupJob(backupConfig) {
    try {
      const response = await apiClient.post('/jobs/new', backupConfig)
      return response.data
    } catch (error) {
      console.error("Error creating backup job:", error)
      throw error
    }
  },

  /**
   * Mengirim konfigurasi Job Restore baru ke backend.
   * Endpoint: POST /api/v1/jobs/restore
   * @param {object} restoreConfig - DTO (JSON) dari form Restore Config.
   */
  async createRestoreJob(restoreConfig) {
    try {
      const response = await apiClient.post('/jobs/restore', restoreConfig)
      return response.data
    } catch (error) {
      console.error("Error creating restore job:", error)
      throw error
    }
  },

  /**
   * Memicu Job (Auto atau Manual) yang sudah ada di DB untuk berjalan SEKARANG.
   * Endpoint: POST /api/v1/jobs/trigger/:id
   * @param {number} jobId - ID dari Job yang akan dijalankan.
   */
  async triggerManualJob(jobId) {
    try {
      const response = await apiClient.post(`/jobs/trigger/${jobId}`)
      return response.data
    } catch (error) {
      console.error(`Error triggering job ${jobId}:`, error)
      throw error
    }
  },

  /**
   * Mengambil pratinjau script lengkap untuk Job ID tertentu.
   * Endpoint: GET /api/v1/jobs/script/:id
   * @param {number} jobId - ID dari Job.
   * @returns {Promise<object>} Object berisi "script_preview".
   */
  async getManualJobs() {
    try {
      const response = await apiClient.get('/jobs/manual');
      return response.data;
    } catch (error) {
      console.error("Error fetching manual jobs:", error);
      throw error;
    }
  },

  async getJobScript(jobId) {
    try {
      const response = await apiClient.get(`/jobs/script/${jobId}`)
      return response.data
    } catch (error) {
      console.error(`Error fetching job script ${jobId}:`, error)
      throw error
    }
  },

  /**
   * Menjelajahi file di GDrive (Untuk UI Restore).
   * Endpoint: GET /api/v1/browser/list
   * @param {string} remoteName - Nama remote GDrive.
   * @param {string} path - Path folder yang akan dijelajahi.
   * @returns {Promise<Array>} Daftar file/folder.
   */
  async browseFiles(remoteName, path) {
    try {
      const response = await apiClient.get('/browser/list', {
        params: {
          remote: remoteName,
          path: path,
        },
      })
      return response.data
    } catch (error) {
      console.error(`Error browsing files at ${remoteName}:${path}:`, error)
      throw error
    }
  },

  /**
   * Delete job by ID
   * Endpoint: DELETE /api/v1/jobs/delete/:id
   * @param {number} jobId - ID dari Job yang akan dihapus
   */
  async deleteJob(jobId) {
    try {
      const response = await apiClient.delete(`/jobs/delete/${jobId}`);
      return response.data;
    } catch (error) {
      console.error(`Error deleting job ${jobId}:`, error);
      throw error;
    }
  },

  /**
   * ✅ Get single job by ID (untuk edit)
   * Endpoint: GET /api/v1/jobs/:id
   * Response format: { success: true, data: { id, job_name, ... } }
   * @param {number} jobId - ID dari Job
   * @returns {Promise<object>} Job data (sudah di-extract dari wrapper)
   */
  async getJobById(jobId) {
    try {
      const response = await apiClient.get(`/jobs/${jobId}`);
      console.log('✅ Raw response from backend:', response.data);
      
      // ✅ Backend returns: { success: true, data: { ... } }
      // Kita return seluruh response (termasuk wrapper) agar bisa dihandle di modal
      return response.data;
    } catch (error) {
      console.error(`Error fetching job ${jobId}:`, error);
      throw error;
    }
  },

  /**
   * ✅ Update job by ID (partial update)
   * Endpoint: PUT /api/v1/jobs/update/:id
   * @param {number} jobId - ID dari Job
   * @param {object} updates - Object berisi field yang akan diupdate
   * @returns {Promise<object>} Success message
   */
  async updateJob(jobId, updates) {
    try {
      console.log('✅ Sending update request:', {
        jobId,
        updates
      });
      
      // ✅ Kirim updates sebagai body (sesuai dengan backend handler)
      const response = await apiClient.put(`/jobs/update/${jobId}`, updates);
      
      console.log('✅ Update response:', response.data);
      return response.data;
    } catch (error) {
      console.error(`Error updating job ${jobId}:`, error);
      throw error;
    } 
  },

  async getAllJobs() {
    try {
      const response = await apiClient.get('/jobs/alljobs');
      return response.data;
    } catch (error) {
      console.error("Error fetching all jobs:", error);
      throw error;
    }
  }
}