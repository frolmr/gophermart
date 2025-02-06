package handlers

import (
	"bytes"
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

func TestRegisterWithdrawal_Success(t *testing.T) {
	mockRepo := &mocks.MockWithdrawalRepository{
		GetUserCurrentBalanceFunc: func(userID int) (float64, error) {
			return 100.0, nil
		},
		CreateWithdrawalFunc: func(orderNumber string, sum float64, userID int) error {
			return nil
		},
	}

	logger := zap.NewNop().Sugar()
	handler := NewWithdrawalsHandler(logger, mockRepo)

	withdrawal := domain.Withdrawal{
		Order: "12345678903",
		Sum:   50.0,
	}
	payload, _ := json.Marshal(withdrawal)
	req := httptest.NewRequest(http.MethodPost, "/withdrawals", bytes.NewBuffer(payload))
	userID := 123
	req.Header.Set(domain.UserIDHeader, strconv.Itoa(userID))

	rr := httptest.NewRecorder()

	handler.RegisterWithdrawal(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRegisterWithdrawal_InvalidOrderNumber(t *testing.T) {
	mockRepo := &mocks.MockWithdrawalRepository{}

	logger := zap.NewNop().Sugar()
	handler := NewWithdrawalsHandler(logger, mockRepo)

	withdrawal := domain.Withdrawal{
		Order: "12345678902",
		Sum:   50.0,
	}
	payload, _ := json.Marshal(withdrawal)
	req := httptest.NewRequest(http.MethodPost, "/withdrawals", bytes.NewBuffer(payload))
	userID := 123
	req.Header.Set(domain.UserIDHeader, strconv.Itoa(userID))

	rr := httptest.NewRecorder()

	handler.RegisterWithdrawal(rr, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	assert.Contains(t, rr.Body.String(), "Order number is invalid")
}

func TestRegisterWithdrawal_InsufficientFunds(t *testing.T) {
	mockRepo := &mocks.MockWithdrawalRepository{
		GetUserCurrentBalanceFunc: func(userID int) (float64, error) {
			return 30.0, nil
		},
	}

	logger := zap.NewNop().Sugar()
	handler := NewWithdrawalsHandler(logger, mockRepo)

	withdrawal := domain.Withdrawal{
		Order: "12345678903",
		Sum:   50.0,
	}
	payload, _ := json.Marshal(withdrawal)
	req := httptest.NewRequest(http.MethodPost, "/withdrawals", bytes.NewBuffer(payload))
	userID := 123
	req.Header.Set(domain.UserIDHeader, strconv.Itoa(userID))

	rr := httptest.NewRecorder()

	handler.RegisterWithdrawal(rr, req)

	assert.Equal(t, http.StatusPaymentRequired, rr.Code)
	assert.Contains(t, rr.Body.String(), "Not enough funds")
}

func TestGetWithdrawals_Success(t *testing.T) {
	mockRepo := &mocks.MockWithdrawalRepository{
		GetAllUserWithdrawalsFunc: func(userID int) ([]*domain.Withdrawal, error) {
			return []*domain.Withdrawal{
				{Order: "12345678903", Sum: 50.0},
				{Order: "98765432109", Sum: 30.0},
			}, nil
		},
	}

	logger := zap.NewNop().Sugar()
	handler := NewWithdrawalsHandler(logger, mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/withdrawals", nil)
	userID := 123
	req.Header.Set(domain.UserIDHeader, strconv.Itoa(userID))

	rr := httptest.NewRecorder()

	handler.GetWithdrawals(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var withdrawals []*domain.Withdrawal
	err := json.NewDecoder(rr.Body).Decode(&withdrawals)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(withdrawals))
}

func TestGetWithdrawals_NoWithdrawals(t *testing.T) {
	mockRepo := &mocks.MockWithdrawalRepository{
		GetAllUserWithdrawalsFunc: func(userID int) ([]*domain.Withdrawal, error) {
			return []*domain.Withdrawal{}, nil
		},
	}

	logger := zap.NewNop().Sugar()
	handler := NewWithdrawalsHandler(logger, mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/withdrawals", nil)
	userID := 123
	req.Header.Set(domain.UserIDHeader, strconv.Itoa(userID))

	rr := httptest.NewRecorder()

	handler.GetWithdrawals(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}
