package storage

import "github.com/frolmr/gophermart/pkg/formatter"

func (s *Storage) GetUserCurrentBalance(userID int) (float64, error) {
	query := `
            WITH accruals_cte AS (
                SELECT COALESCE(SUM(a.accrual), 0) AS total_accruals
                FROM accruals a
                LEFT JOIN orders o ON a.order_id = o.id
                WHERE o.user_id = $1
            ),
            withdrawals_cte AS (
                SELECT COALESCE(SUM(w.sum), 0) AS total_withdrawals
                FROM withdrawals w
                WHERE w.user_id = $1
            )
            SELECT
                accruals_cte.total_accruals - withdrawals_cte.total_withdrawals AS net_difference
            FROM
                accruals_cte, withdrawals_cte;`

	var total int
	err := s.db.QueryRow(query, userID).Scan(&total)
	if err != nil {
		s.logger.Errorf("Balance calculation fail for user_id: %s, err: %s", userID, err.Error())
		return 0, err
	}

	return formatter.ConvertToCurrency(total), nil
}
