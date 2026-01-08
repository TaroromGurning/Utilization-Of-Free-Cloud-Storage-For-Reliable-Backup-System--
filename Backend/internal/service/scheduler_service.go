package service

import (
	"fmt" // Diperlukan untuk string join
	"time"

	// Sesuaikan path module

	"gbackup-new/backend/internal/repository" // Sesuaikan path module

	"github.com/robfig/cron/v3"
)

// JobMonitoringDTO: Struct untuk output JSON Job Monitoring
type JobMonitoringDTO struct {
	ID           uint   `json:"id"`
	JobName      string `json:"job_name"`
	Type         string `json:"type"`
	SourcePath   string `json:"source_path"`
	GdriveTarget string `json:"gdrive_target"`
	Mode         string `json:"mode"`
	LastRun      string `json:"last_run"`
	Status       string `json:"status"`
	NextRun      string `json:"next_run"`
	FullScript   string `json:"full_script"`
}

// Interface (Kontrak)
type SchedulerService interface {
	StartDaemon()
	RunScheduledJobs() error
	CalculateNextRun(schedule string, lastRun time.Time) time.Time
	GetScheduledJobsInfo() ([]JobMonitoringDTO, error)
	GetGeneratedScript(jobID uint) (string, error) // Untuk Pratinjau Script
	GetManualJob() ([]JobMonitoringDTO, error)
}

// Implementasi Struct
type schedulerServiceImpl struct {
	JobRepo     repository.JobRepository
	BackupSvc   BackupService // Dependency ke BackupService
	intervalCek time.Duration
}

// Constructor (Dependency Injection)
func NewSchedulerService(jRepo repository.JobRepository, bSvc BackupService) SchedulerService {
	return &schedulerServiceImpl{
		JobRepo:     jRepo,
		BackupSvc:   bSvc,
		intervalCek: 1 * time.Minute, // Daemon mengecek setiap 1 menit
	}
}

// ----------------------------------------------------
// LOGIKA UTAMA SCHEDULER (JOB DISPATCHER)
// ----------------------------------------------------

// CalculateNextRun: Menghitung waktu run job selanjutnya
func (s *schedulerServiceImpl) CalculateNextRun(schedule string, lastRun time.Time) time.Time {
	if schedule == "" {
		return time.Time{} // Return waktu nol jika cron kosong
	}
	parser := cron.NewParser(
		cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow,
	)
	sched, err := parser.Parse(schedule)
	if err != nil {
		fmt.Printf("❌ ERROR: Gagal parsing cron expression '%s': %v\n", schedule, err)
		return time.Time{}
	}

	baseTime := lastRun
	if lastRun.IsZero() {
		baseTime = time.Now()
	}
	return sched.Next(baseTime)
}

// RunScheduledJobs: Me-Run Job yang terjadwal
func (s *schedulerServiceImpl) RunScheduledJobs() error {
	jobs, err := s.JobRepo.FindAllActiveJobs() // Mengambil Job (CRON != NULL)
	if err != nil {
		return fmt.Errorf("gagal mengambil job aktif dari DB: %w", err)
	}

	now := time.Now()

	for _, job := range jobs {

		if job.ScheduleCron == "" {
			// Job Manual hanya dipicu oleh TriggerManualJob (user/API).
			// Abaikan sepenuhnya oleh loop scheduler otomatis.
			fmt.Printf("[SCHEDULER] Job %d (%s) dilewati karena Job Manual (schedule kosong).\n", job.ID, job.JobName)
			continue
		}

		// Tentukan Waktu Basis (Base Time)
		var baseTime time.Time
		if job.LastRun != nil {
			baseTime = *job.LastRun
		} else {
			baseTime = job.CreatedAt // WAJIB: Gunakan CreatedAt jika Job baru
		}

		nextRun := s.CalculateNextRun(job.ScheduleCron, baseTime)

		// Cek apakah waktu NextRun sudah tiba atau terlewat
		if nextRun.Before(now) || nextRun.Equal(now) {

			// Cek apakah Job sudah RUNNING (Locking)
			if job.StatusQueue == "RUNNING" {
				fmt.Printf("[SCHEDULER] Job %d (%s) dilewati karena sudah berjalan.\n", job.ID, job.JobName)
				continue
			}

			lockTime := time.Now()
			// Update Status menjadi RUNNING (Locking)
			// (Catatan: StatusQueue akan diubah menjadi COMPLETED/FAILED oleh handleJobCompletion)
			err := s.JobRepo.UpdateLastRunStatus(job.ID, lockTime, "RUNNING")

			if err != nil {
				fmt.Printf("Job %d (%s) Gagal di-lock: %v\n", job.ID, job.JobName, err)
				continue
			}

			fmt.Printf("[SCHEDULER] Dispatching Job %d (%s)\n", job.ID, job.JobName)

			// Panggil BackupService (yang akan meluncurkan Goroutine Eksekusi 3 Fase)
			s.BackupSvc.TriggerManualJob(job.ID)
		}
	}
	return nil
}

// StartDaemon: Meluncurkan Goroutine tak terbatas
func (s *schedulerServiceImpl) StartDaemon() {
	go func() {
		fmt.Printf("🚀 Scheduler Daemon Aktif, Pengecekan Job tiap %s\n", s.intervalCek)
		// Loop tak terbatas
		for {
			err := s.RunScheduledJobs()
			if err != nil {
				fmt.Printf("⚠️ Daemon Error: %v\n", err)
			}
			// Tunggu sebelum pengecekan berikutnya
			time.Sleep(s.intervalCek)
		}
	}()
}

// ----------------------------------------------------
// LOGIKA JOB MONITORING (DATA READ)
// ----------------------------------------------------

// GetScheduledJobsInfo: Mengambil data untuk UI Monitoring
func (s *schedulerServiceImpl) GetScheduledJobsInfo() ([]JobMonitoringDTO, error) {
	jobs, err := s.JobRepo.FindAllActiveJobs()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil job aktif dari DB: %w", err)
	}

	var output []JobMonitoringDTO

	for _, job := range jobs {
		// Tentukan Waktu Basis (Base Time)
		var baseTime time.Time
		lastRunStr := ""

		if job.LastRun != nil {
			baseTime = *job.LastRun
			lastRunStr = job.LastRun.Format("02-01-2006 15:04")
		} else {
			baseTime = job.CreatedAt
		}

		nextRunTime := s.CalculateNextRun(job.ScheduleCron, baseTime)
		mode := "Auto"

		fullScript, err := s.GetGeneratedScript(job.ID)
		if err != nil {
			fullScript = "error generating script" + err.Error()
		}

		// Menggabungkan Tipe dan Path
		jobTypeFormatted := fmt.Sprintf("%s: %s", job.RcloneMode, job.SourcePath)

		output = append(output, JobMonitoringDTO{
			ID:           job.ID,
			JobName:      job.JobName,
			Type:         jobTypeFormatted,
			SourcePath:   job.SourcePath,
			GdriveTarget: job.RemoteName,
			Mode:         mode,
			LastRun:      lastRunStr,
			Status:       job.StatusQueue,
			NextRun:      nextRunTime.Format("2006-01-02 15:04:05"),
			FullScript:   fullScript,
		})
	}
	return output, nil
}

// GetGeneratedScript: Mengambil data job dan menggabungkan script untuk preview
func (s *schedulerServiceImpl) GetGeneratedScript(jobID uint) (string, error) {
	job, err := s.JobRepo.FindJobByID(jobID)
	if err != nil {
		return "", fmt.Errorf("job tidak ditemukan: %w", err)
	}

	// Generate Rclone Command
	// (PENTING: Kita perlu memindahkan logic 'buildRcloneArgs' ke utility
	// agar bisa diakses oleh SchedulerService)

	// Simulasi command Rclone untuk preview
	rcloneCmd := fmt.Sprintf("rclone %s %s %s:%s",
		job.RcloneMode,
		job.SourcePath,
		job.RemoteName,
		job.DestinationPath)

	// 1. Header (Wajib untuk Bash)
	scriptHeader := "#!/bin/bash\nset -eo pipefail\n\n"

	// 2. Pre-Script
	preScript := fmt.Sprintf("# // --- PRE-SCRIPT ---\n%s\n", job.PreScript)

	// 3. Rclone Command
	rcloneCmdStr := fmt.Sprintf("\n# // --- RCLONE COMMAND ---\n%s\n", rcloneCmd)

	// 4. Post-Script
	postScript := fmt.Sprintf("\n# // --- POST-SCRIPT ---\n%s\n", job.PostScript)

	return scriptHeader + preScript + rcloneCmdStr + postScript, nil

}

func (s *schedulerServiceImpl) GetManualJob() ([]JobMonitoringDTO, error) {
	jobs, err := s.JobRepo.FindManualJob()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil job manual: %w", err)
	}

	var output []JobMonitoringDTO
	for _, job := range jobs {
		lastRunstr := ""
		if job.LastRun != nil {
			lastRunstr = job.LastRun.Format("02-01-2006 15:04")
		}
		mode := "manual"

		jobTypeFormatted := fmt.Sprintf("%s: %s", job.RcloneMode, job.SourcePath)

		output = append(output, JobMonitoringDTO{
			ID:           job.ID,
			JobName:      job.JobName,
			Type:         jobTypeFormatted,
			SourcePath:   job.SourcePath,
			GdriveTarget: job.RemoteName,
			Mode:         mode,
			LastRun:      lastRunstr,
			Status:       job.StatusQueue,
			NextRun:      "N/A",
			FullScript:   "N/A",
		})
	}
	return output, nil
}
