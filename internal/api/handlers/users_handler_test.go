package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/frolmr/gophermart/internal/config"
	"github.com/frolmr/gophermart/internal/domain"
	"github.com/frolmr/gophermart/internal/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func readResponseBody(t *testing.T, res *http.Response) string {
	body, err := io.ReadAll(res.Body)
	assert.NoError(t, err)
	return string(body)
}

func TestRegisterUser_Success(t *testing.T) {
	dbUser := domain.DBUser{
		ID:           1,
		Login:        "testuser",
		PasswordHash: "testpassword",
	}
	mockRepo := &mocks.MockUsersRepository{
		CreateAndReturnUserFunc: func(login, password string) (*domain.DBUser, error) {
			return &dbUser, nil
		},
		GetUserByLoginFunc: func(login string) (*domain.DBUser, error) {
			return nil, nil
		},
	}

	logger := zap.NewNop().Sugar()

	authConfig := &config.AuthConfig{
		JWTKey:       []byte("test-secret"),
		JWTExpiresIn: time.Hour,
	}
	handler := NewUsersHandler(logger, mockRepo)

	ts := httptest.NewServer(handler.RegisterUser(authConfig))
	defer ts.Close()

	user := domain.User{
		Login:    "testuser",
		Password: "testpassword",
	}
	payload, _ := json.Marshal(user)

	//nolint:noctx // Do not need context here
	res, err := http.Post(ts.URL+"/register", "application/json", bytes.NewBuffer(payload))
	assert.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Contains(t, readResponseBody(t, res), "User created successfully")

	cookies := res.Cookies()
	assert.NotEmpty(t, cookies)
	assert.Equal(t, "access_token", cookies[0].Name)
}

func TestRegisterUser_UserAlreadyExists(t *testing.T) {
	mockRepo := &mocks.MockUsersRepository{
		GetUserByLoginFunc: func(login string) (*domain.DBUser, error) {
			return &domain.DBUser{Login: "testuser"}, nil // Simulate existing user
		},
	}
	logger := zap.NewNop().Sugar()

	authConfig := &config.AuthConfig{
		JWTKey:       []byte("test-secret"),
		JWTExpiresIn: time.Hour,
	}

	handler := NewUsersHandler(logger, mockRepo)

	ts := httptest.NewServer(handler.RegisterUser(authConfig))
	defer ts.Close()

	user := domain.User{
		Login:    "testuser",
		Password: "testpassword",
	}
	payload, _ := json.Marshal(user)

	//nolint:noctx // Do not need context here
	res, err := http.Post(ts.URL+"/register", "application/json", bytes.NewBuffer(payload))
	assert.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, http.StatusConflict, res.StatusCode)
	assert.Contains(t, readResponseBody(t, res), "User already registered")
}

func TestLoginUser_Success(t *testing.T) {
	mockRepo := &mocks.MockUsersRepository{
		GetUserByLoginFunc: func(login string) (*domain.DBUser, error) {
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("testpassword"), bcrypt.DefaultCost)
			return &domain.DBUser{
				ID:           1,
				Login:        "testuser",
				PasswordHash: string(hashedPassword),
			}, nil
		},
	}

	logger := zap.NewNop().Sugar()

	authConfig := &config.AuthConfig{
		JWTKey:       []byte("test-secret"),
		JWTExpiresIn: time.Hour,
	}
	handler := NewUsersHandler(logger, mockRepo)

	ts := httptest.NewServer(handler.LoginUser(authConfig))
	defer ts.Close()

	user := domain.User{
		Login:    "testuser",
		Password: "testpassword",
	}
	payload, _ := json.Marshal(user)

	//nolint:noctx // Do not need context here
	res, err := http.Post(ts.URL+"/login", "application/json", bytes.NewBuffer(payload))
	assert.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Contains(t, readResponseBody(t, res), "Login successful")

	cookies := res.Cookies()
	assert.NotEmpty(t, cookies)
	assert.Equal(t, "access_token", cookies[0].Name)
}

func TestLoginUser_InvalidCredentials(t *testing.T) {
	mockRepo := &mocks.MockUsersRepository{
		GetUserByLoginFunc: func(login string) (*domain.DBUser, error) {
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("testpassword"), bcrypt.DefaultCost)
			return &domain.DBUser{
				ID:           1,
				Login:        "testuser",
				PasswordHash: string(hashedPassword),
			}, nil
		},
	}

	logger := zap.NewNop().Sugar()

	authConfig := &config.AuthConfig{
		JWTKey:       []byte("test-secret"),
		JWTExpiresIn: time.Hour,
	}
	handler := NewUsersHandler(logger, mockRepo)

	ts := httptest.NewServer(handler.LoginUser(authConfig))
	defer ts.Close()

	user := domain.User{
		Login:    "testuser",
		Password: "wrongpassword",
	}
	payload, _ := json.Marshal(user)

	//nolint:noctx // Do not need context here
	res, err := http.Post(ts.URL+"/login", "application/json", bytes.NewBuffer(payload))
	assert.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
	assert.Contains(t, readResponseBody(t, res), "Invalid login or password")
}
