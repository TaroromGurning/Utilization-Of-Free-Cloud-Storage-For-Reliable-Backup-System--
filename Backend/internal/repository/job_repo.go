package repository

import (
	"errors"
	"fmt"
	"gbackup-new/backend/internal/models" // Sesuaikan path module
	"time"

	"gorm.io/gorm"
)

// JobRepository mendefinisikan kontrak untuk operasi Job
type JobRepository interface {
	Create(job *models.ScheduledJob) error
	FindJobByID(jobID uint) (*models.ScheduledJob, error)
	FindAllActiveJobs() ([]models.ScheduledJob, error)
	FindManualJob() ([]models.ScheduledJob, error)
	UpdateLastRunStatus(jobID uint, lastRunTime time.Time, status string) error
	UpdateJobActivity(JobID uint, isActive bool) error
	CountJobOnRemote(remoteName string) (int64, error)
	DeleteJob(JobID uint) error
	UpdateJob(jobID uint, updates map[string]interface{}) error
	FindAllJobs() ([]models.ScheduledJob, error)
}

type jobRepositoryImpl struct {
	DB *gorm.DB
}

func NewJobRepository(db *gorm.DB) JobRepository {
	return &jobRepositoryImpl{DB: db}
}

func (r *jobRepositoryImpl) CountJobOnRemote(remoteName string) (int64, error) {
	var count int64

	err := r.DB.Model(&models.ScheduledJob{}).
		Where("remote_name = ?", remoteName).
		Where("operation_mode != ?", "RESTORE"). // ✅ Exclude restore one-shot
		Count(&count).Error

	if err != nil {
		fmt.Printf("[ERROR] Gagal menghitung job untuk remote '%s': %v\n", remoteName, err)
		return 0, fmt.Errorf("gagal menghitung job untuk remote %s: %w", remoteName, err)
	}

	fmt.Printf("[DEBUG] Remote '%s': %d active jobs\n", remoteName, count)
	return count, nil
}

// Create: Menyimpan template Job baru
func (r *jobRepositoryImpl) Create(job *models.ScheduledJob) error {
	result := r.DB.Create(job)
	return result.Error
}

// FindJobByID: Mengambil satu job berdasarkan ID (untuk preview script / trigger manual)
func (r *jobRepositoryImpl) FindJobByID(jobID uint) (*models.ScheduledJob, error) {
	var job models.ScheduledJob
	result := r.DB.First(&job, jobID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("job ID %d tidak ditemukan", jobID)
		}
		return nil, result.Error
	}
	return &job, nil
}

// FindAllActiveJobs: Mengambil job yang siap dieksekusi oleh Scheduler
func (r *jobRepositoryImpl) FindAllActiveJobs() ([]models.ScheduledJob, error) {
	var jobs []models.ScheduledJob
	// Ambil Job yang aktif dan BUKAN Job Manual (karena Job Manual tidak punya cron)
	result := r.DB.Where("schedule_cron IS NOT NULL AND schedule_cron != ?", "").Find(&jobs)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, result.Error
	}
	return jobs, nil
}

// UpdateLastRunStatus: Mengupdate status (PENDING, RUNNING, COMPLETED, FAILED_PRE...)
func (r *jobRepositoryImpl) UpdateLastRunStatus(jobID uint, lastRunTime time.Time, status string) error {
	result := r.DB.Model(&models.ScheduledJob{}).
		Where("id = ?", jobID).
		Updates(map[string]interface{}{
			"last_run_at":  lastRunTime,
			"status_queue": status,
		})
	return result.Error
}

func (r *jobRepositoryImpl) UpdateJobActivity(jobID uint, isActive bool) error {
	result := r.DB.Model(&models.ScheduledJob{}).
		Where("id = ?", jobID).
		Update("is_active", isActive)
	return result.Error
}

func (r *jobRepositoryImpl) FindManualJob() ([]models.ScheduledJob, error) {
	var jobs []models.ScheduledJob
	result := r.DB.Where("schedule_cron IS NULL OR schedule_cron = ?", "").Find(&jobs)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, result.Error
	}
	return jobs, nil
}

func (r *jobRepositoryImpl) DeleteJob(JobID uint) error {
	result := r.DB.Delete(&models.ScheduledJob{}, JobID)
	if result != nil {
		return result.Error
	}
	return nil
}

// repository/job_repository_impl.go

func (r *jobRepositoryImpl) UpdateJob(jobID uint, updates map[string]interface{}) error {
	// ✅ Validate job exists
	var job models.ScheduledJob
	if err := r.DB.First(&job, jobID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("job ID %d tidak ditemukan", jobID)
		}
		return err
	}

	// ✅ Update dengan map
	result := r.DB.Model(&models.ScheduledJob{}).
		Where("id = ?", jobID).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("tidak ada perubahan pada job ID %d", jobID)
	}

	return nil
}

func (r *jobRepositoryImpl) FindAllJobs() ([]models.ScheduledJob, error) {
	var jobs []models.ScheduledJob
	err := r.DB.Where("operation_mode != ?", "RESTORE").
		Find(&jobs).Error
	return jobs, err
}
