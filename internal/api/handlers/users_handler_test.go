package handlers

import (
	"bytes"
	"encoding/json"
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

			handler.RegisterUser(&config.AuthConfig{JWTExpiresIn: time.Hour}).ServeHTTP(w, req)

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

			handler.LoginUser(&config.AuthConfig{JWTExpiresIn: time.Hour}).ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)
		})
	}
}
