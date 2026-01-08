package repository

import (
	"encoding/json"
	"fmt"
	"gbackup-new/backend/internal/models"
	"os/exec"
	"path/filepath"
	"strings"
)

type BrowserRepository interface {
	ListFiles(remoteName string, path string) ([]models.FileItem, error)
	GetFileInfo(remoteName string, filePath string) (*models.FileItem, error)
}

type browserRepositoryImpl struct{}

func NewBrowserRepository() BrowserRepository {
	return &browserRepositoryImpl{}
}

// ============================================
// ✅ LIST FILES/FOLDERS (Repository Layer)
// ============================================
// Menggunakan rclone lsjson untuk mendapatkan list files
func (r *browserRepositoryImpl) ListFiles(remoteName string, path string) ([]models.FileItem, error) {
	// 1. Normalize path
	if path == "" || path == "." {
		path = "/"
	}
	// Pastikan path dimulai dengan /
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// 2. Build rclone command
	// Format: rclone lsjson remoteName:path
	rcloneRemotePath := fmt.Sprintf("%s:%s", remoteName, path)

	cmd := exec.Command("rclone", "lsjson", "--recursive=false", rcloneRemotePath)

	// 3. Execute command dan capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("rclone lsjson failed: %w - %s", err, string(output))
	}

	// 4. Parse JSON output dari rclone
	var rcloneItems []map[string]interface{}

	jsonStr := string(output)

	// Handle empty response
	if strings.TrimSpace(jsonStr) == "" || strings.TrimSpace(jsonStr) == "[]" {
		return []models.FileItem{}, nil
	}

	err = json.Unmarshal([]byte(jsonStr), &rcloneItems)
	if err != nil {
		return nil, fmt.Errorf("failed to parse rclone output: %w", err)
	}

	// 5. Convert rclone items ke FileItem model
	files := []models.FileItem{}

	for _, item := range rcloneItems {
		fileItem := models.FileItem{}

		// Extract Name
		if name, ok := item["Name"].(string); ok {
			fileItem.Name = name
		}

		// Extract IsDir
		if isDir, ok := item["IsDir"].(bool); ok {
			fileItem.IsDir = isDir
		}

		// Extract Size
		if size, ok := item["Size"].(float64); ok {
			fileItem.Size = int64(size)
		}

		// Extract ModTime (waktu modifikasi)
		if modTime, ok := item["ModTime"].(string); ok {
			fileItem.ModTime = modTime
		}

		// Build full path
		fileItem.Path = filepath.Join(path, fileItem.Name)

		// Normalize path untuk Windows/Linux compatibility
		fileItem.Path = strings.ReplaceAll(fileItem.Path, "\\", "/")

		// Extract MimeType jika ada
		if mimeType, ok := item["MimeType"].(string); ok {
			fileItem.MimeType = mimeType
		}

		files = append(files, fileItem)
	}

	return files, nil
}

// ============================================
// ✅ GET SINGLE FILE INFO
// ============================================
// Menggunakan rclone stat untuk mendapatkan info file individual
func (r *browserRepositoryImpl) GetFileInfo(remoteName string, filePath string) (*models.FileItem, error) {
	// 1. Normalize path
	if !strings.HasPrefix(filePath, "/") {
		filePath = "/" + filePath
	}

	// 2. Build rclone command
	// Format: rclone stat remoteName:filepath
	rcloneRemotePath := fmt.Sprintf("%s:%s", remoteName, filePath)

	cmd := exec.Command("rclone", "stat", rcloneRemotePath)

	// 3. Execute command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("rclone stat failed: %w - %s", err, string(output))
	}

	// 4. Parse rclone stat output
	// Output format dari rclone stat:
	// Name: filename
	// Size: 1024
	// Modify time: 2025-11-19T10:00:00Z
	// IsDir: false

	fileInfo := &models.FileItem{}
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse setiap line
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "Name":
			fileInfo.Name = value
		case "Size":
			var size int64
			fmt.Sscanf(value, "%d", &size)
			fileInfo.Size = size
		case "Modify time":
			fileInfo.ModTime = value
		case "IsDir":
			fileInfo.IsDir = value == "true"
		}
	}

	// Set path
	fileInfo.Path = filePath

	// Jika kosong, coba dengan lsjson untuk single file
	if fileInfo.Name == "" {
		return r.getFileInfoViaLsJson(remoteName, filePath)
	}

	return fileInfo, nil
}

// ============================================
// Helper: Get File Info via lsjson (fallback)
// ============================================
// Fallback jika rclone stat tidak berhasil
func (r *browserRepositoryImpl) getFileInfoViaLsJson(remoteName string, filePath string) (*models.FileItem, error) {
	// Extract parent path dan filename
	parentPath := filepath.Dir(filePath)
	if parentPath == "." {
		parentPath = "/"
	}
	fileName := filepath.Base(filePath)

	// List files di parent directory
	files, err := r.ListFiles(remoteName, parentPath)
	if err != nil {
		return nil, err
	}

	// Cari file yang sesuai
	for _, f := range files {
		if f.Name == fileName {
			return &f, nil
		}
	}

	return nil, fmt.Errorf("file not found: %s", filePath)
}

// ============================================
// Helper: Format File Size (Human Readable)
// ============================================
// Mengubah size dalam bytes menjadi format readable (KB, MB, GB)
func FormatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// ============================================
// Helper: Validate Remote Name
// ============================================
// Memastikan remote name valid
func ValidateRemoteName(remoteName string) error {
	if remoteName == "" {
		return fmt.Errorf("remote name cannot be empty")
	}

	// Check if remote exists via rclone listremotes
	cmd := exec.Command("rclone", "listremotes")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to list remotes: %w", err)
	}

	remotes := strings.Split(string(output), "\n")
	for _, remote := range remotes {
		remote = strings.TrimSpace(remote)
		if strings.HasPrefix(remote, remoteName) {
			return nil
		}
	}

	return fmt.Errorf("remote '%s' not found", remoteName)
}
