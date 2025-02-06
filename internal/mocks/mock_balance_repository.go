package mocks

type MockBalanceRepository struct {
	GetUserCurrentBalanceFunc func(userID int) (float64, error)
	GetUserWithdrawalsSumFunc func(userID int) (float64, error)
}

func (m *MockBalanceRepository) GetUserCurrentBalance(userID int) (float64, error) {
	return m.GetUserCurrentBalanceFunc(userID)
}

func (m *MockBalanceRepository) GetUserWithdrawalsSum(userID int) (float64, error) {
	return m.GetUserWithdrawalsSumFunc(userID)
}
