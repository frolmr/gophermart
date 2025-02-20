package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/frolmr/gophermart/internal/domain"
	"github.com/frolmr/gophermart/internal/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestWithdrawalsHandler_RegisterWithdrawal(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockWithdrawalRepository(ctrl)
	logger := zap.NewNop().Sugar()
	handler := NewWithdrawalsHandler(logger, mockRepo)

	tests := []struct {
		name           string
		withdrawal     domain.Withdrawal
		userID         string
		mockSetup      func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Successful withdrawal registration",
			withdrawal: domain.Withdrawal{
				Order: "12345678903",
				Sum:   50.0,
			},
			userID: "1",
			mockSetup: func() {
				mockRepo.EXPECT().
					GetUserCurrentBalance(int64(1)).
					Return(100.0, nil)

				mockRepo.EXPECT().
					CreateWithdrawal("12345678903", 50.0, int64(1)).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name: "Invalid order number (Luhn check fails)",
			withdrawal: domain.Withdrawal{
				Order: "12345678902",
				Sum:   50.0,
			},
			userID:         "1",
			mockSetup:      func() {},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   "Order number is invalid",
		},
		{
			name: "Not enough funds",
			withdrawal: domain.Withdrawal{
				Order: "12345678903",
				Sum:   150.0,
			},
			userID: "1",
			mockSetup: func() {
				mockRepo.EXPECT().
					GetUserCurrentBalance(int64(1)).
					Return(100.0, nil)
			},
			expectedStatus: http.StatusPaymentRequired,
			expectedBody:   "Not enough funds",
		},
		{
			name: "Invalid user ID",
			withdrawal: domain.Withdrawal{
				Order: "12345678903",
				Sum:   50.0,
			},
			userID:         "invalid",
			mockSetup:      func() {},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Invalid user id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			body, _ := json.Marshal(tt.withdrawal)
			req := httptest.NewRequest(http.MethodPost, "/withdrawals", bytes.NewReader(body))
			req.Header.Set(domain.UserIDHeader, tt.userID)
			w := httptest.NewRecorder()

			handler.RegisterWithdrawal(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, w.Body.String(), tt.expectedBody)
			}
		})
	}
}

func TestWithdrawalsHandler_GetWithdrawals(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockWithdrawalRepository(ctrl)
	logger := zap.NewNop().Sugar()
	handler := NewWithdrawalsHandler(logger, mockRepo)

	tests := []struct {
		name           string
		userID         string
		mockSetup      func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Successful retrieval of withdrawals",
			userID: "1",
			mockSetup: func() {
				mockRepo.EXPECT().
					GetAllUserWithdrawals(int64(1)).
					Return([]*domain.Withdrawal{
						{Order: "12345678903", Sum: 50.0, ProcessedAt: time.Now()},
						{Order: "98765432103", Sum: 30.0, ProcessedAt: time.Now()},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"order":"12345678903","sum":50,"processed_at":"`,
		},
		{
			name:   "No withdrawals found",
			userID: "1",
			mockSetup: func() {
				mockRepo.EXPECT().
					GetAllUserWithdrawals(int64(1)).
					Return([]*domain.Withdrawal{}, nil)
			},
			expectedStatus: http.StatusNoContent,
			expectedBody:   "",
		},
		{
			name:           "Invalid user ID",
			userID:         "invalid",
			mockSetup:      func() {},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Invalid user id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest(http.MethodGet, "/withdrawals", nil)
			req.Header.Set(domain.UserIDHeader, tt.userID)
			w := httptest.NewRecorder()

			handler.GetWithdrawals(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, w.Body.String(), tt.expectedBody)
			}
		})
	}
}
