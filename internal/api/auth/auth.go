package auth

import (
	"time"

	"github.com/frolmr/gophermart/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

func GenerateJWT(userID int, authConf *config.AuthConfig) (string, error) {
	expirationTime := time.Now().Add(authConf.JWTExpiresIn)

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(authConf.JWTKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
