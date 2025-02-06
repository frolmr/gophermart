package client

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/frolmr/gophermart/internal/config"
	"github.com/frolmr/gophermart/internal/domain"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

type AccrualClient struct {
	httpClient *resty.Client
	config     *config.AppConfig
	logger     *zap.SugaredLogger
}

type AccrualClientInterface interface {
	RequestOrderState(number string) (*domain.AccrualOrder, time.Duration, error)
}

var (
	ErrTooManyRequests = errors.New("too many requests")
)

const (
	defaultRetryAfter = 60 * time.Second
)

func NewAccrualClient(httpClient *resty.Client, conf *config.AppConfig, lgr *zap.SugaredLogger) *AccrualClient {
	return &AccrualClient{
		httpClient: httpClient,
		config:     conf,
		logger:     lgr,
	}
}

func (ac *AccrualClient) RequestOrderState(number string) (*domain.AccrualOrder, time.Duration, error) {
	resp, err := ac.httpClient.R().
		SetResult(&domain.AccrualOrder{}).
		Get(ac.config.AccrualSystemAddress + "/api/orders/" + number)

	if err != nil {
		ac.logger.Error("Error sending request to accrual system: ", err.Error())
		return nil, domain.ZeroRetryAfter, err
	}

	if resp.StatusCode() == http.StatusOK {
		orderResp := resp.Result().(*domain.AccrualOrder)
		return orderResp, domain.ZeroRetryAfter, nil
	} else if resp.StatusCode() == http.StatusNoContent {
		return nil, domain.ZeroRetryAfter, nil
	} else {
		if resp.StatusCode() == http.StatusTooManyRequests {
			retryAfter := defaultRetryAfter
			retryAfterResp := resp.Header().Get("Retry-After")
			retryAfterSec, err := strconv.Atoi(retryAfterResp)
			if err == nil {
				retryAfter = time.Duration(retryAfterSec * int(time.Second))
			}

			return nil, retryAfter, ErrTooManyRequests
		} else {
			ac.logger.Error("Accrual system responsed error ", resp.StatusCode(), resp.Body())
			return nil, domain.ZeroRetryAfter, errors.New("Accrual system responded with status code: " + resp.Status())
		}
	}
}
