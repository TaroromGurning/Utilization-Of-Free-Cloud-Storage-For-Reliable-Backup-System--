package repository

import (
	"fmt"
	"gbackup-new/backend/internal/models"

	"gorm.io/gorm"
)

const maxLogs = 20

// LogRepository mendefinisikan kontrak
type LogRepository interface {
	CreateLog(log *models.Log) error
	FindAllLogs() ([]models.Log, error)
}

type logRepositoryImpl struct {
	DB *gorm.DB
}

func NewLogRepository(db *gorm.DB) LogRepository {
	return &logRepositoryImpl{DB: db}
}

// CreateLog: Mencatat hasil eksekusi (dipanggil oleh Service)
func (r *logRepositoryImpl) CreateLog(log *models.Log) error {
	if err := r.DB.Create(log).Error; err != nil {
		return fmt.Errorf("gagal menyimpan log baru: %w", err)
	}

	var count int64
	if err := r.DB.Model(&models.Log{}).Count(&count).Error; err != nil {
		fmt.Printf("Gagal menghitung Total Log: %v\n", err)
		return nil
	}

	if count > maxLogs {
		toDelete := count - maxLogs
		var oldlogs []models.Log

		if err := r.DB.Order("timestamp ASC").Limit(int(toDelete)).Find(&oldlogs).Error; err != nil {
			return fmt.Errorf("gagal mengambil log paling tua: %w", &err)
		}

		if len(oldlogs) > 0 {
			// Ambil ID dari record log tertua
			var idsToDelete []uint
			for _, l := range oldlogs {
				idsToDelete = append(idsToDelete, l.ID)
			}

			// Hapus semua log tertua dalam satu transaksi
			if err := r.DB.Where("id IN (?)", idsToDelete).Delete(&models.Log{}).Error; err != nil {
				return fmt.Errorf("gagal menghapus log tertua: %w", err)
			}
			fmt.Sprintf("[LOG CLEANUP] Berhasil menghapus %d log tertua (Max: %d)\n", toDelete, maxLogs)

		}
	}
	return nil
}

// FindAllLogs: Mengambil riwayat untuk UI Logs Page
func (r *logRepositoryImpl) FindAllLogs() ([]models.Log, error) {
	var logs []models.Log
	// Preload "ScheduledJob" untuk mengambil JobName (jika JobID tidak NULL)
	result := r.DB.Preload("ScheduledJob").Order("timestamp desc").Find(&logs)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, result.Error
	}
	return logs, nil
}
