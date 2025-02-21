package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/frolmr/gophermart/internal/config"
	"github.com/frolmr/gophermart/internal/domain"
	"github.com/frolmr/gophermart/internal/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func TestUsersHandler_RegisterUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUsersRepository(ctrl)
	logger := zap.NewNop().Sugar()
	handler := NewUsersHandler(logger, mockRepo)

	tests := []struct {
		name           string
		input          domain.User
		mockSetup      func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Successful registration",
			input: domain.User{
				Login:    "testuser",
				Password: "testpassword",
			},
			mockSetup: func() {
				mockRepo.EXPECT().
					GetUserByLogin("testuser").
					Return(nil, nil)

				mockRepo.EXPECT().
					CreateAndReturnUser("testuser", gomock.Any()).
					Return(&domain.DBUser{ID: 1, Login: "testuser"}, nil)

				mockRepo.EXPECT().
					StoreRefreshToken(int64(1), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "User created successfully",
		},
		{
			name: "User already exists",
			input: domain.User{
				Login:    "existinguser",
				Password: "testpassword",
			},
			mockSetup: func() {
				mockRepo.EXPECT().
					GetUserByLogin("existinguser").
					Return(&domain.DBUser{ID: 1, Login: "existinguser"}, nil)
			},
			expectedStatus: http.StatusConflict,
			expectedBody:   "User already registered",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			body, _ := json.Marshal(tt.input)
			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
			w := httptest.NewRecorder()

			authConfig := &config.AuthConfig{
				JWTKey:                   []byte("secret"),
				JWTAccessTokenExpiresIn:  time.Hour,
				JWTRefreshTokenExpiresIn: time.Hour,
			}

			handler.RegisterUser(authConfig).ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)
		})
	}
}

func TestUsersHandler_LoginUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUsersRepository(ctrl)
	logger := zap.NewNop().Sugar()
	handler := NewUsersHandler(logger, mockRepo)

	tests := []struct {
		name           string
		input          domain.User
		mockSetup      func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Successful login",
			input: domain.User{
				Login:    "testuser",
				Password: "testpassword",
			},
			mockSetup: func() {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("testpassword"), bcrypt.DefaultCost)
				mockRepo.EXPECT().
					GetUserByLogin("testuser").
					Return(&domain.DBUser{ID: 1, Login: "testuser", PasswordHash: string(hashedPassword)}, nil)

				mockRepo.EXPECT().
					StoreRefreshToken(int64(1), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Login successful",
		},
		{
			name: "Invalid login or password",
			input: domain.User{
				Login:    "testuser",
				Password: "wrongpassword",
			},
			mockSetup: func() {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("testpassword"), bcrypt.DefaultCost)
				mockRepo.EXPECT().
					GetUserByLogin("testuser").
					Return(&domain.DBUser{ID: 1, Login: "testuser", PasswordHash: string(hashedPassword)}, nil)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid login or password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			body, _ := json.Marshal(tt.input)
			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
			w := httptest.NewRecorder()

			authConfig := &config.AuthConfig{
				JWTKey:                   []byte("secret"),
				JWTAccessTokenExpiresIn:  time.Hour,
				JWTRefreshTokenExpiresIn: time.Hour,
			}

			handler.LoginUser(authConfig).ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)
		})
	}
}

func TestUsersHandler_RefreshToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUsersRepository(ctrl)
	logger := zap.NewNop().Sugar()
	handler := NewUsersHandler(logger, mockRepo)

	authConfig := &config.AuthConfig{
		JWTKey:                   []byte("secret"),
		JWTAccessTokenExpiresIn:  15 * time.Minute,
		JWTRefreshTokenExpiresIn: 24 * time.Hour,
	}

	tests := []struct {
		name           string
		refreshToken   string
		mockSetup      func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name:         "Successful token refresh",
			refreshToken: "valid-refresh-token",
			mockSetup: func() {
				mockRepo.EXPECT().
					GetRefreshToken("valid-refresh-token").
					Return(&domain.RefreshToken{
						ID:        1,
						UserID:    1,
						Token:     "valid-refresh-token",
						ExpiresAt: time.Now().Add(time.Hour),
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Token refreshed successfully",
		},
		{
			name:         "Invalid refresh token",
			refreshToken: "invalid-refresh-token",
			mockSetup: func() {
				mockRepo.EXPECT().
					GetRefreshToken("invalid-refresh-token").
					Return(nil, nil)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Refresh token is expired or invalid",
		},
		{
			name:         "Expired refresh token",
			refreshToken: "expired-refresh-token",
			mockSetup: func() {
				mockRepo.EXPECT().
					GetRefreshToken("expired-refresh-token").
					Return(&domain.RefreshToken{
						ID:        1,
						UserID:    1,
						Token:     "expired-refresh-token",
						ExpiresAt: time.Now().Add(-time.Hour),
					}, nil)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Refresh token is expired or invalid",
		},
		{
			name:         "Database error",
			refreshToken: "valid-refresh-token",
			mockSetup: func() {
				mockRepo.EXPECT().
					GetRefreshToken("valid-refresh-token").
					Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest(http.MethodPost, "/refresh", nil)
			req.AddCookie(&http.Cookie{Name: "refresh_token", Value: tt.refreshToken})
			w := httptest.NewRecorder()

			handler.RefreshToken(authConfig).ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)
		})
	}
}
