package handler

import (
	"fmt"
	"net/http"

	"gbackup-new/backend/internal/service"

	"github.com/labstack/echo/v4"
)

// LoginRequest merepresentasikan payload untuk permintaan login
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// SetupHandler memegang dependency untuk Setup Wizard
type SetupHandler struct {
	AuthSvc service.AuthService // Menggunakan interface AuthService
}

// NewSetupHandler adalah constructor untuk SetupHandler
func NewSetupHandler(authSvc service.AuthService) *SetupHandler {
	return &SetupHandler{
		AuthSvc: authSvc,
	}
}

// ====================================================
// SETUP WIZARD ENDPOINTS
// ====================================================

// GetSetupStatus memeriksa apakah admin pertama sudah terdaftar.
// Endpoint: GET /api/v1/setup/status
func (h *SetupHandler) GetSetupStatus(c echo.Context) error {
	// Memanggil service untuk mengecek status setup
	isComplete, err := h.AuthSvc.IsAdminSetupComplete()
	if err != nil {
		c.Logger().Error(fmt.Errorf("error checking admin setup status: %w", err))
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Gagal memeriksa status setup dari database",
		})
	}

	// ⭐ KOREKSI: Menggunakan key "is_admin_registered" yang diharapkan frontend
	return c.JSON(http.StatusOK, map[string]bool{
		"is_admin_registered": isComplete,
	})
}

// RegisterInitialAdmin mendaftarkan user admin pertama.
// Endpoint: POST /api/v1/setup/register
func (h *SetupHandler) RegisterInitialAdmin(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Input tidak valid"})
	}

	if req.Username == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Username dan password wajib diisi"})
	}

	// Cek sekali lagi di lapisan handler sebelum registrasi
	isComplete, err := h.AuthSvc.IsAdminSetupComplete()
	if err != nil {
		c.Logger().Error(fmt.Errorf("error saat cek status sebelum registrasi: %w", err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal server error"})
	}
	if isComplete {
		return c.JSON(http.StatusConflict, map[string]string{"message": "Setup telah selesai. Admin sudah ada."})
	}

	// Lakukan registrasi
	if err := h.AuthSvc.RegisterAdmin(req.Username, req.Password); err != nil {
		// Log error di server
		c.Logger().Error(fmt.Errorf("gagal mendaftarkan admin: %w", err))
		// Kirim respons user friendly
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": fmt.Sprintf("Gagal registrasi: %s", err.Error())})
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "Admin berhasil didaftarkan. Silakan login."})
}

// CATATAN: setupHandler.Login dihapus/tidak diimplementasikan karena
// rute login publik sudah ditangani oleh authHandler.Login.
