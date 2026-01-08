package models

import "time"

type Log struct {
	ID               uint    `gorm:"primaryKey"`
	JobID            *uint   `gorm:"column:job_id;index"`
	JobName          string  `gorm:"size:100;nullable"` // ✅ BARU: Nama job
	SourcePath       string  `gorm:"size:255;nullable"`
	Status           string  `gorm:"type:enum('SUCCESS', 'FAIL_PRE_SCRIPT', 'FAIL_RCLONE', 'FAIL_POST_SCRIPT', 'FAIL_SOURCE_CHECK', 'ERROR')"`
	ConfigSnapshot   *string `gorm:"type:json;nullable"`
	Message          string  `gorm:"type:text"`
	DurationSec      int     `gorm:"column:duration_sec"`
	TransferredBytes int64   `gorm:"column:transferred_bytes;default:0"` // ✅ SUDAH AD
	Timestamp        time.Time
	ScheduledJob     ScheduledJob `gorm:"foreignKey:JobID"`
}
