package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/frolmr/gophermart/internal/domain"
	"github.com/frolmr/gophermart/internal/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestBalancesHandler_GetBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockBalanceRepository(ctrl)
	logger := zap.NewNop().Sugar()
	handler := NewBalancesHandler(logger, mockRepo)

	tests := []struct {
		name           string
		userID         string
		mockSetup      func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Successful balance retrieval",
			userID: "1",
			mockSetup: func() {
				mockRepo.EXPECT().
					GetUserCurrentBalance(int64(1)).
					Return(100.5, nil)

				mockRepo.EXPECT().
					GetUserWithdrawalsSum(int64(1)).
					Return(50.25, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"current":100.5,"withdrawn":50.25}`,
		},
		{
			name:   "Failed to get user current balance",
			userID: "1",
			mockSetup: func() {
				mockRepo.EXPECT().
					GetUserCurrentBalance(int64(1)).
					Return(0.0, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Failed to get user accrual sum"}`,
		},
		{
			name:   "Failed to get user withdrawals sum",
			userID: "1",
			mockSetup: func() {
				mockRepo.EXPECT().
					GetUserCurrentBalance(int64(1)).
					Return(100.5, nil)

				mockRepo.EXPECT().
					GetUserWithdrawalsSum(int64(1)).
					Return(0.0, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Failed to get user withdrawal sum"}`,
		},
		{
			name:           "Invalid user ID",
			userID:         "invalid",
			mockSetup:      func() {},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Invalid user id"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest(http.MethodGet, "/balance", nil)
			req.Header.Set(domain.UserIDHeader, tt.userID)
			w := httptest.NewRecorder()

			handler.GetBalance(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			}
		})
	}
}
