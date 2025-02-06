package service

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/frolmr/gophermart/internal/client"
	"github.com/frolmr/gophermart/internal/config"
	"github.com/frolmr/gophermart/internal/domain"
	"github.com/frolmr/gophermart/internal/mocks"
	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestOrderProcessor_ProcessUnprocessedOrders_Success(t *testing.T) {
	mockRepo := &mocks.MockOrdersRepository{
		GetAllUnprocessedOrdersFunc: func() ([]*domain.DBOrder, error) {
			return []*domain.DBOrder{
				{ID: 1, Number: "12345678903", Status: "NEW"},
				{ID: 2, Number: "98765432109", Status: "PROCESSING"},
			}, nil
		},
		UpdateOrderAccrualStatusFunc: func(id int, status string, accrual *float64) error {
			return nil
		},
	}

	httpClient := resty.New()
	httpmock.ActivateNonDefault(httpClient.GetClient())

	conf := &config.AppConfig{AccrualSystemAddress: "http://accrual-system"}
	logger := zap.NewNop().Sugar()
	client := client.NewAccrualClient(httpClient, conf, logger)

	processor := NewOrderProcessor(logger, mockRepo, client)

	responderFirst, _ := httpmock.NewJsonResponder(http.StatusOK, json.RawMessage(`{"order": "12345678903", "status": "PROCESSED", "accrual": 10.5}`))
	httpmock.RegisterResponder("GET", "http://accrual-system/api/orders/12345678903", responderFirst)
	defer httpmock.DeactivateAndReset()

	responderSecond, _ := httpmock.NewJsonResponder(http.StatusOK, json.RawMessage(`{"order": "98765432109", "status": "PROCESSED", "accrual": 10.5}`))
	httpmock.RegisterResponder("GET", "http://accrual-system/api/orders/98765432109", responderSecond)
	defer httpmock.DeactivateAndReset()

	err := processor.processUnprocessedOrders()
	assert.NoError(t, err)
}

func TestOrderProcessor_ProcessUnprocessedOrders_NoOrders(t *testing.T) {
	mockRepo := &mocks.MockOrdersRepository{
		GetAllUnprocessedOrdersFunc: func() ([]*domain.DBOrder, error) {
			return []*domain.DBOrder{}, nil
		},
	}

	httpClient := resty.New()
	httpmock.ActivateNonDefault(httpClient.GetClient())
	defer httpmock.DeactivateAndReset()

	conf := &config.AppConfig{AccrualSystemAddress: "http://accrual-system"}
	logger := zap.NewNop().Sugar()
	client := client.NewAccrualClient(httpClient, conf, logger)

	processor := NewOrderProcessor(logger, mockRepo, client)

	err := processor.processUnprocessedOrders()
	assert.NoError(t, err)
}

func TestOrderProcessor_ProcessUnprocessedOrders_DatabaseError(t *testing.T) {
	mockRepo := &mocks.MockOrdersRepository{
		GetAllUnprocessedOrdersFunc: func() ([]*domain.DBOrder, error) {
			return nil, errors.New("database error")
		},
	}

	httpClient := resty.New()
	httpmock.ActivateNonDefault(httpClient.GetClient())
	defer httpmock.DeactivateAndReset()

	conf := &config.AppConfig{AccrualSystemAddress: "http://accrual-system"}
	logger := zap.NewNop().Sugar()
	client := client.NewAccrualClient(httpClient, conf, logger)

	processor := NewOrderProcessor(logger, mockRepo, client)

	err := processor.processUnprocessedOrders()
	assert.Error(t, err)
}

func TestOrderProcessor_ProcessOrder_TooManyRequests(t *testing.T) {
	mockRepo := &mocks.MockOrdersRepository{
		UpdateOrderAccrualStatusFunc: func(id int, status string, accrual *float64) error {
			return nil
		},
	}

	httpClient := resty.New()
	httpmock.ActivateNonDefault(httpClient.GetClient())
	defer httpmock.DeactivateAndReset()

	conf := &config.AppConfig{AccrualSystemAddress: "http://accrual-system"}
	logger := zap.NewNop().Sugar()
	client := client.NewAccrualClient(httpClient, conf, logger)

	headers := http.Header{}
	headers.Set("Retry-After", "1")

	responder := httpmock.NewStringResponder(http.StatusTooManyRequests, "")
	httpmock.RegisterResponder("GET", "http://accrual-system/api/orders/123", responder.HeaderSet(headers))

	processor := NewOrderProcessor(logger, mockRepo, client)

	order := &domain.DBOrder{ID: 1, Number: "123", Status: "NEW"}
	err := processor.processOrder(order)
	assert.NoError(t, err)
}
