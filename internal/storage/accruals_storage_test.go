package storage

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/frolmr/gophermart/pkg/formatter"
	"github.com/stretchr/testify/assert"
)

func TestCreateOrderAccrual_Success(t *testing.T) {
	storage, mock := NewMockStorage(t)

	orderID := 1
	value := 50.0
	accrualInSubunit := formatter.ConvertToSubunit(value)

	mock.ExpectExec("INSERT INTO accruals").
		WithArgs(orderID, accrualInSubunit).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := storage.CreateOrderAccrual(orderID, value)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateOrderAccrual_DatabaseError(t *testing.T) {
	storage, mock := NewMockStorage(t)

	orderID := 1
	value := 50.0
	accrualInSubunit := formatter.ConvertToSubunit(value)

	mock.ExpectExec("INSERT INTO accruals").
		WithArgs(orderID, accrualInSubunit).
		WillReturnError(errors.New("database error"))

	err := storage.CreateOrderAccrual(orderID, value)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
func TestGetUserAccrualsSum_Success(t *testing.T) {
	storage, mock := NewMockStorage(t)

	userID := 1
	totalAccruals := 10000

	rows := sqlmock.NewRows([]string{"total_accruals"}).
		AddRow(totalAccruals)

	mock.ExpectQuery(`
			SELECT COALESCE\(SUM\(accrual\), 0\) AS total_accruals
			FROM accruals a
            LEFT JOIN orders o ON a\.order_id = o\.id
			WHERE o\.user_id = \$1`).
		WithArgs(userID).
		WillReturnRows(rows)

	result, err := storage.GetUserAccrualsSum(userID)

	assert.NoError(t, err)
	assert.Equal(t, formatter.ConvertToCurrency(totalAccruals), result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserAccrualsSum_NoData(t *testing.T) {
	storage, mock := NewMockStorage(t)

	userID := 1
	totalAccruals := 0

	rows := sqlmock.NewRows([]string{"total_accruals"}).
		AddRow(totalAccruals)

	mock.ExpectQuery(`
			SELECT COALESCE\(SUM\(accrual\), 0\) AS total_accruals
			FROM accruals a
            LEFT JOIN orders o ON a\.order_id = o\.id
			WHERE o\.user_id = \$1`).
		WithArgs(userID).
		WillReturnRows(rows)

	result, err := storage.GetUserAccrualsSum(userID)

	assert.NoError(t, err)
	assert.Equal(t, formatter.ConvertToCurrency(totalAccruals), result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserAccrualsSum_DatabaseError(t *testing.T) {
	storage, mock := NewMockStorage(t)

	userID := 1

	mock.ExpectQuery(`
			SELECT COALESCE\(SUM\(accrual\), 0\) AS total_accruals
			FROM accruals a
            LEFT JOIN orders o ON a\.order_id = o\.id
			WHERE o\.user_id = \$1`).
		WithArgs(userID).
		WillReturnError(errors.New("database error"))

	result, err := storage.GetUserAccrualsSum(userID)

	assert.Error(t, err)
	assert.Equal(t, 0.0, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}
