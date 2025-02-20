package client

import (
	"errors"
	"fmt"
	"net/http"

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
	RequestOrderState(number string) (*domain.AccrualOrder, error)
}

func NewAccrualClient(httpClient *resty.Client, conf *config.AppConfig, lgr *zap.SugaredLogger) *AccrualClient {
	return &AccrualClient{
		httpClient: httpClient,
		config:     conf,
		logger:     lgr,
	}
}

func (ac AccrualClient) RequestOrderState(number string) (*domain.AccrualOrder, error) {
	resp, err := ac.httpClient.R().
		SetResult(&domain.AccrualOrder{}).
		Get(ac.config.AccrualSystemAddress + "/api/orders/" + number)

	if err != nil {
		errMessage := "error sending request to accrual system: %w"
		ac.logger.Errorf(errMessage, err.Error())
		return nil, fmt.Errorf(errMessage, err)
	}

	if resp.StatusCode() == http.StatusOK {
		orderResp := resp.Result().(*domain.AccrualOrder)
		return orderResp, nil
	} else if resp.StatusCode() == http.StatusNoContent {
		return nil, nil
	} else {
		ac.logger.Error("Accrual system responsed error ", resp.StatusCode(), resp.Body())
		return nil, errors.New("Accrual system responded with status code: " + resp.Status())
	}
}
