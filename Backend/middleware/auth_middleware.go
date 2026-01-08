package middleware

import (
	"net/http"

	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

// JWTGuard: Middleware yang melindungi endpoint API.
func JWTGuard(jwtSecretKey string) echo.MiddlewareFunc {
	// Gunakan middleware JWT bawaan Echo
	return echojwt.WithConfig(echojwt.Config{
		SigningKey:  []byte(jwtSecretKey),
		TokenLookup: "header:Authorization", // Mencari token di Authorization header
		ErrorHandler: func(c echo.Context, err error) error {
			// Mengembalikan respons 401 jika token hilang atau tidak valid
			return echo.NewHTTPError(http.StatusUnauthorized, "Token hilang atau tidak valid.")
		},
	})
}
