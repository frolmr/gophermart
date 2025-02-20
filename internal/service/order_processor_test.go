package service

import (
	"errors"
	"testing"

	"github.com/frolmr/gophermart/internal/domain"
	"github.com/frolmr/gophermart/internal/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestOrderProcessor_ProcessUnprocessedOrders_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockOrdersRepository(ctrl)
	mockClient := mocks.NewMockAccrualClientInterface(ctrl)

	logger := zap.NewNop().Sugar()
	processor := NewOrderProcessor(logger, mockRepo, mockClient)

	ordersToProcess := []*domain.DBOrder{
		{ID: 1, Number: "12345678903", Status: "NEW"},
		{ID: 2, Number: "98765432109", Status: "PROCESSING"},
	}

	mockRepo.EXPECT().
		GetAllUnprocessedOrders().
		Return(ordersToProcess, nil)

	mockClient.EXPECT().
		RequestOrderState("12345678903").
		Return(&domain.AccrualOrder{Order: "12345678903", Status: "PROCESSED", Accrual: 10.5}, nil)

	mockRepo.EXPECT().
		UpdateOrderAccrualStatus(int64(1), "PROCESSED", gomock.Any()).
		Return(nil)

	mockClient.EXPECT().
		RequestOrderState("98765432109").
		Return(&domain.AccrualOrder{Order: "98765432109", Status: "PROCESSED", Accrual: 10.5}, nil)

	mockRepo.EXPECT().
		UpdateOrderAccrualStatus(int64(2), "PROCESSED", gomock.Any()).
		Return(nil)

	err := processor.processUnprocessedOrders()

	assert.NoError(t, err)
}

func TestOrderProcessor_ProcessUnprocessedOrders_NoOrders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockOrdersRepository(ctrl)
	mockClient := mocks.NewMockAccrualClientInterface(ctrl)

	logger := zap.NewNop().Sugar()

	processor := NewOrderProcessor(logger, mockRepo, mockClient)

	mockRepo.EXPECT().
		GetAllUnprocessedOrders().
		Return([]*domain.DBOrder{}, nil)

	err := processor.processUnprocessedOrders()

	assert.NoError(t, err)
}

func TestOrderProcessor_ProcessUnprocessedOrders_DatabaseError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockOrdersRepository(ctrl)
	mockClient := mocks.NewMockAccrualClientInterface(ctrl)

	logger := zap.NewNop().Sugar()
	processor := NewOrderProcessor(logger, mockRepo, mockClient)

	mockRepo.EXPECT().
		GetAllUnprocessedOrders().
		Return(nil, errors.New("database error"))

	err := processor.processUnprocessedOrders()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}

func TestOrderProcessor_ProcessUnprocessedOrders_ClientError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockOrdersRepository(ctrl)
	mockClient := mocks.NewMockAccrualClientInterface(ctrl)

	logger := zap.NewNop().Sugar()

	processor := NewOrderProcessor(logger, mockRepo, mockClient)

	ordersToProcess := []*domain.DBOrder{
		{ID: 1, Number: "12345678903", Status: "NEW"},
	}

	mockRepo.EXPECT().
		GetAllUnprocessedOrders().
		Return(ordersToProcess, nil)

	mockClient.EXPECT().
		RequestOrderState("12345678903").
		Return(nil, errors.New("client error"))

	err := processor.processUnprocessedOrders()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "client error")
}
