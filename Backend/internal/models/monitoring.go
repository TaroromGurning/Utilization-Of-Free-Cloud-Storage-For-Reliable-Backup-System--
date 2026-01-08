package models

import "time"

type Monitoring struct {
	ID               uint      `gorm:"primaryKey;type:int unsigned"`
	RemoteName       string    `gorm:"column:remote_name;size:100;unique;not null"`
	StatusConnect    string    `gorm:"column:status_connect;type:enum('CONNECTED','DISCONNECTED','PENDING');not null"` // ✅ ADDED PENDING
	SystemMessage    string    `gorm:"column:system_message;size:255;default:''"`
	TotalStorageGB   float64   `gorm:"column:total_storage_gb;default:0"`
	UsedStorageGB    float64   `gorm:"column:used_storage_gb;default:0"`
	FreeStorageGB    float64   `gorm:"column:free_storage_gb;default:0"`
	LastCheckedAt    time.Time `gorm:"column:last_checked_at"`
	ActiveJobCount   int64     `gorm:"default:0"`
	TransferredBytes int64     `gorm:"default:0"`
	OwnerEmail       string    `gorm:"column:owner_email;size:100;default:''"`
}
