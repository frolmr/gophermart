package handlers

import (
	"github.com/frolmr/gophermart/internal/storage"
	"go.uber.org/zap"
)

type RequestHandlers struct {
	UsersHandler       *UsersHandler
	OrdersHandler      *OrdersHandler
	WithdrawalsHandler *WithdrawalsHandler
	BalancesHandler    *BalancesHandler
}

func NewRequestHandlers(lgr *zap.SugaredLogger, stor *storage.Storage) *RequestHandlers {
	return &RequestHandlers{
		UsersHandler:       NewUsersHandler(lgr, stor),
		OrdersHandler:      NewOrdersHandler(lgr, stor),
		WithdrawalsHandler: NewWithdrawalsHandler(lgr, stor),
		BalancesHandler:    NewBalancesHandler(lgr, stor),
	}
}
