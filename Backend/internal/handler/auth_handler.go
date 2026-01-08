package handler

import (
	"net/http"

	"gbackup-new/backend/internal/service" // Sesuaikan path module

	"github.com/labstack/echo/v4"
)

// AuthHandler struct menampung dependency ke AuthService
type AuthHandler struct {
	AuthSvc service.AuthService
}

func NewAuthHandler(authSvc service.AuthService) *AuthHandler {
	return &AuthHandler{AuthSvc: authSvc}
}

// Login: Endpoint POST /api/v1/auth/login (Endpoint Publik)
func (h *AuthHandler) Login(c echo.Context) error {
	req := new(service.LoginRequest)

	// 1. Binding JSON request body
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Format input tidak valid."})
	}

	// 2. Panggil Auth Service untuk verifikasi dan generate token
	token, err := h.AuthSvc.Authenticate(req)

	if err != nil {
		// Mengembalikan 401 Unauthorized jika kredensial tidak valid
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Kredensial tidak valid."})
	}

	// 3. Kirim token JWT kembali ke frontend
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Login berhasil.",
		"token":   token,
	})
}
