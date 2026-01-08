package service

import (
	"fmt"
	"time"

	"gbackup-new/backend/internal/models"
	"gbackup-new/backend/internal/repository"

	jwt "github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecretKey string // Global variable

// Struct untuk payload login
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthService interface (Kontrak)
type AuthService interface {
	Authenticate(req *LoginRequest) (string, error)
	RegisterAdmin(username, password string) error
	// ✅ Tambahkan method ini untuk mengecek status sistem
	IsAdminSetupComplete() (bool, error)
}

// authServiceImpl: Struct implementasi
type authServiceImpl struct {
	UserRepo  repository.UserRepository
	SecretKey string
}

// NewAuthService: Constructor untuk Dependency Injection
func NewAuthService(uRepo repository.UserRepository, secretKey string) AuthService {
	return &authServiceImpl{
		UserRepo:  uRepo,
		SecretKey: secretKey,
	}
}

// ----------------------------------------------------
// LOGIKA UTAMA OTENTIKASI
// ----------------------------------------------------

// Authenticate: Memverifikasi user dan menghasilkan JWT.
func (s *authServiceImpl) Authenticate(req *LoginRequest) (string, error) {
	// Cari user di DB
	user, err := s.UserRepo.FindByUsername(req.Username)
	if err != nil {
		return "", fmt.Errorf("error database saat mencari user: %w", err)
	}
	if user == nil {
		return "", fmt.Errorf("kredensial tidak valid (user tidak ditemukan)")
	}

	// Verifikasi password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return "", fmt.Errorf("kredensial tidak valid (password salah)")
	}

	// Generate JWT
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Gunakan s.SecretKey yang sudah di-inject
	signedToken, err := token.SignedString([]byte(s.SecretKey))
	if err != nil {
		return "", fmt.Errorf("gagal menandatangani token JWT: %w", err)
	}

	return signedToken, nil
}

// RegisterAdmin: Untuk registrasi user admin pertama dari input form
func (s *authServiceImpl) RegisterAdmin(username, password string) error {
	// Pengecekan Duplikasi (tetap penting)
	user, err := s.UserRepo.FindByUsername(username)
	if err != nil {
		return err
	}
	if user != nil {
		return fmt.Errorf("user admin sudah terdaftar")
	}

	// Pengecekan KRITIS: Pastikan belum ada user terdaftar (hanya untuk setup pertama)
	// Cek ini akan dilakukan di Handler, tapi sebaiknya diulang di sini untuk keamanan service
	isComplete, err := s.IsAdminSetupComplete()
	if err != nil {
		return err
	}
	if isComplete {
		return fmt.Errorf("setup telah selesai. Registrasi admin ditolak")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("gagal hash password: %w", err)
	}

	newUser := &models.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
		Email:        fmt.Sprintf("%s@internal.com", username), // Email default
		// Anda mungkin perlu menambahkan Role: "admin" di model User jika menggunakan sistem role
	}
	return s.UserRepo.CreateUser(newUser)
}

// ----------------------------------------------------
// LOGIKA SETUP WIZARD (KONTROL AKSES)
// ----------------------------------------------------

// IsAdminSetupComplete: Memeriksa apakah akun admin pertama sudah dibuat.
func (s *authServiceImpl) IsAdminSetupComplete() (bool, error) {
	// ASUMSI: Setup dianggap selesai jika ada user manapun di database.
	// Asumsi UserRepo memiliki CountUsers() method
	count, err := s.UserRepo.CountUsers()
	if err != nil {
		return false, fmt.Errorf("gagal menghitung user di database: %w", err)
	}
	// Jika count > 0, setup sudah selesai
	return count > 0, nil
}
