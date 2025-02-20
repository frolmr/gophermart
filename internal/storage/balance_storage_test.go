package storage

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/frolmr/gophermart/pkg/formatter"
	"github.com/stretchr/testify/assert"
)

func TestGetUserCurrentBalance_Success(t *testing.T) {
	storage, mock := NewMockStorage(t)

	userID := int64(1)
	totalAccruals := int64(10000)
	totalWithdrawals := int64(3000)
	netDifference := totalAccruals - totalWithdrawals

	rows := sqlmock.NewRows([]string{"net_difference"}).
		AddRow(netDifference)

	mock.ExpectQuery(`
            WITH accruals_cte AS \(
                SELECT COALESCE\(SUM\(a\.accrual\), 0\) AS total_accruals
                FROM accruals a
                LEFT JOIN orders o ON a\.order_id = o\.id
                WHERE o\.user_id = \$1
            \),
            withdrawals_cte AS \(
                SELECT COALESCE\(SUM\(w\.sum\), 0\) AS total_withdrawals
                FROM withdrawals w
                WHERE w\.user_id = \$1
            \)
            SELECT
                accruals_cte\.total_accruals - withdrawals_cte\.total_withdrawals AS net_difference
            FROM
                accruals_cte, withdrawals_cte;`).
		WithArgs(userID).
		WillReturnRows(rows)

	result, err := storage.GetUserCurrentBalance(userID)

	assert.NoError(t, err)
	assert.Equal(t, formatter.ConvertToCurrency(netDifference), result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserCurrentBalance_NoData(t *testing.T) {
	storage, mock := NewMockStorage(t)

	userID := int64(1)
	netDifference := int64(0)

	rows := sqlmock.NewRows([]string{"net_difference"}).
		AddRow(netDifference)

	mock.ExpectQuery(`
            WITH accruals_cte AS \(
                SELECT COALESCE\(SUM\(a\.accrual\), 0\) AS total_accruals
                FROM accruals a
                LEFT JOIN orders o ON a\.order_id = o\.id
                WHERE o\.user_id = \$1
            \),
            withdrawals_cte AS \(
                SELECT COALESCE\(SUM\(w\.sum\), 0\) AS total_withdrawals
                FROM withdrawals w
                WHERE w\.user_id = \$1
            \)
            SELECT
                accruals_cte\.total_accruals - withdrawals_cte\.total_withdrawals AS net_difference
            FROM
                accruals_cte, withdrawals_cte;`).
		WithArgs(userID).
		WillReturnRows(rows)

	result, err := storage.GetUserCurrentBalance(userID)

	assert.NoError(t, err)
	assert.Equal(t, formatter.ConvertToCurrency(netDifference), result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserCurrentBalance_DatabaseError(t *testing.T) {
	storage, mock := NewMockStorage(t)

	userID := int64(1)

	mock.ExpectQuery(`
            WITH accruals_cte AS \(
                SELECT COALESCE\(SUM\(a\.accrual\), 0\) AS total_accruals
                FROM accruals a
                LEFT JOIN orders o ON a\.order_id = o\.id
                WHERE o\.user_id = \$1
            \),
            withdrawals_cte AS \(
                SELECT COALESCE\(SUM\(w\.sum\), 0\) AS total_withdrawals
                FROM withdrawals w
                WHERE w\.user_id = \$1
            \)
            SELECT
                accruals_cte\.total_accruals - withdrawals_cte\.total_withdrawals AS net_difference
            FROM
                accruals_cte, withdrawals_cte;`).
		WithArgs(userID).
		WillReturnError(errors.New("database error"))

	result, err := storage.GetUserCurrentBalance(userID)

	assert.Error(t, err)
	assert.Equal(t, 0.0, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}
