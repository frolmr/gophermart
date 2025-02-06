package client

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/frolmr/gophermart/internal/config"
	"github.com/frolmr/gophermart/internal/domain"
	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func setupTest() (*AccrualClient, func()) {
	httpClient := resty.New()
	httpmock.ActivateNonDefault(httpClient.GetClient())

	conf := &config.AppConfig{AccrualSystemAddress: "http://accrual-system"}
	logger := zap.NewNop().Sugar()
	client := NewAccrualClient(httpClient, conf, logger)

	teardown := func() {
		httpmock.DeactivateAndReset()
	}

	return client, teardown
}

func TestRequestOrderState_Success(t *testing.T) {
	client, teardown := setupTest()
	defer teardown()

	responder, _ := httpmock.NewJsonResponder(http.StatusOK, json.RawMessage(`{"order": "123", "status": "PROCESSED", "accrual": 10.5}`))
	httpmock.RegisterResponder("GET", "http://accrual-system/api/orders/123", responder)

	order, retryAfter, err := client.RequestOrderState("123")

	assert.NoError(t, err)
	assert.Equal(t, domain.ZeroRetryAfter, retryAfter)
	assert.NotNil(t, order)
	assert.Equal(t, "123", order.Order)
	assert.Equal(t, "PROCESSED", order.Status)
	assert.Equal(t, 10.5, order.Accrual)
}

func TestRequestOrderState_NoContent(t *testing.T) {
	client, teardown := setupTest()
	defer teardown()

	responder := httpmock.NewStringResponder(http.StatusNoContent, "")
	httpmock.RegisterResponder("GET", "http://accrual-system/api/orders/123", responder)

	order, retryAfter, err := client.RequestOrderState("123")

	assert.NoError(t, err)
	assert.Equal(t, domain.ZeroRetryAfter, retryAfter)
	assert.Nil(t, order)
}

func TestRequestOrderState_TooManyRequests(t *testing.T) {
	client, teardown := setupTest()
	defer teardown()

	headers := http.Header{}
	headers.Set("Retry-After", "30")

	responder := httpmock.NewStringResponder(http.StatusTooManyRequests, "")
	httpmock.RegisterResponder("GET", "http://accrual-system/api/orders/123", responder.HeaderSet(headers))

	order, retryAfter, err := client.RequestOrderState("123")

	assert.Error(t, err)
	assert.Equal(t, ErrTooManyRequests, err)
	assert.Equal(t, 30*time.Second, retryAfter)
	assert.Nil(t, order)
}

func TestRequestOrderState_Error(t *testing.T) {
	client, teardown := setupTest()
	defer teardown()

	responder := httpmock.NewStringResponder(http.StatusInternalServerError, "")
	httpmock.RegisterResponder("GET", "http://accrual-system/api/orders/123", responder)

	order, retryAfter, err := client.RequestOrderState("123")

	assert.Error(t, err)
	assert.Equal(t, domain.ZeroRetryAfter, retryAfter)
	assert.Nil(t, order)
}
