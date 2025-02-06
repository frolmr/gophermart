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

func TestLoadOrder_Success(t *testing.T) {
	mockRepo := &mocks.MockOrdersRepository{
		FindOrderByNumberFunc: func(number string) (*domain.DBOrder, error) {
			return nil, nil
		},
		CreateOrderFunc: func(number string, userID int) error {
			return nil
		},
	}

	logger := zap.NewNop().Sugar()
	handler := NewOrdersHandler(logger, mockRepo)

	orderNumber := "12345678903"
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBufferString(orderNumber))
	userID := 123
	req.Header.Set(domain.UserIDHeader, strconv.Itoa(userID))

	rr := httptest.NewRecorder()

	handler.LoadOrder(rr, req)

	assert.Equal(t, http.StatusAccepted, rr.Code)
	assert.Contains(t, rr.Body.String(), "Order uploaded")
}

func TestLoadOrder_InvalidOrderNumber(t *testing.T) {
	mockRepo := &mocks.MockOrdersRepository{}

	logger := zap.NewNop().Sugar()
	handler := NewOrdersHandler(logger, mockRepo)

	orderNumber := "12345678902"
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBufferString(orderNumber))
	userID := 123
	req.Header.Set(domain.UserIDHeader, strconv.Itoa(userID))

	rr := httptest.NewRecorder()

	handler.LoadOrder(rr, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	assert.Contains(t, rr.Body.String(), "Order number is invalid")
}

func TestLoadOrder_OrderAlreadyExists(t *testing.T) {
	mockRepo := &mocks.MockOrdersRepository{
		FindOrderByNumberFunc: func(number string) (*domain.DBOrder, error) {
			return &domain.DBOrder{
				Number: "12345678903",
				UserID: 123,
			}, nil
		},
	}

	logger := zap.NewNop().Sugar()
	handler := NewOrdersHandler(logger, mockRepo)

	orderNumber := "12345678903"
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBufferString(orderNumber))
	userID := 123
	req.Header.Set(domain.UserIDHeader, strconv.Itoa(userID))

	rr := httptest.NewRecorder()

	handler.LoadOrder(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestLoadOrder_OrderAlreadyExistsForAnotherUser(t *testing.T) {
	mockRepo := &mocks.MockOrdersRepository{
		FindOrderByNumberFunc: func(number string) (*domain.DBOrder, error) {
			return &domain.DBOrder{
				Number: "12345678903",
				UserID: 456,
			}, nil
		},
	}

	logger := zap.NewNop().Sugar()
	handler := NewOrdersHandler(logger, mockRepo)

	orderNumber := "12345678903"
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBufferString(orderNumber))
	userID := 123
	req.Header.Set(domain.UserIDHeader, strconv.Itoa(userID))

	rr := httptest.NewRecorder()

	handler.LoadOrder(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
	assert.Contains(t, rr.Body.String(), "Already downloaded")
}

func TestGetOrders_Success(t *testing.T) {
	mockRepo := &mocks.MockOrdersRepository{
		GetAllUserOrdersFunc: func(userID int) ([]*domain.Order, error) {
			return []*domain.Order{
				{Number: "12345678903", Status: "PROCESSED"},
				{Number: "98765432109", Status: "NEW"},
			}, nil
		},
	}

	logger := zap.NewNop().Sugar()
	handler := NewOrdersHandler(logger, mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/orders", nil)
	userID := 123
	req.Header.Set(domain.UserIDHeader, strconv.Itoa(userID))

	rr := httptest.NewRecorder()

	handler.GetOrders(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var orders []*domain.Order
	err := json.NewDecoder(rr.Body).Decode(&orders)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(orders))
}

func TestGetOrders_NoOrders(t *testing.T) {
	mockRepo := &mocks.MockOrdersRepository{
		GetAllUserOrdersFunc: func(userID int) ([]*domain.Order, error) {
			return []*domain.Order{}, nil
		},
	}

	logger := zap.NewNop().Sugar()
	handler := NewOrdersHandler(logger, mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/orders", nil)
	userID := 123
	req.Header.Set(domain.UserIDHeader, strconv.Itoa(userID))

	rr := httptest.NewRecorder()

	handler.GetOrders(rr, req)
	assert.Equal(t, http.StatusNoContent, rr.Code)
}
