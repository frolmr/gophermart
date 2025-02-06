package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/frolmr/gophermart/internal/api/auth"
	"github.com/frolmr/gophermart/internal/config"
	"github.com/frolmr/gophermart/internal/domain"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type UsersRepository interface {
	CreateAndReturnUser(login, password string) (*domain.DBUser, error)
	GetUserByLogin(login string) (*domain.DBUser, error)
}

type UsersHandler struct {
	logger *zap.SugaredLogger
	repo   UsersRepository
}

func NewUsersHandler(lgr *zap.SugaredLogger, repo UsersRepository) *UsersHandler {
	return &UsersHandler{
		logger: lgr,
		repo:   repo,
	}
}

func (uh *UsersHandler) RegisterUser(authConfig *config.AuthConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var user domain.User

		err := json.NewDecoder(req.Body).Decode(&user)
		if err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		if user.Login == "" || user.Password == "" {
			http.Error(w, "Login and password are required", http.StatusBadRequest)
			return
		}

		existingUser, err := uh.repo.GetUserByLogin(user.Login)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		if existingUser != nil {
			http.Error(w, "User already registered", http.StatusConflict)
			return
		}

		dbUser, err := uh.repo.CreateAndReturnUser(user.Login, user.Password)
		if err != nil {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}

		tokenString, err := auth.GenerateJWT(dbUser.ID, authConfig)
		if err != nil {
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "access_token",
			Value:    tokenString,
			Expires:  time.Now().Add(authConfig.JWTExpiresIn),
			HttpOnly: true,
			Path:     "/",
		})

		_, _ = w.Write([]byte("User created successfully"))
	}
}

func (uh *UsersHandler) LoginUser(authConfig *config.AuthConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var user domain.User

		if err := json.NewDecoder(req.Body).Decode(&user); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		if user.Login == "" || user.Password == "" {
			http.Error(w, "Login and password are required", http.StatusBadRequest)
			return
		}

		dbUser, err := uh.repo.GetUserByLogin(user.Login)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		if dbUser == nil {
			http.Error(w, "Invalid login or password", http.StatusUnauthorized)
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(dbUser.PasswordHash), []byte(user.Password))
		if err != nil {
			http.Error(w, "Invalid login or password", http.StatusUnauthorized)
			return
		}

		tokenString, err := auth.GenerateJWT(dbUser.ID, authConfig)
		if err != nil {
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "access_token",
			Value:    tokenString,
			Expires:  time.Now().Add(authConfig.JWTExpiresIn),
			HttpOnly: true,
			Path:     "/",
		})

		_, _ = w.Write([]byte("Login successful"))
	}
}
