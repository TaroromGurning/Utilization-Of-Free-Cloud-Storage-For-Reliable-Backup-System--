import apiClient from './api' // Import instance Axios (sudah ada interceptor JWT)

export default {
  /**
   * Mengambil daftar status remote dari backend Golang.
   * Endpoint: GET /api/v1/monitoring/remotes
   * @returns {Promise<Array>} Daftar remote
   */
  async getRemoteStatus() {
    try {
      const response = await apiClient.get('/monitoring/remotes')
      return response.data
    } catch (error) {
      console.error("Error fetching remote status:", error)
      throw error
    }
  },

  /**
   * Mengambil semua riwayat log eksekusi.
   * Endpoint: GET /api/v1/monitoring/logs
   * @returns {Promise<Array>} Daftar logs
   */
  async getLogs() {
    try {
      const response = await apiClient.get('/monitoring/logs')
      return response.data
    } catch (error) {
      console.error("Error fetching logs:", error)
      throw error
    }
  },

  /**
   * Mengambil daftar Job terjadwal (Job Monitoring).
   * Endpoint: GET /api/v1/monitoring/jobs
   * @returns {Promise<Array>} Daftar job (termasuk NextRun)
   */
  async getScheduledJobs() {
    try {
      const response = await apiClient.get('/monitoring/jobs')
      return response.data
    } catch (error) {
      console.error("Error fetching scheduled jobs:", error)
      throw error
    }
  }
}