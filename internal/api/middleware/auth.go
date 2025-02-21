package middleware

import (
	"net/http"
	"time"

	"github.com/frolmr/gophermart/internal/api/auth"
	"github.com/frolmr/gophermart/internal/config"
	"github.com/frolmr/gophermart/internal/domain"
	"github.com/frolmr/gophermart/pkg/formatter"
	"github.com/golang-jwt/jwt/v5"
)

func WithAuth(authCfg *config.AuthConfig) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, req *http.Request) {
			accessCookie, err := req.Cookie("access_token")
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			accessTokenString := accessCookie.Value

			claims := &auth.Claims{}
			accessToken, err := jwt.ParseWithClaims(accessTokenString, claims, func(token *jwt.Token) (interface{}, error) {
				return authCfg.JWTKey, nil
			})

			if err != nil || !accessToken.Valid {
				refreshCookie, err := req.Cookie("refresh_token")
				if err != nil {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}

				refreshTokenString := refreshCookie.Value

				refreshToken, err := jwt.ParseWithClaims(refreshTokenString, claims, func(token *jwt.Token) (interface{}, error) {
					return authCfg.JWTKey, nil
				})

				if err != nil || !refreshToken.Valid {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}

				newAccessToken, err := auth.GenerateAccessToken(claims.UserID, authCfg)
				if err != nil {
					http.Error(w, "Failed to generate access token", http.StatusInternalServerError)
					return
				}

				http.SetCookie(w, &http.Cookie{
					Name:     "access_token",
					Value:    newAccessToken,
					Expires:  time.Now().Add(authCfg.JWTAccessTokenExpiresIn),
					HttpOnly: true,
					Path:     "/",
				})

				req.Header.Set("Authorization", "Bearer "+newAccessToken)
			}

			req.Header.Set(domain.UserIDHeader, formatter.Int64ToString(claims.UserID))
			next.ServeHTTP(w, req)
		}

		return http.HandlerFunc(fn)
	}
}
