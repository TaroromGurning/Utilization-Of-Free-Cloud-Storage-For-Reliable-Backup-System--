import apiClient from './api' // Import instance Axios dengan JWT interceptor

export default {
  /**
   * Browse files/folders di cloud storage
   * Endpoint: GET /api/v1/browser/files
   * @param {string} remoteName - Nama remote (e.g., "Gdrive1")
   * @param {string} path - Path folder (e.g., "/backups" atau "/")
   * @returns {Promise<object>} BrowserResponse dengan list files
   */
  async browseFiles(remoteName, path = '/') {
    try {
      const response = await apiClient.get('/browser/files', {
        params: {
          remote: remoteName,
          path: path
        }
      })
      return response.data
    } catch (error) {
      console.error(`Error browsing files in ${remoteName}:${path}:`, error)
      throw error
    }
  },

  /**
   * Get info dari single file
   * Endpoint: GET /api/v1/browser/info
   * @param {string} remoteName - Nama remote
   * @param {string} filePath - Full path ke file
   * @returns {Promise<object>} FileItem object
   */
  async getFileInfo(remoteName, filePath) {
    try {
      const response = await apiClient.get('/browser/info', {
        params: {
          remote: remoteName,
          file: filePath
        }
      })
      return response.data
    } catch (error) {
      console.error(`Error getting file info for ${remoteName}:${filePath}:`, error)
      throw error
    }
  },

  /**
   * Get list remotes yang tersedia
   * (Optional - untuk future use)
   */
  async listRemotes() {
    try {
      const response = await apiClient.get('/browser/remotes')
      return response.data
    } catch (error) {
      console.error('Error listing remotes:', error)
      throw error
    }
  }
}