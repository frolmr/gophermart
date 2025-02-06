package middleware

import (
	"net/http"
	"strconv"

	"github.com/frolmr/gophermart/internal/api/auth"
	"github.com/frolmr/gophermart/internal/config"
	"github.com/frolmr/gophermart/internal/domain"
	"github.com/golang-jwt/jwt/v5"
)

func WithAuth(authCfg *config.AuthConfig) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, req *http.Request) {
			cookie, err := req.Cookie("access_token")
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			tokenString := cookie.Value

			claims := &auth.Claims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				return authCfg.JWTKey, nil
			})

			if err != nil || !token.Valid {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			req.Header.Set(domain.UserIDHeader, strconv.Itoa(claims.UserID))
			next.ServeHTTP(w, req)
		}

		return http.HandlerFunc(fn)
	}
}
