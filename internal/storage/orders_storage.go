package storage

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/frolmr/gophermart/internal/domain"
	"github.com/frolmr/gophermart/pkg/formatter"
)

func (s *Storage) FindOrderByNumber(number string) (*domain.DBOrder, error) {
	stmt, err := s.db.Prepare("SELECT id, number, status, uploaded_at, user_id FROM orders WHERE number = $1")
	if err != nil {
		s.logger.Errorf("Can't prepare statement for order# %s, err: %s", number, err.Error())
		return nil, fmt.Errorf("error getting order: %w", err)
	}
	defer stmt.Close()

	var order domain.DBOrder
	err = stmt.QueryRow(number).Scan(&order.ID, &order.Number, &order.Status, &order.UploadedAt, &order.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		} else {
			s.logger.Errorf("Order query fails for order# %s, err: %s", number, err.Error())
			return nil, fmt.Errorf("error getting order: %w", err)
		}
	}

	return &order, nil
}

func (s *Storage) CreateOrder(number string, userID int64) error {
	if _, err := s.db.Exec("INSERT INTO orders (number, user_id) VALUES ($1, $2)", number, userID); err != nil {
		s.logger.Errorf("Order insert fail for order# %s, user_id: %d; err: %s", number, userID, err.Error())
		return fmt.Errorf("error creating order: %w", err)
	}

	return nil
}

func (s *Storage) GetAllUnprocessedOrders() ([]*domain.DBOrder, error) {
	var orders []*domain.DBOrder

	stmt, err := s.db.Prepare("SELECT id, number, status, uploaded_at, user_id FROM orders WHERE status in ('NEW', 'PROCESSING')")
	if err != nil {
		s.logger.Errorf("Can't prepare statement for unprocessed order selection, err: %s", err.Error())
		return nil, fmt.Errorf("error getting all unprocessed orders: %w", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		s.logger.Errorf("Can't query unprocessed orders, err: %s", err.Error())
		return nil, fmt.Errorf("error getting all unprocessed orders: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var order domain.DBOrder
		err := rows.Scan(&order.ID, &order.Number, &order.Status, &order.UploadedAt, &order.UserID)
		if err != nil {
			s.logger.Errorf("Can't scan order to struct, err: %s", err.Error())
			return nil, fmt.Errorf("error getting all unprocessed orders: %w", err)
		}
		orders = append(orders, &order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error getting all unprocessed orders: %w", err)
	}

	return orders, nil
}

func (s *Storage) GetAllUserOrders(userID int64) ([]*domain.Order, error) {
	var orders []*domain.Order

	query := `
            SELECT o.number, o.status, a.accrual, o.uploaded_at
            FROM orders o
	        LEFT JOIN accruals a ON o.id = a.order_id
	        WHERE o.user_id = $1
            ORDER BY o.uploaded_at DESC`

	stmt, err := s.db.Prepare(query)
	if err != nil {
		s.logger.Errorf("Can't prepare query for user_id: %d orders, err: %s", userID, err.Error())
		return nil, fmt.Errorf("error getting all users orders: %w", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(userID)
	if err != nil {
		s.logger.Errorf("Can't query orders for user_id: %d, err: %s", userID, err.Error())
		return nil, fmt.Errorf("error getting all users orders: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var order domain.Order
		var accrual sql.NullInt64
		err := rows.Scan(&order.Number, &order.Status, &accrual, &order.UploadedAt)
		if err != nil {
			s.logger.Errorf("Can't scan order to struct for user_id: %d, err: %s", userID, err.Error())
			return nil, fmt.Errorf("error getting all users orders: %w", err)
		}
		if accrual.Valid {
			accrualValue := formatter.ConvertToCurrency(accrual.Int64)
			order.Accrual = &accrualValue
		}
		orders = append(orders, &order)
	}

	if err := rows.Err(); err != nil {
		s.logger.Errorf("Got rows.Err() for user_id: %d, err: %s", userID, err.Error())
		return nil, fmt.Errorf("error getting all users orders: %w", err)
	}

	return orders, nil
}

func (s *Storage) UpdateOrderAccrualStatus(id int64, status string, accrual *float64) error {
	tx, err := s.db.Begin()
	if err != nil {
		s.logger.Errorf("Transaction for order and accrual update error order_id: %d, err: %s", id, err.Error())
		return fmt.Errorf("error updating orders status: %w", err)
	}

	if _, err := s.db.Exec("UPDATE orders SET status = $2 WHERE id = $1", id, status); err != nil {
		s.logger.Errorf("Failed to update order status, order_id: %d, status: %s; err: %s", id, status, err.Error())
		_ = tx.Rollback()
		return fmt.Errorf("error updating orders status: %w", err)
	}

	if accrual != nil {
		accrualValue := formatter.ConvertToSubunit(*accrual)
		if _, err := s.db.Exec("INSERT INTO accruals (order_id, accrual) VALUES ($1, $2)", id, accrualValue); err != nil {
			s.logger.Errorf("Failed to insert new accrual, order_id: %d, err: %s", id, err.Error())
			_ = tx.Rollback()
			return fmt.Errorf("error inserting accrual: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		s.logger.Errorf("Transaction for order and accrual update commit error order_id: %d, err: %s", id, err.Error())
		return fmt.Errorf("error updating orders status: %w", err)
	}

	return nil
}
