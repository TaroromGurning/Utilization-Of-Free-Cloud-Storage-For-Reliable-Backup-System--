package models

import "time"

// User merepresentasikan satu baris di tabel 'users'
type User struct {
	ID           uint   `gorm:"primaryKey;type:int unsigned"`
	Username     string `gorm:"size:50;unique;not null"`
	Email        string `gorm:"size:100;unique;not null"`
	PasswordHash string `gorm:"column:password_hash;size:255;not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
