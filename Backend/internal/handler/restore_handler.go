package handler

import (
	"fmt"
	"gbackup-new/backend/internal/models"
	"gbackup-new/backend/internal/service"
	"net/http"

	"github.com/labstack/echo/v4"
)

type RestoreRequestDTO struct {
	SourcePath      string `json:"source_path" validate:"required"`
	RemoteName      string `json:"remote_name" validate:"required"`
	DestinationPath string `json:"destination_path" validate:"required"`
}

type RestoreHandler struct {
	BackupSvc service.BackupService
}

func NewRestoreHandler(svc service.BackupService) *RestoreHandler {
	return &RestoreHandler{BackupSvc: svc}
}

func (h *RestoreHandler) TriggerRestore(c echo.Context) error {
	var req RestoreRequestDTO

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Format input JSON tidak valid.",
		})
	}

	userID := uint(1)

	restoreJob := &models.ScheduledJob{
		UserID:          userID,
		JobName:         fmt.Sprintf("Restore-%s", req.RemoteName),
		OperationMode:   "RESTORE",
		RcloneMode:      "copy",
		SourcePath:      req.SourcePath,
		RemoteName:      req.RemoteName,
		DestinationPath: req.DestinationPath,
		ScheduleCron:    "",
		StatusQueue:     "PENDING",
	}

	if err := h.BackupSvc.CreateJobAndDispatch(restoreJob); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Gagal memulai Restore: %v", err.Error()),
		})
	}

	return c.JSON(http.StatusAccepted, map[string]interface{}{
		"message":     "Restore job dimulai.",
		"source":      req.SourcePath,
		"remote":      req.RemoteName,
		"destination": req.DestinationPath,
	})
}
