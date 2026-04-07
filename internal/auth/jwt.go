package auth

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type MyClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func GenerateTokens(userID string) (string, string, error) {
	// Access Token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, MyClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		},
	})
	aToken, err := accessToken.SignedString([]byte(os.Getenv("JWT_ACCESS_SECRET")))
	// Refresh Token
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)),
	})
	rToken, err := refreshToken.SignedString([]byte(os.Getenv("JWT_REFRESH_SECRET")))

	return aToken, rToken, err
}
