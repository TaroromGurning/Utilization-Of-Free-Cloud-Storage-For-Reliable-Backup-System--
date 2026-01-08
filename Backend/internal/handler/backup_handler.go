package handler

import (
	"fmt"
	"net/http"

	"gbackup-new/backend/internal/models"
	"gbackup-new/backend/internal/service"

	"github.com/labstack/echo/v4"
)

// BackupRequestDTO (Data Transfer Object) untuk menerima input Job Baru
type BackupRequestDTO struct {
	JobName         string `json:"job_name" validate:"required"`
	SourcePath      string `json:"source_path" validate:"required"`
	RemoteName      string `json:"remote_name" validate:"required"`
	DestinationPath string `json:"destination_path" validate:"required"`
	ScheduleCron    string `json:"schedule_cron"`
	// ⭐ NEW: Tambah 2 field baru untuk support COPY & SYNC
	OperationMode string `json:"operation_mode"` // BACKUP, RESTORE (default: BACKUP)
	RcloneMode    string `json:"rclone_mode"`    // copy, sync (default: copy)
	PreScript     string `json:"pre_script"`
	PostScript    string `json:"post_script"`
	MaxRetention  int    `json:"max_retention"`
}

type BackupHandler struct {
	BackupSvc service.BackupService
}

// NewBackupHandler adalah constructor (Factory)
func NewBackupHandler(svc service.BackupService) *BackupHandler {
	return &BackupHandler{BackupSvc: svc}
}

// ============================================================
// CreateNewJob: Endpoint POST /api/v1/backup/create
// ============================================================
func (h *BackupHandler) CreateNewJob(c echo.Context) error {
	var req BackupRequestDTO

	// 1. Binding dan Validasi Input
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Format input JSON tidak valid",
		})
	}

	// 2. Validate required fields
	if req.JobName == "" || req.SourcePath == "" || req.RemoteName == "" || req.DestinationPath == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Field required: job_name, source_path, remote_name, destination_path",
		})
	}

	// ⭐ HIGHLIGHT 1: SET DEFAULT OPERATION MODE
	// ✅ Jika OperationMode kosong, default ke BACKUP
	if req.OperationMode == "" {
		req.OperationMode = "BACKUP"
	}

	// ⭐ HIGHLIGHT 1.5: VALIDATE OPERATION MODE
	// ✅ Check apakah valid value (BACKUP atau RESTORE)
	if req.OperationMode != "BACKUP" && req.OperationMode != "RESTORE" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid operation_mode. Must be 'BACKUP' or 'RESTORE'",
		})
	}

	// ⭐ HIGHLIGHT 2: SET DEFAULT RCLONE MODE
	// ✅ Jika RcloneMode kosong, default ke "copy"
	if req.RcloneMode == "" {
		req.RcloneMode = "copy"
	}

	// ⭐ HIGHLIGHT 3: VALIDATE RCLONE MODE
	// ✅ Check apakah valid value (copy atau sync)
	if req.RcloneMode != "copy" && req.RcloneMode != "sync" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid rclone_mode. Must be 'copy' or 'sync'",
		})
	}

	// ⭐ HIGHLIGHT 4: CONDITIONAL VALIDATION BERDASARKAN MODE
	// ✅ Logic berbeda untuk COPY vs SYNC
	fmt.Printf("[HANDLER] RcloneMode: %s\n", req.RcloneMode)

	if req.RcloneMode == "copy" {
		// ✅ COPY MODE: MaxRetention wajib 1-100
		if req.MaxRetention <= 0 {
			req.MaxRetention = 10 // Default 10
			fmt.Printf("[HANDLER] MaxRetention not provided, using default: %d\n", req.MaxRetention)
		}

		if req.MaxRetention > 100 {
			req.MaxRetention = 100
			fmt.Printf("[HANDLER] MaxRetention capped at 100\n")
		}

		fmt.Printf("[HANDLER VALIDATION] COPY Mode: MaxRetention = %d ✅\n", req.MaxRetention)

	} else if req.RcloneMode == "sync" {
		// ✅ SYNC MODE: MaxRetention HARUS 0 (diabaikan)
		// ⭐ KEY POINT: Jika user kirim max_retention apapun, akan di-force ke 0
		req.MaxRetention = 0
		fmt.Printf("[HANDLER VALIDATION] SYNC Mode: MaxRetention forced to 0 ✅\n")
	}

	// Placeholder untuk user ID
	userID := uint(1)

	// 3. Buat Struct Job
	// ⭐ HIGHLIGHT 5: GUNAKAN DARI REQUEST, BUKAN HARDCODE
	// ❌ LAMA: rcloneMode := "COPY" (hardcoded!)
	// ✅ BARU: Ambil dari req.RcloneMode dan req.OperationMode
	newJob := models.ScheduledJob{
		UserID:          userID,
		JobName:         req.JobName,
		OperationMode:   req.OperationMode, // ⭐ NEW: dari request
		RcloneMode:      req.RcloneMode,    // ⭐ NEW: dari request (bukan hardcoded!)
		SourcePath:      req.SourcePath,
		RemoteName:      req.RemoteName,
		DestinationPath: req.DestinationPath,
		PreScript:       req.PreScript,
		PostScript:      req.PostScript,
		ScheduleCron:    req.ScheduleCron,
		StatusQueue:     "PENDING",
		MaxRetention:    req.MaxRetention, // ⭐ SUDAH DIVALIDASI
	}

	// 4. Panggil Service untuk Dispatch Job
	if err := h.BackupSvc.CreateJobAndDispatch(&newJob); err != nil {
		fmt.Printf("[HANDLER ERROR] CreateJobAndDispatch failed: %v\n", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Gagal mendispatch Job: %v", err),
		})
	}

	// 5. Determine schedule type
	scheduleType := "manual"
	if req.ScheduleCron != "" {
		scheduleType = "scheduled"
	}

	fmt.Printf("[HANDLER SUCCESS] Job created - ID: %d, Mode: %s, OperationMode: %s, MaxRetention: %d\n",
		newJob.ID, req.RcloneMode, req.OperationMode, req.MaxRetention)

	// 6. Respon Sukses
	// ⭐ HIGHLIGHT 6: RESPONSE LEBIH DETAIL
	// ✅ Tampilkan rclone_mode, operation_mode, max_retention, schedule_type
	return c.JSON(http.StatusAccepted, map[string]interface{}{
		"success":        true,
		"message":        "Job berhasil diterima dan dikirim untuk eksekusi",
		"job_id":         newJob.ID,
		"rclone_mode":    req.RcloneMode,    // ⭐ NEW: show rclone mode
		"operation_mode": req.OperationMode, // ⭐ NEW: show operation mode
		"max_retention":  req.MaxRetention,  // ⭐ NEW: show max retention
		"schedule_type":  scheduleType,      // ⭐ NEW: show manual/scheduled
	})
}
