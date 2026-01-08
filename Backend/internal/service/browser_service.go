package service

import (
	"fmt"
	"gbackup-new/backend/internal/models"
	"gbackup-new/backend/internal/repository"
	"os/exec"
	"strings"
)

type BrowserService interface {
	BrowseFiles(remoteName string, path string) (*models.BrowserResponse, error)
	GetFileInfo(remoteName string, filePath string) (*models.FileItem, error)
	GetAvailableRemotes() ([]map[string]string, error)
}

type browserServiceImpl struct {
	browserRepo repository.BrowserRepository
}

func NewBrowserService(browserRepo repository.BrowserRepository) BrowserService {
	return &browserServiceImpl{
		browserRepo: browserRepo,
	}
}

// ============================================
// ✅ BROWSE FILES
// ============================================
func (s *browserServiceImpl) BrowseFiles(remoteName string, path string) (*models.BrowserResponse, error) {
	files, err := s.browserRepo.ListFiles(remoteName, path)
	if err != nil {
		return nil, err
	}

	// Hitung total size
	totalSize := int64(0)
	for _, f := range files {
		if !f.IsDir {
			totalSize += f.Size
		}
	}

	response := &models.BrowserResponse{
		Path:      path,
		Files:     files,
		TotalSize: totalSize,
	}

	return response, nil
}

// ============================================
// ✅ GET FILE INFO
// ============================================
func (s *browserServiceImpl) GetFileInfo(remoteName string, filePath string) (*models.FileItem, error) {
	file, err := s.browserRepo.GetFileInfo(remoteName, filePath)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// ============================================
// ✅ GET AVAILABLE REMOTES (rclone listremotes)
// ============================================
func (s *browserServiceImpl) GetAvailableRemotes() ([]map[string]string, error) {

	cmd := exec.Command("rclone", "listremotes")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list remotes: %w", err)
	}

	remotesList := strings.Split(strings.TrimSpace(string(output)), "\n")

	remotes := []map[string]string{}
	for _, remote := range remotesList {
		remote = strings.TrimSuffix(strings.TrimSpace(remote), ":")
		if remote != "" {
			remotes = append(remotes, map[string]string{
				"name":        remote,
				"description": remote + " (Cloud Storage)",
			})
		}
	}

	return remotes, nil
}
