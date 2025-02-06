package mocks

import "github.com/frolmr/gophermart/internal/domain"

type MockUsersRepository struct {
	CreateAndReturnUserFunc func(login, password string) (*domain.DBUser, error)
	GetUserByLoginFunc      func(login string) (*domain.DBUser, error)
}

func (m *MockUsersRepository) CreateAndReturnUser(login, password string) (*domain.DBUser, error) {
	return m.CreateAndReturnUserFunc(login, password)
}

func (m *MockUsersRepository) GetUserByLogin(login string) (*domain.DBUser, error) {
	return m.GetUserByLoginFunc(login)
}
