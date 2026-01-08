package models

import "time"

// Remote merepresentasikan entri remote storage (Untuk fitur Add Remote)
type Remote struct {
	ID         uint   `gorm:"primaryKey;type:int unsigned"`
	Name       string `gorm:"size:100;unique;not null"`
	RemoteType string `gorm:"column:remote_type;size:50;not null"`
	ConfigJSON string `gorm:"column:config_json;type:json"`
	IsActive   bool   `gorm:"default:true"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
