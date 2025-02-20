package storage

import (
	"fmt"

	"github.com/frolmr/gophermart/internal/domain"
	"github.com/frolmr/gophermart/pkg/formatter"
)

func (s *Storage) CreateWithdrawal(orderNumber string, sum float64, userID int64) error {
	sumInSubunit := formatter.ConvertToSubunit(sum)
	query := `INSERT INTO withdrawals (order_number, sum, user_id) VALUES ($1, $2, $3)`

	if _, err := s.db.Exec(query, orderNumber, sumInSubunit, userID); err != nil {
		s.logger.Errorf("Inserting order# %s, failed; err: %s ", orderNumber, err.Error())
		return fmt.Errorf("error creating withdrawal: %w", err)
	}

	return nil
}

func (s *Storage) GetAllUserWithdrawals(userID int64) ([]*domain.Withdrawal, error) {
	var withdrawals []*domain.Withdrawal

	query := `
            SELECT order_number, sum, processed_at
            FROM withdrawals
	        WHERE user_id = $1
            ORDER BY processed_at DESC`

	stmt, err := s.db.Prepare(query)
	if err != nil {
		s.logger.Errorf("Can't prepare statement for withdrawals selection for user_id: %s, err: %s ", userID, err.Error())
		return nil, fmt.Errorf("error getting withdrawals: %w", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(userID)
	if err != nil {
		s.logger.Errorf("Query for withdrawals for user_id: %s selection fialed, err: %s", userID, err.Error())
		return nil, fmt.Errorf("error getting withdrawals: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var withdrawal domain.Withdrawal
		err := rows.Scan(&withdrawal.Order, &withdrawal.Sum, &withdrawal.ProcessedAt)
		if err != nil {
			s.logger.Error("Error scanning withdrawals ", err.Error())
			return nil, fmt.Errorf("error getting withdrawals: %w", err)
		}
		withdrawal.Sum /= domain.ToSubunitDelimeter
		withdrawals = append(withdrawals, &withdrawal)
	}

	if err := rows.Err(); err != nil {
		s.logger.Errorf("Got rows.Err() for user_id: %d, err: %s", userID, err.Error())
		return nil, fmt.Errorf("error getting withdrawals: %w", err)
	}

	return withdrawals, nil
}

func (s *Storage) GetUserWithdrawalsSum(userID int64) (float64, error) {
	query := `
			SELECT COALESCE(SUM(sum), 0) AS total_withdrawals
			FROM withdrawals
			WHERE user_id = $1`

	var totalWithdrawals int64
	err := s.db.QueryRow(query, userID).Scan(&totalWithdrawals)
	if err != nil {
		s.logger.Errorf("Withdrawal sum selection fail for user_id: %s, err: %s", userID, err.Error())
		return 0, fmt.Errorf("error getting withdrawals sum: %w", err)
	}

	return formatter.ConvertToCurrency(totalWithdrawals), nil
}
