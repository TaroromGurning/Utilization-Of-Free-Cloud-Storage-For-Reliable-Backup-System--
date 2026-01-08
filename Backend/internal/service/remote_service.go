package service

import (
	"encoding/json"
	"fmt"
	"gbackup-new/backend/internal/repository"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

// RemoteService - Extract email dari rclone.conf
type RemoteService interface {
	ExtractEmailFromConfig(remoteName string) (string, error)
}

type remoteServiceImpl struct {
	MonitorRepo repository.MonitoringRepository
}

func NewRemoteService(mRepo repository.MonitoringRepository) RemoteService {
	return &remoteServiceImpl{
		MonitorRepo: mRepo,
	}
}

// ============================================================
// ExtractEmailFromConfig: Parse email dari rclone.conf
// ============================================================
func (s *remoteServiceImpl) ExtractEmailFromConfig(remoteName string) (string, error) {
	fmt.Printf("[RemoteService] Extract email untuk remote: %s\n", remoteName)

	// Cari rclone.conf di home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("gagal ambil home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".config", "rclone", "rclone.conf")

	// Fallback untuk Windows
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configPath = filepath.Join(homeDir, "AppData", "Roaming", "rclone", "rclone.conf")
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("gagal baca rclone.conf: %w", err)
	}

	// Parse token dari config
	// Format: [remoteName]...token = {"access_token":"xxx",...}
	pattern := fmt.Sprintf(`\[%s\](.*?)(?:\[|$)`, remoteName)
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(string(content))

	if len(matches) < 2 {
		return "", fmt.Errorf("remote %s tidak ditemukan di rclone.conf", remoteName)
	}

	// Extract JSON token
	tokenPattern := regexp.MustCompile(`token\s*=\s*({.*?})`)
	tokenMatches := tokenPattern.FindStringSubmatch(matches[1])

	if len(tokenMatches) < 2 {
		return "", fmt.Errorf("token tidak ditemukan untuk remote %s", remoteName)
	}

	// Parse JSON token untuk ambil access_token
	var tokenData struct {
		AccessToken string `json:"access_token"`
	}

	if err := json.Unmarshal([]byte(tokenMatches[1]), &tokenData); err != nil {
		return "", fmt.Errorf("gagal parse token JSON: %w", err)
	}

	// Call Google Drive API untuk ambil email
	email, err := s.fetchEmailFromGoogleAPI(tokenData.AccessToken)
	if err != nil {
		fmt.Printf("⚠️ Warning: Gagal fetch email dari API: %v\n", err)
		return "", err
	}

	fmt.Printf("✅ Email ditemukan untuk %s: %s\n", remoteName, email)
	return email, nil
}

// ============================================================
// fetchEmailFromGoogleAPI: Call Google Drive API untuk ambil email
// ============================================================
func (s *remoteServiceImpl) fetchEmailFromGoogleAPI(accessToken string) (string, error) {
	const maxRetries = 3
	const timeout = 10 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequest("GET",
			"https://www.googleapis.com/drive/v3/about?fields=user(emailAddress,displayName)",
			nil)
		if err != nil {
			return "", fmt.Errorf("gagal create request: %w", err)
		}

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

		client := &http.Client{Timeout: timeout}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("⚠️ Attempt %d gagal: %v\n", attempt, err)
			if attempt < maxRetries {
				time.Sleep(time.Duration(attempt) * time.Second) // Exponential backoff
				continue
			}
			return "", fmt.Errorf("gagal call Google Drive API setelah %d attempts: %w", maxRetries, err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("gagal baca response body: %w", err)
		}

		// Handle error responses
		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("google grive api error (status %d): %s", resp.StatusCode, string(body))
		}

		// Parse response
		var apiResp struct {
			User struct {
				EmailAddress string `json:"emailAddress"`
				DisplayName  string `json:"displayName"`
			} `json:"user"`
		}

		if err := json.Unmarshal(body, &apiResp); err != nil {
			return "", fmt.Errorf("gagal parse API response: %w", err)
		}

		email := apiResp.User.EmailAddress

		if email == "" {
			return "", fmt.Errorf("email kosong dari API response")
		}

		return email, nil
	}

	return "", fmt.Errorf("failed to fetch email after retries")
}
