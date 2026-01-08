package service

import (
	"encoding/json"
	"fmt"
	"gbackup-new/backend/internal/models"
	"gbackup-new/backend/internal/repository"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type MonitoringService interface {
	UpdateRemoteStatus(remoteName string) error
	GetRemoteStatusList() ([]models.Monitoring, error)
	GetRcloneConfiguredRemotes() ([]string, error)
	GetJobLogs() ([]models.Log, error)
	GetAllJobs() ([]models.ScheduledJob, error)
	DiscoverAndSaveRemote() error
	RunRemoteChecks() error
	StartMonitoringDaemon()
	SyncRemotesWithRclone() error
	ExtractEmailFromConfig(remoteName string) (string, error)
}

type monitoringServiceImpl struct {
	MonitorRepo repository.MonitoringRepository
	LogRepo     repository.LogRepository
	JobRepo     repository.JobRepository
}

const (
	intervalCek  = 5 * time.Minute
	intervalSync = 1 * time.Minute
)

func NewMonitoringService(mRepo repository.MonitoringRepository, lRepo repository.LogRepository, jRepo repository.JobRepository) MonitoringService {
	return &monitoringServiceImpl{
		MonitorRepo: mRepo,
		LogRepo:     lRepo,
		JobRepo:     jRepo,
	}
}

func (s *monitoringServiceImpl) UpdateRemoteStatus(remoteName string) error {
	fmt.Printf("[Monitoring] Mengecek remote: %s\n", remoteName)

	rcloneArgs := []string{"rclone", "about", remoteName + ":", "--json"}
	result := ExecuteCliJob(rcloneArgs)

	monitor := &models.Monitoring{
		RemoteName:    remoteName,
		LastCheckedAt: time.Now(),
	}

	// ============================================================
	// ✅ Hitung semua job (scheduled + manual) ke remote ini
	// EXCLUDE: restore one-shot jobs
	// ============================================================
	count, errJob := s.JobRepo.CountJobOnRemote(remoteName)
	if errJob != nil {
		fmt.Printf("⚠️ Gagal hitung job pada %s: %v\n", remoteName, errJob)
		monitor.ActiveJobCount = 0
	} else {
		monitor.ActiveJobCount = count
	}

	if !result.Success {
		fmt.Printf("❌ Rclone Error pada %s: %s\n", remoteName, result.ErrorMsg)
		monitor.StatusConnect = "DISCONNECTED"
		monitor.SystemMessage = result.ErrorMsg

		s.MonitorRepo.UpsertRemoteStatus(monitor)
		return fmt.Errorf("gagal terhubung ke %s: %s", remoteName, result.ErrorMsg)
	}

	var rcloneData struct {
		Total uint64 `json:"total"`
		Used  uint64 `json:"used"`
		Free  uint64 `json:"free"`
	}

	if err := json.Unmarshal([]byte(result.Output), &rcloneData); err != nil {
		fmt.Printf("❌ JSON Parse Error: %v\n", err)
		return fmt.Errorf("gagal parsing JSON: %v", err)
	}

	const BytesToGB = 1073741824.0

	monitor.StatusConnect = "CONNECTED"
	monitor.TotalStorageGB = float64(rcloneData.Total) / BytesToGB
	monitor.UsedStorageGB = float64(rcloneData.Used) / BytesToGB
	monitor.FreeStorageGB = float64(rcloneData.Free) / BytesToGB

	usedPercentage := (monitor.UsedStorageGB / monitor.TotalStorageGB) * 100
	const WarningThreshold = 85.0

	if usedPercentage >= WarningThreshold {
		msg := fmt.Sprintf("⚠️ Storage terisi %.1f%%. PERINGATAN!", usedPercentage)
		monitor.SystemMessage = msg
	} else {
		monitor.SystemMessage = ""
	}

	// 🆕 Extract email dari rclone.conf
	email, err := s.ExtractEmailFromConfig(remoteName)
	if err != nil {
		fmt.Printf("⚠️ Warning: Gagal extract email untuk %s: %v\n", remoteName, err)
		email = "" // Tetap proses, jangan error
	}
	monitor.OwnerEmail = email

	fmt.Printf("✅ %s: %.2f GB / %.2f GB (%.1f%%) | %d active jobs | Email: %s\n",
		remoteName, monitor.UsedStorageGB, monitor.TotalStorageGB, usedPercentage, monitor.ActiveJobCount, email)

	return s.MonitorRepo.UpsertRemoteStatus(monitor)
}

func (s *monitoringServiceImpl) GetRemoteStatusList() ([]models.Monitoring, error) {
	remotes, err := s.MonitorRepo.FindAllRemotes()
	if err != nil {
		return nil, err
	}

	// Hitung job count untuk setiap remote
	for i := range remotes {
		count, _ := s.JobRepo.CountJobOnRemote(remotes[i].RemoteName)
		remotes[i].ActiveJobCount = count
	}

	return remotes, nil
}

func (s *monitoringServiceImpl) GetRcloneConfiguredRemotes() ([]string, error) {
	result := ExecuteCliJob([]string{"rclone", "listremotes"})

	if !result.Success {
		return nil, fmt.Errorf("gagal mendapatkan daftar remote: %s", result.ErrorMsg)
	}

	remotes := strings.Split(result.Output, "\n")
	var cleanNames []string

	for _, remote := range remotes {
		name := strings.TrimSpace(remote)
		if len(name) > 0 && strings.HasSuffix(name, ":") {
			cleanNames = append(cleanNames, name[:len(name)-1])
		}
	}
	return cleanNames, nil
}

func (s *monitoringServiceImpl) GetJobLogs() ([]models.Log, error) {
	return s.LogRepo.FindAllLogs()
}

func (s *monitoringServiceImpl) SyncRemotesWithRclone() error {
	// ✅ CHANGED: Wider separator for better visibility
	fmt.Println("\n" + strings.Repeat("=", 70)) // Was 60
	// ✅ CHANGED: English + emoji for clarity
	fmt.Println("[SYNC] 🔄 Starting database sync with rclone.conf")
	fmt.Println(strings.Repeat("=", 70))

	// 1. Get remotes from rclone.conf
	// ✅ CHANGED: Numbered steps with emoji
	fmt.Println("[SYNC] 1️⃣ Fetching remotes from rclone.conf...")
	rcloneRemotes, err := s.GetRcloneConfiguredRemotes()
	if err != nil {
		fmt.Printf("[ERROR] Failed to get remotes from rclone: %v\n", err)
		return fmt.Errorf("failed to get remotes from rclone: %w", err)
	}
	// ✅ CHANGED: Better success message
	fmt.Printf("[SYNC] ✓ Found %d remote(s) in rclone.conf: %v\n", len(rcloneRemotes), rcloneRemotes)

	// 2. Get remotes from database
	fmt.Println("[SYNC] 2️⃣ Fetching remotes from database...")
	dbRemotes, err := s.MonitorRepo.GetAllRemoteNames()
	if err != nil {
		fmt.Printf("[ERROR] Failed to get remotes from DB: %v\n", err)
		return fmt.Errorf("failed to get remotes from DB: %w", err)
	}
	fmt.Printf("[SYNC] ✓ Found %d remote(s) in database: %v\n", len(dbRemotes), dbRemotes)

	// 3. Find remotes to delete
	fmt.Println("[SYNC] 3️⃣ Finding remotes to delete...")
	remoteToDelete := findMissingRemotes(dbRemotes, rcloneRemotes)

	if len(remoteToDelete) > 0 {
		// ✅ CHANGED: Better emoji and message
		fmt.Printf("[SYNC] 🗑️  Found %d remote(s) to delete: %v\n", len(remoteToDelete), remoteToDelete)

		for _, remoteName := range remoteToDelete {
			fmt.Printf("[SYNC] → Deleting '%s' from database...\n", remoteName)
			if err := s.MonitorRepo.DeleteRemoteByName(remoteName); err != nil {
				fmt.Printf("[ERROR] Failed to delete '%s': %v\n", remoteName, err)
			} else {
				fmt.Printf("[SYNC] ✓ Successfully deleted '%s'\n", remoteName)
			}
		}
		// ... deletion logic sama ...
	} else {
		fmt.Println("[SYNC] ✓ No remotes to delete")
	}

	// 4. Find new remotes to add
	fmt.Println("[SYNC] 4️⃣ Finding new remotes to add...")
	remoteToAdd := findMissingRemotes(rcloneRemotes, dbRemotes)

	if len(remoteToAdd) > 0 {
		// ✅ CHANGED: Better emoji
		fmt.Printf("[SYNC] ➕ Found %d new remote(s): %v\n", len(remoteToAdd), remoteToAdd)

		for _, remoteName := range remoteToAdd {
			fmt.Printf("[SYNC] → Adding '%s' to database...\n", remoteName)
			monitor := &models.Monitoring{
				RemoteName:    remoteName,
				StatusConnect: "PENDING", // ✅ CHANGED: PENDING instead of DISCONNECTED
				LastCheckedAt: time.Now(),
			}
			if err := s.MonitorRepo.UpsertRemoteStatus(monitor); err != nil {
				fmt.Printf("[ERROR] Failed to add '%s': %v\n", remoteName, err)
			} else {
				fmt.Printf("[SYNC] ✓ Successfully added '%s'\n", remoteName)

				// ✅ NEW: Immediately check status for new remote
				go func(name string) {
					time.Sleep(1 * time.Second) // Brief delay
					s.UpdateRemoteStatus(name)
				}(remoteName)
			}
		}
	} else {
		fmt.Println("[SYNC] ✓ No new remotes to add")
	}

	// ✅ NEW: Better summary message
	if len(remoteToDelete) == 0 && len(remoteToAdd) == 0 {
		fmt.Println("[SYNC] ✅ Database is already in sync with rclone.conf")
	} else {
		fmt.Printf("[SYNC] ✅ Sync completed: +%d added, -%d deleted\n", len(remoteToAdd), len(remoteToDelete))
	}

	fmt.Println(strings.Repeat("=", 70) + "\n")
	return nil
}

func (s *monitoringServiceImpl) DiscoverAndSaveRemote() error {
	fmt.Println("[Discovery] Menyelaraskan remote dari rclone.conf...")

	if err := s.SyncRemotesWithRclone(); err != nil {
		return fmt.Errorf("gagal sync remote: %w", err)
	}

	remotes, err := s.MonitorRepo.FindAllRemotes()
	if err != nil {
		return fmt.Errorf("gagal mengambil remote dari DB: %w", err)
	}

	if len(remotes) == 0 {
		fmt.Println("[Discovery] ⚠️ Tidak ada remote yang dikonfigurasi")
		return nil
	}

	fmt.Printf("[Discovery] Memulai update status untuk %d remote...\n", len(remotes))

	for _, remote := range remotes {
		go s.UpdateRemoteStatus(remote.RemoteName)
	}

	return nil
}

func (s *monitoringServiceImpl) StartMonitoringDaemon() {
	go func() {
		fmt.Printf("[Daemon] Monitoring Daemon aktif, pengecekan tiap %v\n", intervalCek)

		if err := s.RunRemoteChecks(); err != nil {
			fmt.Printf("⚠️ Initial check error: %v\n", err)
		}

		if err := s.SyncRemotesWithRclone(); err != nil {
			fmt.Printf("❌ ERROR: Gagal sync remote: %v\n", err)
			return
		}

		statusTicker := time.NewTicker(intervalCek)
		defer statusTicker.Stop()

		syncTicker := time.NewTicker(intervalSync)
		defer syncTicker.Stop()

		for {
			select {
			case <-statusTicker.C:
				fmt.Println("[Daemon] Menjalankan pengecekan status remote...")
				if err := s.RunRemoteChecks(); err != nil {
					fmt.Printf("⚠️ ERROR saat pengecekan remote: %v\n", err)
				}

			case <-syncTicker.C:
				fmt.Println("[Daemon] Menjalankan sinkronisasi remote dengan rclone.conf...")
				if err := s.SyncRemotesWithRclone(); err != nil {
					fmt.Printf("⚠️ ERROR saat sinkronisasi remote: %v\n", err)
				} else {
					if err := s.RunRemoteChecks(); err != nil {
						fmt.Printf("⚠️ ERROR saat pengecekan remote setelah sinkronisasi: %v\n", err)
					}
				}
			}
		}
	}()
}

func (s *monitoringServiceImpl) RunRemoteChecks() error {
	remotes, err := s.MonitorRepo.FindAllRemotes()
	if err != nil {
		return err
	}

	if len(remotes) == 0 {
		fmt.Println("[Daemon] ℹ️ No remotes found in database")
		return nil
	}

	fmt.Printf("[Daemon] 📡 Checking status for %d remote(s)...\n", len(remotes))

	for _, remote := range remotes {
		go s.UpdateRemoteStatus(remote.RemoteName)
	}
	return nil
}

func (s *monitoringServiceImpl) FindByRemoteName(remoteName string) (*models.Monitoring, error) {
	return s.MonitorRepo.FindRemoteByName(remoteName)
}

// ============================================================
// 🆕 ExtractEmailFromConfig: Extract email dari rclone.conf
// ============================================================

// ============================================================
// FIXED: ExtractEmailFromConfig - Handle multi-line tokens
// ============================================================
func (s *monitoringServiceImpl) ExtractEmailFromConfig(remoteName string) (string, error) {
	fmt.Printf("[Service] Extract email untuk remote: %s\n", remoteName)

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

	contentStr := string(content)

	// ============================================================
	// Parse: Find section dan extract full content
	// ============================================================
	sectionStart := strings.Index(contentStr, fmt.Sprintf("[%s]", remoteName))
	if sectionStart == -1 {
		return "", fmt.Errorf("remote %s tidak ditemukan di rclone.conf", remoteName)
	}

	// Find end of section (next [) atau end of file
	sectionEnd := strings.Index(contentStr[sectionStart+1:], "[")
	if sectionEnd == -1 {
		sectionEnd = len(contentStr)
	} else {
		sectionEnd = sectionEnd + sectionStart + 1
	}

	sectionContent := contentStr[sectionStart:sectionEnd]
	fmt.Printf("[DEBUG] Section found at position %d-%d\n", sectionStart, sectionEnd)

	// ============================================================
	// Extract token line (dimulai dari "token = ")
	// ============================================================
	tokenStartIdx := strings.Index(sectionContent, "token = ")
	if tokenStartIdx == -1 {
		return "", fmt.Errorf("token tidak ditemukan untuk remote %s", remoteName)
	}

	// Extract dari "token = " sampai end of line pertama dengan "}"
	tokenLineStart := tokenStartIdx + 8 // len("token = ")
	tokenContent := sectionContent[tokenLineStart:]

	// Find matching closing brace untuk JSON object
	braceCount := 0
	tokenEnd := 0
	inString := false
	escapeNext := false

	for i, ch := range tokenContent {
		if escapeNext {
			escapeNext = false
			continue
		}

		if ch == '\\' {
			escapeNext = true
			continue
		}

		if ch == '"' && !escapeNext {
			inString = !inString
			continue
		}

		if !inString {
			if ch == '{' {
				braceCount++
			} else if ch == '}' {
				braceCount--
				if braceCount == 0 {
					tokenEnd = i + 1
					break
				}
			}
		}
	}

	if tokenEnd == 0 {
		return "", fmt.Errorf("gagal find matching brace untuk token")
	}

	tokenJSON := strings.TrimSpace(tokenContent[:tokenEnd])
	fmt.Printf("[DEBUG] Token JSON extracted (first 100 chars): %s\n", tokenJSON[:min(100, len(tokenJSON))])

	// ============================================================
	// Parse JSON token untuk ambil access_token
	// ============================================================

	var tokenData struct {
		AccessToken string `json:"access_token"`
	}

	if err := json.Unmarshal([]byte(tokenJSON), &tokenData); err != nil {
		fmt.Printf("[DEBUG] JSON parse error: %v\n", err)
		return "", fmt.Errorf("gagal parse token JSON: %w", err)
	}

	if tokenData.AccessToken == "" {
		return "", fmt.Errorf("access_token kosong di token JSON")
	}

	fmt.Printf("[DEBUG] Access token extracted (first 30 chars): %s...\n", tokenData.AccessToken[:min(30, len(tokenData.AccessToken))])

	// ============================================================
	// Call Google Drive API untuk ambil email
	// ============================================================
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
func (s *monitoringServiceImpl) fetchEmailFromGoogleAPI(accessToken string) (string, error) {
	const maxRetries = 3
	const timeout = 5 * time.Minute

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
				time.Sleep(time.Duration(attempt) * time.Second)
				continue
			}
			return "", fmt.Errorf("gagal call Google Drive API setelah %d attempts: %w", maxRetries, err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("gagal baca response body: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("Google Drive API error (status %d): %s", resp.StatusCode, string(body))
		}

		var apiResp struct {
			User struct {
				EmailAddress string `json:"emailAddress"`
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

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func findMissingRemotes(sourceList, targetList []string) []string {
	missing := []string{}

	targetMap := make(map[string]bool)
	for _, item := range targetList {
		targetMap[item] = true
	}

	for _, item := range sourceList {
		if !targetMap[item] {
			missing = append(missing, item)
		}
	}

	return missing
}

func (s *monitoringServiceImpl) GetAllJobs() ([]models.ScheduledJob, error) {
	return s.JobRepo.FindAllJobs()
}
