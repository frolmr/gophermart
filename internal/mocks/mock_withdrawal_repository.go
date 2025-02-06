package mocks

import "github.com/frolmr/gophermart/internal/domain"

type MockWithdrawalRepository struct {
	CreateWithdrawalFunc      func(orderNumber string, sum float64, userID int) error
	GetUserCurrentBalanceFunc func(userID int) (float64, error)
	GetAllUserWithdrawalsFunc func(userID int) ([]*domain.Withdrawal, error)
}

func (m *MockWithdrawalRepository) CreateWithdrawal(orderNumber string, sum float64, userID int) error {
	return m.CreateWithdrawalFunc(orderNumber, sum, userID)
}

func (m *MockWithdrawalRepository) GetUserCurrentBalance(userID int) (float64, error) {
	return m.GetUserCurrentBalanceFunc(userID)
}

func (m *MockWithdrawalRepository) GetAllUserWithdrawals(userID int) ([]*domain.Withdrawal, error) {
	return m.GetAllUserWithdrawalsFunc(userID)
}
