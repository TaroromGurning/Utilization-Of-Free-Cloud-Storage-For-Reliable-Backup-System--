package handler

import (
	"fmt"
	"gbackup-new/backend/internal/models"
	"gbackup-new/backend/internal/repository"
	"gbackup-new/backend/internal/service"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// JobHandler struct menampung dependency (Service Layer)
type JobHandler struct {
	SchedulerSvc service.SchedulerService
	BackupSvc    service.BackupService
	JobRepo      repository.JobRepository
}

// NewJobHandler adalah constructor (Factory)
func NewJobHandler(
	schedulerSvc service.SchedulerService,
	backupSvc service.BackupService,
	repo repository.JobRepository,
) *JobHandler {
	return &JobHandler{
		SchedulerSvc: schedulerSvc,
		BackupSvc:    backupSvc,
		JobRepo:      repo,
	}
}

// ============================================================
// GetJobByID: GET /api/v1/jobs/:id
// ============================================================
func (h *JobHandler) GetJobByID(c echo.Context) error {
	jobIDStr := c.Param("id")
	jobID, err := strconv.ParseUint(jobIDStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid job ID format",
		})
	}

	job, err := h.JobRepo.FindJobByID(uint(jobID))
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": err.Error(),
		})
	}

	scheduleType := "manual"
	if job.ScheduleCron != "" {
		scheduleType = "scheduled"
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"id":               job.ID,
			"job_name":         job.JobName,
			"operation_mode":   job.OperationMode,
			"rclone_mode":      job.RcloneMode,
			"source_path":      job.SourcePath,
			"destination_path": job.DestinationPath,
			"remote_name":      job.RemoteName,
			"max_retention":    job.MaxRetention,
			"schedule_cron":    job.ScheduleCron,
			"schedule_type":    scheduleType,
			"status":           job.StatusQueue,
			"last_run":         job.LastRun,
			"pre_script":       job.PreScript,
			"post_script":      job.PostScript,
		},
	})
}

// ============================================================
// GetScheduledJobs: GET /api/v1/jobs/scheduled
// ============================================================
func (h *JobHandler) GetScheduledJobs(c echo.Context) error {
	jobsDTO, err := h.SchedulerSvc.GetScheduledJobsInfo()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Gagal mengambil daftar Job terformat: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, jobsDTO)
}

// ============================================================
// GetManualJob: GET /api/v1/jobs/manual
// ============================================================
func (h *JobHandler) GetManualJob(c echo.Context) error {
	jobsDTO, err := h.SchedulerSvc.GetManualJob()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Gagal mengambil daftar Job Manual: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, jobsDTO)
}

// ============================================================
// GetJobScript: GET /api/v1/jobs/script/:id
// ============================================================
func (h *JobHandler) GetJobScript(c echo.Context) error {
	jobIDStr := c.Param("id")
	jobID, err := strconv.ParseUint(jobIDStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Job ID tidak valid",
		})
	}

	script, err := h.SchedulerSvc.GetGeneratedScript(uint(jobID))
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success":        true,
		"script_preview": script,
	})
}

// ============================================================
// TriggerManualJob: POST /api/v1/jobs/trigger/:id
// ============================================================
func (h *JobHandler) TriggerManualJob(c echo.Context) error {
	jobIDStr := c.Param("id")
	jobID, err := strconv.ParseUint(jobIDStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Job ID tidak valid",
		})
	}

	if err := h.BackupSvc.TriggerManualJob(uint(jobID)); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Gagal memicu Job: %v", err),
		})
	}

	fmt.Printf("[HANDLER] Job triggered - ID: %d\n", jobID)

	return c.JSON(http.StatusAccepted, map[string]interface{}{
		"success": true,
		"message": "Job berhasil dipicu",
		"job_id":  jobID,
	})
}

// ============================================================
// DeleteJob: DELETE /api/v1/jobs/:id
// ============================================================
func (h *JobHandler) DeleteJob(c echo.Context) error {
	jobIDStr := c.Param("id")
	jobID, err := strconv.ParseUint(jobIDStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Job ID tidak valid",
		})
	}

	if err := h.BackupSvc.DeleteJob(uint(jobID)); err != nil {
		fmt.Printf("[HANDLER ERROR] DeleteJob failed: %v\n", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Gagal menghapus Job: %v", err),
		})
	}

	fmt.Printf("[HANDLER] Job deleted - ID: %d\n", jobID)

	return c.JSON(http.StatusAccepted, map[string]interface{}{
		"success": true,
		"message": "Job berhasil dihapus",
		"job_id":  jobID,
	})
}

// ============================================================
// UpdateJob: PUT /api/v1/jobs/:id
// ============================================================
// ⭐ HIGHLIGHT: Update dengan COPY & SYNC mode validation
func (h *JobHandler) UpdateJob(c echo.Context) error {
	jobID := c.Param("id")
	id64, err := strconv.ParseUint(jobID, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid job ID",
		})
	}
	id := uint(id64)

	// ⭐ HIGHLIGHT 1: BIND DENGAN POINTER TYPES
	// ✅ Gunakan *string, *int, *bool agar bisa detect apakah field dikirim atau tidak
	var req struct {
		JobName         *string `json:"job_name"`
		OperationMode   *string `json:"operation_mode"` // ⭐ NEW: dapat di-update
		RcloneMode      *string `json:"rclone_mode"`    // ⭐ NEW: dapat di-update
		SourcePath      *string `json:"source_path"`
		DestinationPath *string `json:"destination_path"`
		RemoteName      *string `json:"remote_name"`
		ScheduleCron    *string `json:"schedule_cron"`
		PreScript       *string `json:"pre_script"`
		PostScript      *string `json:"post_script"`
		MaxRetention    *int    `json:"max_retention"` // ⭐ NEW: dapat di-update
		IsActive        *bool   `json:"is_active"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request format",
		})
	}

	// ⭐ HIGHLIGHT 2: VALIDATE RCLONE MODE CHANGE
	// ✅ Check apakah user mengirim rclone_mode untuk di-update
	if req.RcloneMode != nil {
		fmt.Printf("[HANDLER UPDATE] RcloneMode change detected: %s\n", *req.RcloneMode)

		// ✅ Validate value (harus copy atau sync)
		if *req.RcloneMode != "copy" && *req.RcloneMode != "sync" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid rclone_mode. Must be 'copy' or 'sync'",
			})
		}

		// ⭐ HIGHLIGHT 3: CONDITIONAL LOGIC SAAT MODE CHANGE
		// ✅ KEY POINT: Jika user change ke SYNC, force MaxRetention ke 0
		if *req.RcloneMode == "sync" {
			fmt.Printf("[HANDLER UPDATE] Changing to SYNC mode: MaxRetention forced to 0\n")
			zeroVal := 0
			req.MaxRetention = &zeroVal // ⭐ FORCE MaxRetention = 0
		}

		// ✅ Jika user change ke COPY, ensure MaxRetention ada value
		if *req.RcloneMode == "copy" {
			if req.MaxRetention == nil || *req.MaxRetention <= 0 {
				defaultVal := 10
				req.MaxRetention = &defaultVal // ⭐ Set default 10 untuk COPY
				fmt.Printf("[HANDLER UPDATE] Changing to COPY mode: MaxRetention set to default 10\n")
			}

			if req.MaxRetention != nil && *req.MaxRetention > 100 {
				*req.MaxRetention = 100 // ⭐ Cap at 100
				fmt.Printf("[HANDLER UPDATE] MaxRetention capped at 100\n")
			}
		}
	}

	// ⭐ HIGHLIGHT 4: VALIDATE MAX_RETENTION RANGE
	// ✅ Jika update MaxRetention tapi tidak update RcloneMode, ensure valid range
	if req.MaxRetention != nil && req.RcloneMode == nil {
		if *req.MaxRetention < 0 || *req.MaxRetention > 100 {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "MaxRetention must be between 0 and 100",
			})
		}
	}

	// ⭐ HIGHLIGHT 5: BUILD UPDATED JOB
	// ✅ Hanya set field yang dikirim (pointer pattern)
	updated := &models.ScheduledJob{}

	if req.JobName != nil {
		updated.JobName = *req.JobName
	}
	if req.OperationMode != nil {
		updated.OperationMode = *req.OperationMode
	}
	if req.RcloneMode != nil {
		updated.RcloneMode = *req.RcloneMode // ⭐ NEW
	}
	if req.SourcePath != nil {
		updated.SourcePath = *req.SourcePath
	}
	if req.DestinationPath != nil {
		updated.DestinationPath = *req.DestinationPath
	}
	if req.RemoteName != nil {
		updated.RemoteName = *req.RemoteName
	}
	if req.ScheduleCron != nil {
		updated.ScheduleCron = *req.ScheduleCron
	}
	if req.PreScript != nil {
		updated.PreScript = *req.PreScript
	}
	if req.PostScript != nil {
		updated.PostScript = *req.PostScript
	}
	if req.MaxRetention != nil {
		updated.MaxRetention = *req.MaxRetention // ⭐ NEW: sudah di-validate
	}

	// Call service
	if err := h.BackupSvc.UpdateJob(id, updated); err != nil {
		fmt.Printf("[HANDLER ERROR] UpdateJob failed: %v\n", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	fmt.Printf("[HANDLER] Job updated successfully - ID: %d\n", id)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Job berhasil diperbarui",
		"job_id":  id,
	})
}
