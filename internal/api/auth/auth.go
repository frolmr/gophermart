package auth

import (
	"fmt"
	"time"

	"github.com/frolmr/gophermart/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(userID int64, authConf *config.AuthConfig) (string, error) {
	expirationTime := time.Now().Add(authConf.JWTAccessTokenExpiresIn)

	return generateToken(userID, expirationTime, authConf.JWTKey)
}

func GenerateRefreshToken(userID int64, authConf *config.AuthConfig) (string, error) {
	expirationTime := time.Now().Add(authConf.JWTRefreshTokenExpiresIn)

	return generateToken(userID, expirationTime, authConf.JWTKey)
}

func generateToken(userID int64, expirationTime time.Time, key []byte) (string, error) {
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("error signing: %w", err)
	}

	return tokenString, nil
}
