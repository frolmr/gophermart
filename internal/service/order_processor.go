package service

import (
	"errors"
	"sync"
	"time"

	"github.com/frolmr/gophermart/internal/client"
	"github.com/frolmr/gophermart/internal/domain"
	"go.uber.org/zap"
)

const (
	processingInterval = 1 * time.Second // NOTE: original value was 5 minutes
)

type OrdersRepository interface {
	GetAllUnprocessedOrders() ([]*domain.DBOrder, error)
	UpdateOrderAccrualStatus(id int, status string, accrual *float64) error
}

type OrderProcessor struct {
	logger *zap.SugaredLogger
	repo   OrdersRepository
	client *client.AccrualClient
}

func NewOrderProcessor(lgr *zap.SugaredLogger, repo OrdersRepository, client *client.AccrualClient) *OrderProcessor {
	return &OrderProcessor{
		logger: lgr,
		repo:   repo,
		client: client,
	}
}

func (op *OrderProcessor) Run(stopCh <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(processingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := op.processUnprocessedOrders()
			if err != nil {
				continue
			}
		case <-stopCh:
			op.logger.Info("Shutting down Orders Processor")
			ticker.Stop()
			return
		}
	}
}

func (op *OrderProcessor) processUnprocessedOrders() error {
	ordersToProcess, err := op.repo.GetAllUnprocessedOrders()
	if err != nil {
		return err
	}

	if len(ordersToProcess) > 0 {
		op.logger.Infof("Order Processor: Found %d orders to process", len(ordersToProcess))

		for _, order := range ordersToProcess {
			op.logger.Info("Processing order ", order.Number)
			if err := op.processOrder(order); err != nil {
				return err
			}
		}
	} else {
		op.logger.Info("Order Processor: No orders to process")
	}
	return nil
}

func (op *OrderProcessor) processOrder(order *domain.DBOrder) error {
	accrualOrder, retryAfter, err := op.client.RequestOrderState(order.Number)
	if errors.Is(err, client.ErrTooManyRequests) {
		time.Sleep(retryAfter)
	} else if err != nil {
		return err
	}

	if accrualOrder == nil {
		return nil
	}

	if order.Status != accrualOrder.Status {
		if err := op.repo.UpdateOrderAccrualStatus(order.ID, accrualOrder.Status, &accrualOrder.Accrual); err != nil {
			return err
		}
	}

	return nil
}
