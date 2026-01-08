package repository

import (
	"errors"
	"fmt"
	"gbackup-new/backend/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MonitoringRepository interface {
	UpsertRemoteStatus(monitor *models.Monitoring) error
	FindAllRemotes() ([]models.Monitoring, error)
	FindRemoteByName(remoteName string) (*models.Monitoring, error)
	DeleteRemoteByName(remoteName string) error
	GetAllRemoteNames() ([]string, error)
}

type monitoringRepositoryImpl struct {
	DB *gorm.DB
}

func NewMonitoringRepository(db *gorm.DB) MonitoringRepository {
	return &monitoringRepositoryImpl{DB: db}
}

func (r *monitoringRepositoryImpl) UpsertRemoteStatus(monitor *models.Monitoring) error {
	result := r.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "remote_name"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"status_connect",
			"total_storage_gb",
			"used_storage_gb",
			"free_storage_gb",
			"last_checked_at",
			"system_message",
			"active_job_count",
			"owner_email",
		}),
	}).Create(monitor)

	return result.Error
}

func (r *monitoringRepositoryImpl) FindAllRemotes() ([]models.Monitoring, error) {
	var remotes []models.Monitoring
	result := r.DB.Find(&remotes)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, result.Error
	}

	return remotes, nil
}

func (r *monitoringRepositoryImpl) FindRemoteByName(remoteName string) (*models.Monitoring, error) {
	var monitor models.Monitoring
	result := r.DB.Where("remote_name = ?", remoteName).First(&monitor)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("remote %s tidak ditemukan", remoteName)
		}
		return nil, result.Error
	}
	return &monitor, nil
}

func (r *monitoringRepositoryImpl) DeleteRemoteByName(remoteName string) error {
	result := r.DB.Where("remote_name = ?", remoteName).Delete(&models.Monitoring{})
	return result.Error
}

func (r *monitoringRepositoryImpl) GetAllRemoteNames() ([]string, error) {
	var names []string
	result := r.DB.Model(&models.Monitoring{}).Pluck("remote_name", &names)
	return names, result.Error
}
