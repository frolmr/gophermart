package mocks

import "github.com/frolmr/gophermart/internal/domain"

type MockOrdersRepository struct {
	FindOrderByNumberFunc        func(number string) (*domain.DBOrder, error)
	GetAllUserOrdersFunc         func(userID int) ([]*domain.Order, error)
	CreateOrderFunc              func(number string, userID int) error
	GetAllUnprocessedOrdersFunc  func() ([]*domain.DBOrder, error)
	UpdateOrderAccrualStatusFunc func(id int, status string, accrual *float64) error
}

func (m *MockOrdersRepository) FindOrderByNumber(number string) (*domain.DBOrder, error) {
	return m.FindOrderByNumberFunc(number)
}

func (m *MockOrdersRepository) GetAllUserOrders(userID int) ([]*domain.Order, error) {
	return m.GetAllUserOrdersFunc(userID)
}

func (m *MockOrdersRepository) CreateOrder(number string, userID int) error {
	return m.CreateOrderFunc(number, userID)
}

func (m *MockOrdersRepository) GetAllUnprocessedOrders() ([]*domain.DBOrder, error) {
	return m.GetAllUnprocessedOrdersFunc()
}

func (m *MockOrdersRepository) UpdateOrderAccrualStatus(id int, status string, accrual *float64) error {
	return m.UpdateOrderAccrualStatusFunc(id, status, accrual)
}
