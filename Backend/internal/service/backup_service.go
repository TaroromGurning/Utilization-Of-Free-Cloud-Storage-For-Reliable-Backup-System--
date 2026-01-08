package service

import (
	"encoding/json"
	"fmt"
	"gbackup-new/backend/internal/models"
	"gbackup-new/backend/internal/repository"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// BackupService interface (Kontrak)
type BackupService interface {
	CreateJobAndDispatch(job *models.ScheduledJob) error
	TriggerManualJob(jobID uint) error
	DeleteJob(JobId uint) error
	UpdateJob(jobID uint, updatedJob *models.ScheduledJob) error
	GetJobByID(jobID uint) (*models.ScheduledJob, error)
}

type backupServiceImpl struct {
	MonitorRepo repository.MonitoringRepository
	JobRepo     repository.JobRepository
	LogRepo     repository.LogRepository
	MonitorSvc  MonitoringService
}

type RcloneFileInfo struct {
	Name    string    `json:"Name"`    // Nama file/folder
	Size    int64     `json:"Size"`    // Ukuran dalam bytes
	ModTime time.Time `json:"ModTime"` // Waktu modifikasi
	IsDir   bool      `json:"IsDir"`   // True jika folder
}

func NewBackupService(
	jRepo repository.JobRepository,
	lRepo repository.LogRepository,
	mRepo repository.MonitoringRepository,
	mSvc MonitoringService,
) BackupService {
	return &backupServiceImpl{
		JobRepo:     jRepo,
		LogRepo:     lRepo,
		MonitorRepo: mRepo,
		MonitorSvc:  mSvc,
	}
}

const MinFreeGB = 1.0

// ----------------------------------------------------
// FUNGSI UTAMA ORKESTRASI & DISPATCH
// ----------------------------------------------------

// CreateJobAndDispatch: Menerima Job dari Controller, menyimpan, dan/atau memicu
func (s *backupServiceImpl) CreateJobAndDispatch(job *models.ScheduledJob) error {

	if job.OperationMode == "" {
		job.OperationMode = "BACKUP"
	}
	if job.RcloneMode == "" {
		job.RcloneMode = "copy"
	}
	if job.OperationMode == "RESTORE" {
		fmt.Printf("[DISPATCHER] 🔄 RESTORE Job: %s (One-Shot, TIDAK disimpan ke DB)\n", job.JobName)
		go s.executeJobLifecycle(*job)
		return nil
	}
	// 1. SELALU SIMPAN JOB KE DATABASE (sebagai Template)
	if err := s.JobRepo.Create(job); err != nil {
		return fmt.Errorf("gagal menyimpan job template: %w", err)
	}

	// 2. JALANKAN JIKA MANUAL
	if job.ScheduleCron == "" {
		fmt.Printf("[DISPATCHER] Job %s (Manual) disimpan (ID: %d) dan dipicu langsung.\n", job.JobName, job.ID)
		// Memulai eksekusi di background (Goroutine)
		go s.executeJobLifecycle(*job)
		return nil
		// Job Terjadwal (Auto Backup)
		// fmt.Printf("[DISPATCHER] Job %d (%s) disimpan untuk Scheduler.\n", job.ID, job.JobName)
	}
	fmt.Printf("[DISPATCHER] Job %d (%s) disimpan untuk Scheduler.\n", job.ID, job.JobName)
	return nil
}

// TriggerManualJob: Memicu Job yang sudah ada di DB
func (s *backupServiceImpl) TriggerManualJob(jobID uint) error {
	job, err := s.JobRepo.FindJobByID(jobID)
	if err != nil {
		return err
	}

	// Langsung eksekusi di background
	go s.executeJobLifecycle(*job)
	return nil
}

// ----------------------------------------------------
// FUNGSI EKSEKUSI 3 FASE (INTI)
// ----------------------------------------------------

// executeJobLifecycle: Menjalankan Pre-Script, Rclone, dan Post-Script
func (s *backupServiceImpl) executeJobLifecycle(job models.ScheduledJob) {
	fmt.Printf("[WORKER %d] Job %s: Memulai Eksekusi 3 Fase... RcloneMode: %s\n", job.ID, job.JobName, job.RcloneMode)
	// ⭐ HIGHLIGHT: Menampilkan mode di log
	fmt.Printf("[WORKER %d] Menjalankan Rclone (Mode: %s)...\n", job.ID, job.RcloneMode)
	// Set Status RUNNING (Locking)
	s.JobRepo.UpdateLastRunStatus(job.ID, time.Now(), "RUNNING")

	var finalResult RcloneResult
	var finalStatus string

	// 🚨 LANGKAH PENCEGAHAN (Hanya untuk Job Terjadwal) 🚨
	if job.OperationMode == "BACKUP" {
		const MinFreeGB = 1.0 // Batas aman yang dibutuhkan

		// 1. Dapatkan status remote
		monitor, err := s.MonitorRepo.FindRemoteByName(job.RemoteName)
		if err == nil && monitor != nil {

			// 2. Hitung/Estimasi ukuran sumber
			sourceSizeGB, _ := s.CalculateSourceSizeGB(job.SourcePath)
			requiredSpace := sourceSizeGB + MinFreeGB
			availableSpace := monitor.FreeStorageGB

			fmt.Printf("[WORKER %d] 📊 Source size: %.2f GB\n", job.ID, sourceSizeGB)
			fmt.Printf("[WORKER %d] 📊 Available space: %.2f GB\n", job.ID, availableSpace)
			fmt.Printf("[WORKER %d] 📊 Required space: %.2f GB\n", job.ID, requiredSpace)

			if availableSpace < requiredSpace {
				errorMsg := fmt.Sprintf(
					"⛔ STORAGE INSUFFICIENT: Perlu %.2f GB, tapi hanya tersedia %.2f GB di %s",
					requiredSpace, availableSpace, job.RemoteName,
				)
				fmt.Printf("[WORKER %d] %s\n", job.ID, errorMsg)
				s.handleJobCompletion(job, RcloneResult{Success: false, ErrorMsg: errorMsg}, "NOT_ENOUGH_SPACE")
				return
			}
			fmt.Printf("✅ [WORKER %d] Storage OK: Cukup untuk backup\n", job.ID)
		}
		// Jika gagal mendapatkan status monitor, biarkan Job berjalan (risiko kecil).
	}

	// --- FASE 1: PRE-SCRIPT ---
	if job.PreScript != "" {
		fmt.Printf("[WORKER %d] Menjalankan Pre-Script...\n", job.ID)
		hardenedPreScript := fmt.Sprintf("set -eo pipefail; \n%s", job.PreScript)
		preScriptArgs := []string{"bash", "-c", hardenedPreScript}

		result := ExecuteCliJob(preScriptArgs)
		if !result.Success {
			fmt.Printf("❌ [WORKER %d] Pre-Script GAGAL.\n", job.ID)
			finalResult = result
			finalStatus = "FAIL_PRE_SCRIPT"
			s.handleJobCompletion(job, finalResult, finalStatus)
			return
		}
	}

	// Timestamp
	// ============================================================
	// 🆕 FASE 1.5: TIMESTAMP & ROUND ROBIN
	// ============================================================
	var runtimeDestPath string

	if job.OperationMode == "BACKUP" {

		if job.RcloneMode == "copy" {
			fmt.Printf("⚠️ [WORKER %d] Skipping timestamping for non-backup job.\n", job.ID)
			timestamp := time.Now().Format("20060102_150405")

			sourceInfo, err := os.Stat(job.SourcePath)
			if err != nil {
				errorMsg := fmt.Sprintf("Failed to stat source path: %v", err)
				s.handleJobCompletion(job, RcloneResult{Success: false, ErrorMsg: errorMsg}, "FAIL_SOURCE_CHECK")
				return
			}

			var newDestinationName string

			var isSourceDir = sourceInfo.IsDir()

			if isSourceDir {
				// Folder → hasil tetap folder
				folderName := filepath.Base(job.SourcePath)
				newDestinationName = fmt.Sprintf("%s_%s", folderName, timestamp)
			} else {
				// FILE → hasil tetap file
				fileName := filepath.Base(job.SourcePath)
				ext := filepath.Ext(fileName)
				nameWithoutExt := strings.TrimSuffix(fileName, ext)
				newDestinationName = fmt.Sprintf("%s_%s%s", nameWithoutExt, timestamp, ext)
			}

			originalDestPath := job.DestinationPath
			runtimeDestPath = filepath.Join(originalDestPath, newDestinationName)

			fmt.Printf("[WORKER %d] 🎯 Runtime destination: %s:%s\n", job.ID, job.RemoteName, runtimeDestPath)
			fmt.Printf("[WORKER %d] 💾 DB destination (unchanged): %s\n", job.ID, job.DestinationPath)

			// Round Robin Cleanup
			fmt.Printf("[WORKER %d] 🔄 Checking for old backups...\n", job.ID)
			if err := s.CleanupOldBackups(job.RemoteName, originalDestPath, job.MaxRetention); err != nil {
				fmt.Printf("⚠️ [WORKER %d] Cleanup warning: %v\n", job.ID, err)
			}
		} else if job.RcloneMode == "sync" {
			runtimeDestPath = job.DestinationPath
		} else {
			runtimeDestPath = job.DestinationPath
		}
	} else {
		runtimeDestPath = job.DestinationPath
	}

	// ============================================================

	// --- FASE 2: RCLONE EXECUTION ---
	fmt.Printf("[WORKER %d] Menjalankan Rclone...\n", job.ID)
	rcloneArgs := s.buildRcloneArgs(job, runtimeDestPath)
	resultRclone := ExecuteCliJob(rcloneArgs)

	if !resultRclone.Success {
		fmt.Printf("❌ [WORKER %d] Rclone GAGAL.\n", job.ID)
		finalResult = resultRclone
		finalStatus = "FAIL_RCLONE"
		s.handleJobCompletion(job, finalResult, finalStatus)
		return
	}

	lines := strings.Split(resultRclone.Output, "\n")
	var transferStatus string
	for _, line := range lines {
		if strings.Contains(line, "Transferred:") {
			transferStatus = strings.TrimSpace(line)
		}
	}

	if transferStatus != "" {
		fmt.Printf("📊 [WORKER %d] Stats: %s\n", job.ID, transferStatus)
		// menyimpan  Log
		resultRclone.Output = fmt.Sprintf("%s\n\n%s", transferStatus, resultRclone.Output)
	}

	// --- FASE 3: POST-SCRIPT ---
	if job.PostScript != "" {
		fmt.Printf("[WORKER %d] Menjalankan Post-Script...\n", job.ID)
		hardenedPostScript := fmt.Sprintf("set -eo pipefail; \n%s", job.PostScript)
		postScriptArgs := []string{"bash", "-c", hardenedPostScript}

		resultPost := ExecuteCliJob(postScriptArgs)
		if !resultPost.Success {
			fmt.Printf("❌ [WORKER %d] Post-Script GAGAL.\n", job.ID)
			finalResult = resultPost
			finalStatus = "FAIL_POST_SCRIPT"
			s.handleJobCompletion(job, finalResult, finalStatus)
			return // Hentikan eksekusi
		}
	}

	// --- FASE 4: SUKSES ---
	fmt.Printf("✅ [WORKER %d] Job Selesai.\n", job.ID)
	finalResult = resultRclone // Log output Rclone jika sukses
	finalStatus = "SUCCESS"
	s.handleJobCompletion(job, finalResult, finalStatus)
}

// ----------------------------------------------------
// FUNGSI HELPER (COMMAND GENERATION & LOGGING)
// ----------------------------------------------------

// buildRcloneArgs: Menyusun command Rclone
func (s *backupServiceImpl) buildRcloneArgs(job models.ScheduledJob, runtimeDestPath string) []string {
	isRestore := job.OperationMode == "RESTORE"
	command := strings.ToLower(job.RcloneMode)

	// Tentukan apakah sumber adalah file atau folder
	sourceInfo, err := os.Stat(job.SourcePath)
	isSourceDir := false
	if err == nil {
		isSourceDir = sourceInfo.IsDir()
	}

	var SourcePath, Destination string

	if isRestore {
		SourcePath = fmt.Sprintf("%s:%s", job.RemoteName, job.SourcePath)
		Destination = job.DestinationPath
		command = "copy"
	} else {
		SourcePath = job.SourcePath
		Destination = fmt.Sprintf("%s:%s", job.RemoteName, runtimeDestPath)

		switch command {
		case "sync":
			command = "sync"
		case "copy", "":
			if !isSourceDir {
				command = "copyto"
			} else {
				command = "copy"
			}
		default:
			command = "copy"
		}
	}

	args := []string{
		"rclone",
		command,
		SourcePath,
		Destination,
		"--checksum",
		"--no-traverse",
		"--progress",    //Tampilkan progress dengan stats
		"--stats", "5s", //Print stats setiap 5 detik
		"--stats-log-level", "INFO", //Change to INFO level
		"--human-readable",
	}

	if command == "sync" {
		args = append(args, "--delete-during")
		fmt.Printf("[buildRcloneArgs] SYNC mode detected: adding --delete-during flag\n")
	}

	return args
}

// handleJobCompletion: Logika Logging dan Final Status Update
func (s *backupServiceImpl) handleJobCompletion(job models.ScheduledJob, result RcloneResult, status string) {
	LogMutex.Lock()
	defer LogMutex.Unlock()

	// --- 1. LOGIKA PESAN LOG ---
	// Jika Sukses: Ambil output bersih (Size | Time)
	// Jika Gagal: Ambil output + Error message
	logMessage := result.Output
	if status != "SUCCESS" {
		logMessage = result.ErrorMsg
	}

	// --- 2. SIMPAN KE TABEL LOGS (Detail Status) ---
	// Di sini kita simpan status spesifik (misal: FAIL_RCLONE)
	newLog := &models.Log{
		JobID:            nil,
		JobName:          job.JobName,
		SourcePath:       job.SourcePath,
		Status:           status, // ✅ Simpan FAIL_RCLONE/FAIL_PRE_SCRIPT disini
		Message:          logMessage,
		DurationSec:      int(result.Duration.Seconds()),
		TransferredBytes: result.TransferredBytes,
		Timestamp:        time.Now(),
	}

	if job.ID != 0 {
		newLog.JobID = &job.ID
	}

	fmt.Printf("[LOG DEBUG] Saving ID: %d | Status: %s | Bytes: %d\n", job.ID, status, result.TransferredBytes)
	s.LogRepo.CreateLog(newLog)

	if job.ID != 0 {
		var dbStatus string

		if status == "SUCCESS" {
			dbStatus = "COMPLETED"
		} else {
			dbStatus = "FAILED"
		}

		s.JobRepo.UpdateLastRunStatus(job.ID, time.Now(), dbStatus)
	}

	// --- 4. TERMINAL LOG SUMMARY ---
	if status == "SUCCESS" {
		// Parse stats hanya untuk print terminal yang cantik (opsional)
		stats := parseRcloneStats(result.Output)
		fmt.Printf("✅ [COMPLETE] Job %d: Transferred %.2f GB in %d seconds (Speed: %s)\n",
			job.ID,
			job.RcloneMode,
			float64(result.TransferredBytes)/1073741824.0,
			int(result.Duration.Seconds()),
			stats.Speed,
		)
	} else {
		fmt.Printf("❌ [FAILED] Job %d: %s\n", job.ID, status)
	}
}

func (s *backupServiceImpl) CalculateSourceSizeGB(path string) (float64, error) {
	var totalSize int64

	// ============================================================
	// STEP 1: Check apakah path (file/folder) exists
	// ============================================================
	fileInfo, err := os.Stat(path)
	if err != nil {
		return 0, fmt.Errorf("path tidak ditemukan: %w", err)
	}

	// ============================================================
	// STEP 2: Jika file single, return ukurannya saja
	// ============================================================
	if !fileInfo.IsDir() {
		// Path adalah FILE tunggal (bukan folder)
		const BytesToGB = 1073741824.0
		sizeGB := float64(fileInfo.Size()) / BytesToGB
		fmt.Printf("✅ Single file size calculated: %.2f GB (%d bytes)\n", sizeGB, fileInfo.Size())
		return sizeGB, nil
	}

	err = filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		// Error handling: jika ada error read file/folder, skip saja
		if err != nil {
			fmt.Printf("⚠️ Skip item %s (error: %v)\n", filePath, err)
			return nil // return nil = lanjut ke file/folder berikutnya
		}

		if !info.IsDir() {
			// Item adalah FILE (bukan folder)
			totalSize += info.Size() // Tambah size file ke counter
		}

		return nil // return nil = lanjut walk ke file/folder berikutnya
	})

	// Error handling jika walk gagal
	if err != nil && err != filepath.SkipDir {
		return 0, fmt.Errorf("gagal walk directory %s: %w", path, err)
	}

	// ============================================================
	// STEP 4: Convert total bytes to GB
	// ============================================================
	const BytesToGB = 1073741824.0 // 1 GB = 1024^3 bytes
	sizeGB := float64(totalSize) / BytesToGB

	fmt.Printf("✅ Total size calculated: %.2f GB (%d bytes)\n", sizeGB, totalSize)
	return sizeGB, nil
}

func (s *backupServiceImpl) DeleteJob(JobID uint) error {
	fmt.Printf("[AUDIT] User meminta penghapusan Job ID: %d\n", JobID)

	// Panggil Repository untuk menghapus
	if err := s.JobRepo.DeleteJob(JobID); err != nil {
		return fmt.Errorf("gagal menghapus job ID %d: %w", JobID, err)
	}
	return nil
}

func (s *backupServiceImpl) GetJobByID(jobID uint) (*models.ScheduledJob, error) {
	job, err := s.JobRepo.FindJobByID(jobID)
	if err != nil {
		return nil, err
	}
	return job, nil
}
func (s *backupServiceImpl) UpdateJob(jobID uint, updatedJob *models.ScheduledJob) error {
	fmt.Printf("[UPDATE] Memperbarui Job ID: %d\n", jobID)

	// 1. Cek apakah job exist
	_, err := s.JobRepo.FindJobByID(jobID)
	if err != nil {
		return fmt.Errorf("job tidak ditemukan: %w", err)
	}

	// 2. ✅ Build update map (hanya field yang ada)
	updates := make(map[string]interface{})

	if updatedJob.JobName != "" {
		updates["job_name"] = updatedJob.JobName
	}
	if updatedJob.OperationMode != "" {
		updates["operation_mode"] = updatedJob.OperationMode
	}
	if updatedJob.RcloneMode != "" {
		updates["rclone_mode"] = updatedJob.RcloneMode
	}
	if updatedJob.SourcePath != "" {
		updates["source_path"] = updatedJob.SourcePath
	}
	if updatedJob.DestinationPath != "" {
		updates["destination_path"] = updatedJob.DestinationPath
	}
	if updatedJob.RemoteName != "" {
		updates["remote_name"] = updatedJob.RemoteName
	}

	// ✅ Allow empty string untuk script (untuk clear script)
	updates["pre_script"] = updatedJob.PreScript
	updates["post_script"] = updatedJob.PostScript

	// ✅ Schedule cron bisa kosong (untuk ubah jadi manual job)
	updates["schedule_cron"] = updatedJob.ScheduleCron

	updates["updated_at"] = time.Now()

	if updatedJob.MaxRetention > 0 {
		if updatedJob.MaxRetention > 100 {
			return fmt.Errorf("max retention tidak boleh lebih dari 100")
		}
		updates["max_retention"] = updatedJob.MaxRetention
	}

	// 3. ✅ Validate minimal 1 field (exclude updated_at)
	if len(updates) <= 1 {
		return fmt.Errorf("tidak ada field yang diubah")
	}

	// 4. Update ke database
	if err := s.JobRepo.UpdateJob(jobID, updates); err != nil {
		return fmt.Errorf("gagal update job: %w", err)
	}

	fmt.Printf("[UPDATE] Job %d berhasil diperbarui (%d fields)\n", jobID, len(updates)-1)
	return nil
}

// CleanupOldBackups - UPDATED untuk menerima maxRetention sebagai parameter
func (s *backupServiceImpl) CleanupOldBackups(remoteName, destinationPath string, maxRetention int) error {
	// Validate maxRetention
	if maxRetention < 1 {
		maxRetention = 10
		fmt.Printf("[Round Robin] Invalid maxRetention (%d), using default: %d\n", maxRetention, maxRetention)
	}

	fmt.Printf("[Round Robin] Checking backups in %s:%s (Max: %d)...\n",
		remoteName, destinationPath, maxRetention)

	// List file dari remote menggunakan rclone lsjson
	listCmd := fmt.Sprintf("rclone lsjson %s:%s", remoteName, destinationPath)
	output, err := exec.Command("bash", "-c", listCmd).Output()
	if err != nil {
		return fmt.Errorf("failed to list remote files: %w", err)
	}

	var files []RcloneFileInfo
	if err := json.Unmarshal(output, &files); err != nil {
		return fmt.Errorf("failed to parse rclone output: %w", err)
	}

	backupItems := files
	currentCount := len(backupItems)
	fmt.Printf("[Round Robin] Found %d backup items (limit: %d)\n", currentCount, maxRetention)

	// Jika tidak perlu cleanup
	if currentCount < maxRetention {
		fmt.Printf("[Round Robin] No cleanup needed (%d/%d)\n", currentCount, maxRetention)
		return nil
	}

	// Sort berdasarkan waktu modifikasi (ascending = oldest first)
	sort.Slice(backupItems, func(i, j int) bool {
		return backupItems[i].ModTime.Before(backupItems[j].ModTime)
	})

	// Hitung berapa yang perlu dihapus
	// Jika ada 12 backup dan maxRetention=10, hapus 12-10+1 = 3
	itemsToDelete := currentCount - maxRetention + 1
	fmt.Printf("[Round Robin] Deleting %d oldest backup(s)...\n", itemsToDelete)

	var deletedItems []string
	for i := 0; i < itemsToDelete && i < len(backupItems); i++ {
		itemToDelete := backupItems[i]
		fullPath := fmt.Sprintf("%s:%s/%s", remoteName, destinationPath, itemToDelete.Name)

		var deleteCmd string
		if itemToDelete.IsDir {
			deleteCmd = fmt.Sprintf("rclone purge %s", fullPath)
		} else {
			deleteCmd = fmt.Sprintf("rclone delete %s", fullPath)
		}

		// Execute delete command
		if err := exec.Command("bash", "-c", deleteCmd).Run(); err != nil {
			fmt.Printf("⚠️  [Round Robin] Failed to delete %s: %v\n", itemToDelete.Name, err)
			continue
		}

		deletedItems = append(deletedItems, itemToDelete.Name)

		// Format output log
		itemType := "file"
		sizeStr := fmt.Sprintf("%.2f MB", float64(itemToDelete.Size)/(1024*1024))
		if itemToDelete.IsDir {
			itemType = "folder"
			sizeStr = "folder"
		}

		fmt.Printf("🗑️  [Round Robin] Deleted %s: %s (date: %s, size: %s)\n",
			itemType,
			itemToDelete.Name,
			itemToDelete.ModTime.Format("2006-01-02 15:04:05"),
			sizeStr,
		)
	}

	fmt.Printf("✅ [Round Robin] Cleanup complete. Deleted %d item(s). Space available for new backup.\n",
		len(deletedItems))
	return nil
}
