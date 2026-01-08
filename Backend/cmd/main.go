package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"gbackup-new/backend/internal/handler"
	"gbackup-new/backend/internal/repository"
	"gbackup-new/backend/internal/service"
	"gbackup-new/backend/pkg/database"

	"github.com/joho/godotenv"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	echomid "github.com/labstack/echo/v4/middleware"
)

func loadDotEnv() {
	exePath, err := os.Executable()
	if err != nil {
		fmt.Println("Tidak bisa dapatkan path executable:", err)
		return
	}
	exeDir := filepath.Dir(exePath)

	candidates := []string{
		filepath.Join(exeDir, ".env"),
		filepath.Join(exeDir, "..", ".env"),
		filepath.Join(exeDir, "..", "..", ".env"),
	}

	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			if err := godotenv.Load(p); err == nil {
				fmt.Println("✅ Loaded .env dari:", p)
				return
			}
		}
	}

	_ = godotenv.Load(".env", "../.env")
}

func main() {
	loadDotEnv()

	jwtSecretKey := os.Getenv("JWT_SECRET_KEY")
	if jwtSecretKey == "" {
		log.Fatal("FATAL: JWT_SECRET_KEY tidak ada")
	}

	dbInstance := database.Connect()
	if dbInstance == nil {
		log.Fatal("Koneksi DB gagal, instance GORM nil.")
	}

	// Repositories
	userRepo := repository.NewUserRepository(dbInstance)
	jobRepo := repository.NewJobRepository(dbInstance)
	logRepo := repository.NewLogRepository(dbInstance)
	monitorRepo := repository.NewMonitoringRepository(dbInstance)
	browserRepo := repository.NewBrowserRepository()

	// Services
	authSvc := service.NewAuthService(userRepo, jwtSecretKey)
	monitorSvc := service.NewMonitoringService(monitorRepo, logRepo, jobRepo)
	backupSvc := service.NewBackupService(jobRepo, logRepo, monitorRepo, monitorSvc)
	schedulerSvc := service.NewSchedulerService(jobRepo, backupSvc)
	browserSvc := service.NewBrowserService(browserRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authSvc)
	monitorHandler := handler.NewMonitoringHandler(monitorSvc, schedulerSvc, logRepo)
	jobHandler := handler.NewJobHandler(schedulerSvc, backupSvc, jobRepo)
	backupHandler := handler.NewBackupHandler(backupSvc)
	restoreHandler := handler.NewRestoreHandler(backupSvc)
	browserHandler := handler.NewBrowserHandler(browserSvc)
	setupHandler := handler.NewSetupHandler(authSvc)

	// Echo Setup
	e := echo.New()
	e.Use(echomid.Logger())
	e.Use(echomid.Recover())
	e.Use(echomid.CORSWithConfig(echomid.CORSConfig{
		AllowOrigins: []string{"http://localhost:5173"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodPut},
	}))

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "G-Backup New Architecture is Running!")
	})

	// Public Routes
	e.POST("/api/v1/auth/login", authHandler.Login)

	setupGroup := e.Group("/api/v1/setup")
	setupGroup.GET("/status", setupHandler.GetSetupStatus)
	setupGroup.POST("/register", setupHandler.RegisterInitialAdmin)

	// Protected Routes
	r := e.Group("/api/v1")
	r.Use(echojwt.WithConfig(echojwt.Config{SigningKey: []byte(jwtSecretKey)}))

	// Monitoring
	r.GET("/monitoring/remotes", monitorHandler.GetRemoteStatusList)
	r.GET("/monitoring/logs", monitorHandler.GetJobLogs)
	r.GET("/monitoring/jobs", monitorHandler.GetScheduledJobs)
	r.GET("/monitoring/drivemail", monitorHandler.GetRemotes)

	// Jobs
	r.GET("/jobs/scheduled", jobHandler.GetScheduledJobs)
	r.GET("/jobs/script/:id", jobHandler.GetJobScript)
	r.POST("/jobs/trigger/:id", jobHandler.TriggerManualJob)
	r.GET("/jobs/manual", jobHandler.GetManualJob)
	r.DELETE("/jobs/delete/:id", jobHandler.DeleteJob)
	r.PUT("/jobs/update/:id", jobHandler.UpdateJob)
	r.GET("/jobs/:id", jobHandler.GetJobByID)
	r.GET("/jobs/alljobs", monitorHandler.GetAllJobs)

	// Actions
	r.POST("/jobs/new", backupHandler.CreateNewJob)
	r.POST("/jobs/restore", restoreHandler.TriggerRestore)
	r.GET("/browser/files", browserHandler.ListFiles)
	r.GET("/browser/remotes", browserHandler.GetAvailableRemotes)
	r.GET("/browser/info", browserHandler.GetFileInfo)

	// Start Daemons
	schedulerSvc.StartDaemon()
	monitorSvc.StartMonitoringDaemon()

	go func() {
		time.Sleep(2 * time.Second)
		fmt.Println("[Startup] Inisialisasi status remote monitoring...")

		if err := monitorSvc.SyncRemotesWithRclone(); err != nil {
			fmt.Printf("❌ ERROR: Gagal sync remote: %v\n", err)
			return
		}

		remotes, err := monitorSvc.GetRemoteStatusList()
		if err != nil {
			fmt.Printf("❌ ERROR: Gagal mengambil remote: %v\n", err)
			return
		}

		if len(remotes) > 0 {
			fmt.Printf("✅ Menemukan %d remote, memulai update status...\n", len(remotes))
			for _, remote := range remotes {
				go monitorSvc.UpdateRemoteStatus(remote.RemoteName)
			}
		} else {
			fmt.Println("⚠️  Tidak ada remote yang dikonfigurasi di rclone.conf")
		}

		fmt.Println("✅ Inisialisasi remote monitoring selesai.")
	}()

	fmt.Println("\n✅ Backend diinisialisasi. Menjalankan Echo server di port 8080")
	e.Logger.Fatal(e.Start(":8080"))
}
