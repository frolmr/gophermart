package storage

import (
	"fmt"

	"github.com/frolmr/gophermart/pkg/formatter"
)

func (s *Storage) CreateOrderAccrual(orderID int64, value float64) error {
	accrualInSubunit := formatter.ConvertToSubunit(value)
	if _, err := s.db.Exec("INSERT INTO accruals (order_id, accrual) VALUES ($1, $2)", orderID, accrualInSubunit); err != nil {
		s.logger.Errorf("Accrual insert fail for order_id: %s, value %f; err: %s", orderID, value, err.Error())
		return fmt.Errorf("error creating accrual: %w", err)
	}

	return nil
}

func (s *Storage) GetUserAccrualsSum(userID int64) (float64, error) {
	query := `
			SELECT COALESCE(SUM(accrual), 0) AS total_accruals
			FROM accruals a
            LEFT JOIN orders o ON a.order_id = o.id
			WHERE o.user_id = $1`

	var totalAccruals int64
	err := s.db.QueryRow(query, userID).Scan(&totalAccruals)
	if err != nil {
		s.logger.Errorf("Accrual sum selection fail for user_id: %s, err: %s", userID, err.Error())
		return 0, fmt.Errorf("error getting accrual sum: %w", err)
	}

	return formatter.ConvertToCurrency(totalAccruals), nil
}
