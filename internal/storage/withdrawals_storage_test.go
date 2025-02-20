package storage

import (
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/frolmr/gophermart/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestCreateWithdrawal_Success(t *testing.T) {
	storage, mock := NewMockStorage(t)

	orderNumber := "12345678903"
	sum := 50.0
	userID := int64(1)
	sumInSubunit := int(sum * domain.ToSubunitDelimeter)

	mock.ExpectExec("INSERT INTO withdrawals").
		WithArgs(orderNumber, sumInSubunit, userID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := storage.CreateWithdrawal(orderNumber, sum, userID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateWithdrawal_DatabaseError(t *testing.T) {
	storage, mock := NewMockStorage(t)

	orderNumber := "12345678903"
	sum := 50.0
	userID := int64(1)
	sumInSubunit := int(sum * domain.ToSubunitDelimeter)

	mock.ExpectExec("INSERT INTO withdrawals").
		WithArgs(orderNumber, sumInSubunit, userID).
		WillReturnError(errors.New("database error"))

	err := storage.CreateWithdrawal(orderNumber, sum, userID)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAllUserWithdrawals_Success(t *testing.T) {
	storage, mock := NewMockStorage(t)

	userID := int64(1)
	processedAt := time.Now()
	withdrawals := []*domain.Withdrawal{
		{Order: "12345678903", Sum: 50.0, ProcessedAt: processedAt},
		{Order: "98765432109", Sum: 30.0, ProcessedAt: processedAt},
	}

	rows := sqlmock.NewRows([]string{"order_number", "sum", "processed_at"}).
		AddRow(withdrawals[0].Order, int(withdrawals[0].Sum*domain.ToSubunitDelimeter), withdrawals[0].ProcessedAt).
		AddRow(withdrawals[1].Order, int(withdrawals[1].Sum*domain.ToSubunitDelimeter), withdrawals[1].ProcessedAt)

	mock.ExpectPrepare("SELECT order_number, sum, processed_at FROM withdrawals").
		ExpectQuery().
		WithArgs(userID).
		WillReturnRows(rows)

	result, err := storage.GetAllUserWithdrawals(userID)

	assert.NoError(t, err)
	assert.Equal(t, withdrawals, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAllUserWithdrawals_NoWithdrawals(t *testing.T) {
	storage, mock := NewMockStorage(t)

	userID := int64(1)

	rows := sqlmock.NewRows([]string{"order_number", "sum", "processed_at"})

	mock.ExpectPrepare("SELECT order_number, sum, processed_at FROM withdrawals").
		ExpectQuery().
		WithArgs(userID).
		WillReturnRows(rows)

	result, err := storage.GetAllUserWithdrawals(userID)

	assert.NoError(t, err)
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAllUserWithdrawals_DatabaseError(t *testing.T) {
	storage, mock := NewMockStorage(t)

	userID := int64(1)

	mock.ExpectPrepare("SELECT order_number, sum, processed_at FROM withdrawals").
		ExpectQuery().
		WithArgs(userID).
		WillReturnError(errors.New("database error"))

	result, err := storage.GetAllUserWithdrawals(userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserWithdrawalsSum_Success(t *testing.T) {
	storage, mock := NewMockStorage(t)

	userID := int64(1)
	totalWithdrawals := 80.0
	totalWithdrawalsInSubunit := int(totalWithdrawals * domain.ToSubunitDelimeter)

	mock.ExpectQuery("SELECT COALESCE\\(SUM\\(sum\\), 0\\) AS total_withdrawals FROM withdrawals").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"total_withdrawals"}).
			AddRow(totalWithdrawalsInSubunit))

	result, err := storage.GetUserWithdrawalsSum(userID)

	assert.NoError(t, err)
	assert.Equal(t, totalWithdrawals, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserWithdrawalsSum_DatabaseError(t *testing.T) {
	storage, mock := NewMockStorage(t)

	userID := int64(1)

	mock.ExpectQuery("SELECT COALESCE\\(SUM\\(sum\\), 0\\) AS total_withdrawals FROM withdrawals").
		WithArgs(userID).
		WillReturnError(errors.New("database error"))

	result, err := storage.GetUserWithdrawalsSum(userID)

	assert.Error(t, err)
	assert.Equal(t, 0.0, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}
