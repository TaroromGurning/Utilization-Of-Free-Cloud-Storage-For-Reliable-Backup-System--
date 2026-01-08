package models

import "time"

// ScheduledJob merepresentasikan template Job (Auto atau Manual)
type ScheduledJob struct {
	ID      uint   `gorm:"primaryKey;type:int unsigned"`
	UserID  uint   `gorm:"index;type:int unsigned;not null"`
	JobName string `gorm:"column:job_name;size:100;not null"`

	OperationMode string `gorm:"type:enum('BACKUP','RESTORE');not null"`

	RcloneMode      string `gorm:"column:rclone_mode;type:enum('copy','sync');not null"`
	SourcePath      string `gorm:"size:255;not null"`
	RemoteName      string `gorm:"size:100;not null"`
	DestinationPath string `gorm:"size:255;not null"`

	// Script Kustom (Arsitektur "Script Runner")
	PreScript    string `gorm:"column:pre_script;type:text"`
	PostScript   string `gorm:"column:post_script;type:text"`
	MaxRetention int    `gorm:"default:10"`
	// Penjadwalan dan Status
	ScheduleCron string     `gorm:"size:50;nullable"` // Boleh NULL
	Priority     int        `gorm:"default:5"`
	StatusQueue  string     `gorm:"type:enum('PENDING','RUNNING','COMPLETED','FAIL_PRE_SCRIPT','FAIL_RCLONE','FAIL_POST_SCRIPT','FAIL_SOURCE_CHECK');default:'PENDING'"`
	LastRun      *time.Time `gorm:"column:last_run_at;nullable"`

	CreatedAt time.Time
	UpdatedAt time.Time

	// Relasi GORM
	User User  `gorm:"foreignKey:UserID"`
	Logs []Log `gorm:"foreignKey:JobID"`
}
