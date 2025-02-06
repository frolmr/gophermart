package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/frolmr/gophermart/internal/domain"
	"github.com/frolmr/gophermart/internal/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestGetBalance(t *testing.T) {
	mockRepo := &mocks.MockBalanceRepository{
		GetUserCurrentBalanceFunc: func(userID int) (float64, error) {
			return 100.50, nil
		},
		GetUserWithdrawalsSumFunc: func(userID int) (float64, error) {
			return 30.25, nil
		},
	}

	logger := zap.NewNop().Sugar()

	handler := NewBalancesHandler(logger, mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/balance", nil)
	userID := 123
	req.Header.Set(domain.UserIDHeader, strconv.Itoa(userID))

	rr := httptest.NewRecorder()

	handler.GetBalance(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var balance domain.Balance
	err := json.NewDecoder(rr.Body).Decode(&balance)
	assert.NoError(t, err)

	assert.Equal(t, 100.50, balance.BalanceSum)
	assert.Equal(t, 30.25, balance.WithdrawalSum)
}

func TestGetBalance_InvalidUserID(t *testing.T) {
	mockRepo := &mocks.MockBalanceRepository{}

	logger := zap.NewNop().Sugar()

	handler := NewBalancesHandler(logger, mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/balance", nil)
	req.Header.Set(domain.UserIDHeader, "invalid")

	rr := httptest.NewRecorder()

	handler.GetBalance(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestGetBalance_RepositoryError(t *testing.T) {
	mockRepo := &mocks.MockBalanceRepository{
		GetUserCurrentBalanceFunc: func(userID int) (float64, error) {
			return 0, assert.AnError
		},
	}

	logger := zap.NewNop().Sugar()

	handler := NewBalancesHandler(logger, mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/balance", nil)
	userID := 123
	req.Header.Set(domain.UserIDHeader, strconv.Itoa(userID))

	rr := httptest.NewRecorder()

	handler.GetBalance(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}
