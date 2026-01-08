package database

import (
	"fmt"
	"log"
	"os"

	"gbackup-new/backend/internal/models" // Sesuaikan dengan module baru Anda

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

// Connect: Membaca ENV, terhubung ke DB, dan menjalankan AutoMigrate
func Connect() *gorm.DB {

	// Ambil kredensial dari Environment
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	// Buat DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser,
		dbPass,
		dbHost,
		dbPort,
		dbName)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ FATAL: Gagal terhubung ke database (Cek DSN dan ENV): %v", err)
	}

	fmt.Println("✅ Koneksi database berhasil.")

	// AutoMigrate (dengan Models baru)
	err = DB.AutoMigrate(
		&models.User{},
		&models.ScheduledJob{},
		&models.Log{},
		&models.Monitoring{},
		&models.Remote{},
	)
	if err != nil {
		log.Fatalf("❌ Gagal melakukan AutoMigrate tabel: %v", err)
	}
	fmt.Println("✅ AutoMigrate tabel selesai.")

	return DB
}
