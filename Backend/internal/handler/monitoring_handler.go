package handler

import (
	"net/http"
	"time"

	"gbackup-new/backend/internal/repository"
	"gbackup-new/backend/internal/service"

	"github.com/labstack/echo/v4"
)

type MonitoringHandler struct {
	MonitoringSvc service.MonitoringService
	SchedulerSvc  service.SchedulerService
	LogRepo       repository.LogRepository
}

func NewMonitoringHandler(mSvc service.MonitoringService, sSvc service.SchedulerService, lRepo repository.LogRepository) *MonitoringHandler {
	return &MonitoringHandler{
		MonitoringSvc: mSvc,
		SchedulerSvc:  sSvc,
		LogRepo:       lRepo,
	}
}

func (h *MonitoringHandler) SyncRemotes(c echo.Context) error {
	err := h.MonitoringSvc.SyncRemotesWithRclone()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Gagal sync remote: " + err.Error(),
		})
	}

	remotes, err := h.MonitoringSvc.GetRemoteStatusList()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Gagal mengambil remote setelah sync: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Remote berhasil disinkronkan",
		"count":   len(remotes),
		"remotes": remotes,
	})
}

func (h *MonitoringHandler) GetRemoteStatusList(c echo.Context) error {
	remotes, err := h.MonitoringSvc.GetRemoteStatusList()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Gagal mengambil data monitoring: " + err.Error(),
		})
	}

	var responseData []map[string]interface{}
	for _, r := range remotes {
		responseData = append(responseData, map[string]interface{}{
			"remote_name":      r.RemoteName,
			"email":            r.OwnerEmail, // 🆕 Add email
			"status_connect":   r.StatusConnect,
			"used_storage_gb":  r.UsedStorageGB,
			"free_storage_gb":  r.FreeStorageGB,
			"total_storage_gb": r.TotalStorageGB,
			"last_checked_at":  r.LastCheckedAt.Format(time.RFC3339),
			"active_job_count": r.ActiveJobCount,
			"system_message":   r.SystemMessage,
		})
	}

	return c.JSON(http.StatusOK, responseData)
}

// 🆕 GetRemotes: List remotes dengan email dari database
func (h *MonitoringHandler) GetRemotes(c echo.Context) error {
	remotes, err := h.MonitoringSvc.GetRemoteStatusList()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Gagal mengambil daftar remote",
		})
	}

	if len(remotes) == 0 {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"remotes": []interface{}{},
		})
	}

	type RemoteResponse struct {
		Name        string  `json:"name"`
		Email       string  `json:"email"`
		Status      string  `json:"status"`
		TotalGB     float64 `json:"total_gb"`
		UsedGB      float64 `json:"used_gb"`
		FreeGB      float64 `json:"free_gb"`
		UsedPercent float64 `json:"used_percent"`
		ActiveJobs  int64   `json:"active_jobs"`
		LastChecked string  `json:"last_checked"`
		Message     string  `json:"message"`
	}

	var response []RemoteResponse

	for _, remote := range remotes {
		usedPercent := 0.0
		if remote.TotalStorageGB > 0 {
			usedPercent = (remote.UsedStorageGB / remote.TotalStorageGB) * 100
		}

		response = append(response, RemoteResponse{
			Name:        remote.RemoteName,
			Email:       remote.OwnerEmail,
			Status:      remote.StatusConnect,
			TotalGB:     remote.TotalStorageGB,
			UsedGB:      remote.UsedStorageGB,
			FreeGB:      remote.FreeStorageGB,
			UsedPercent: usedPercent,
			ActiveJobs:  remote.ActiveJobCount,
			LastChecked: remote.LastCheckedAt.Format("2006-01-02 15:04:05"),
			Message:     remote.SystemMessage,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"remotes": response,
	})
}

func (h *MonitoringHandler) GetJobLogs(c echo.Context) error {
	logs, err := h.MonitoringSvc.GetJobLogs()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Gagal mengambil riwayat log: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, logs)
}

func (h *MonitoringHandler) GetScheduledJobs(c echo.Context) error {
	jobsDTO, err := h.SchedulerSvc.GetScheduledJobsInfo()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "gagal mengambil daftar jobs: " + err.Error(),
		})
	}
	return c.JSON(http.StatusOK, jobsDTO)
}

// GetAllJobs: Mengambil semua job (Manual + Scheduled) untuk dashboard counter
func (h *MonitoringHandler) GetAllJobs(c echo.Context) error {
	jobs, err := h.MonitoringSvc.GetAllJobs()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Gagal mengambil total jobs: " + err.Error(),
		})
	}

	var response []map[string]interface{}
	for _, job := range jobs {
		response = append(response, map[string]interface{}{
			"id":             job.ID,
			"job_name":       job.JobName,
			"operation_mode": job.OperationMode,
			"schedule_cron":  job.ScheduleCron,
		})
	}

	return c.JSON(http.StatusOK, response)
}
