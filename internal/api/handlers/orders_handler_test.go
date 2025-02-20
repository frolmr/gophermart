package handlers

import (
	"bytes"
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

func TestOrdersHandler_LoadOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockOrdersRepository(ctrl)
	logger := zap.NewNop().Sugar()
	handler := NewOrdersHandler(logger, mockRepo)

	tests := []struct {
		name           string
		orderNumber    string
		userID         string
		mockSetup      func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "Successful order upload",
			orderNumber: "12345678903",
			userID:      "1",
			mockSetup: func() {
				mockRepo.EXPECT().
					FindOrderByNumber("12345678903").
					Return(nil, nil)

				mockRepo.EXPECT().
					CreateOrder("12345678903", int64(1)).
					Return(nil)
			},
			expectedStatus: http.StatusAccepted,
			expectedBody:   "Order uploaded",
		},
		{
			name:        "Order already uploaded by the same user",
			orderNumber: "12345678903",
			userID:      "1",
			mockSetup: func() {
				mockRepo.EXPECT().
					FindOrderByNumber("12345678903").
					Return(&domain.DBOrder{UserID: 1}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name:        "Order already uploaded by another user",
			orderNumber: "12345678903",
			userID:      "2",
			mockSetup: func() {
				mockRepo.EXPECT().
					FindOrderByNumber("12345678903").
					Return(&domain.DBOrder{UserID: 1}, nil)
			},
			expectedStatus: http.StatusConflict,
			expectedBody:   "Already downloaded",
		},
		{
			name:           "Invalid order number (Luhn check fails)",
			orderNumber:    "12345678902",
			userID:         "1",
			mockSetup:      func() {},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   "Order number is invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader([]byte(tt.orderNumber)))
			req.Header.Set(domain.UserIDHeader, tt.userID)
			w := httptest.NewRecorder()

			handler.LoadOrder(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, w.Body.String(), tt.expectedBody)
			}
		})
	}
}

func TestOrdersHandler_GetOrders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockOrdersRepository(ctrl)
	logger := zap.NewNop().Sugar()
	handler := NewOrdersHandler(logger, mockRepo)

	tests := []struct {
		name           string
		userID         string
		mockSetup      func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Successful retrieval of orders",
			userID: "1",
			mockSetup: func() {
				mockRepo.EXPECT().
					GetAllUserOrders(int64(1)).
					Return([]*domain.Order{
						{Number: "12345678903", Status: "PROCESSED", UploadedAt: time.Now()},
						{Number: "98765432103", Status: "NEW", UploadedAt: time.Now()},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"number":"12345678903","status":"PROCESSED","uploaded_at":"`,
		},
		{
			name:   "No orders found",
			userID: "1",
			mockSetup: func() {
				mockRepo.EXPECT().
					GetAllUserOrders(int64(1)).
					Return([]*domain.Order{}, nil)
			},
			expectedStatus: http.StatusNoContent,
			expectedBody:   "",
		},
		{
			name:   "Invalid user ID",
			userID: "invalid",
			mockSetup: func() {
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Invalid user id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest(http.MethodGet, "/orders", nil)
			req.Header.Set(domain.UserIDHeader, tt.userID)
			w := httptest.NewRecorder()

			handler.GetOrders(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, w.Body.String(), tt.expectedBody)
			}
		})
	}
}
