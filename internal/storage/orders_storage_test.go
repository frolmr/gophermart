package storage

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/frolmr/gophermart/internal/domain"
	"github.com/frolmr/gophermart/pkg/formatter"
	"github.com/stretchr/testify/assert"
)

func TestFindOrderByNumber_Success(t *testing.T) {
	storage, mock := NewMockStorage(t)

	orderNumber := "12345678903"
	order := &domain.DBOrder{
		ID:         1,
		Number:     orderNumber,
		Status:     "NEW",
		UploadedAt: time.Now(),
		UserID:     1,
	}

	rows := sqlmock.NewRows([]string{"id", "number", "status", "uploaded_at", "user_id"}).
		AddRow(order.ID, order.Number, order.Status, order.UploadedAt, order.UserID)

	mock.ExpectPrepare("SELECT id, number, status, uploaded_at, user_id FROM orders").
		ExpectQuery().
		WithArgs(orderNumber).
		WillReturnRows(rows)

	result, err := storage.FindOrderByNumber(orderNumber)

	assert.NoError(t, err)
	assert.Equal(t, order, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindOrderByNumber_NotFound(t *testing.T) {
	storage, mock := NewMockStorage(t)

	orderNumber := "12345678903"

	mock.ExpectPrepare("SELECT id, number, status, uploaded_at, user_id FROM orders").
		ExpectQuery().
		WithArgs(orderNumber).
		WillReturnError(sql.ErrNoRows)

	result, err := storage.FindOrderByNumber(orderNumber)

	assert.NoError(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindOrderByNumber_DatabaseError(t *testing.T) {
	storage, mock := NewMockStorage(t)

	orderNumber := "12345678903"

	mock.ExpectPrepare("SELECT id, number, status, uploaded_at, user_id FROM orders").
		ExpectQuery().
		WithArgs(orderNumber).
		WillReturnError(errors.New("database error"))

	result, err := storage.FindOrderByNumber(orderNumber)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateOrder_Success(t *testing.T) {
	storage, mock := NewMockStorage(t)

	orderNumber := "12345678903"
	userID := int64(1)

	mock.ExpectExec("INSERT INTO orders").
		WithArgs(orderNumber, userID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := storage.CreateOrder(orderNumber, userID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateOrder_DatabaseError(t *testing.T) {
	storage, mock := NewMockStorage(t)

	orderNumber := "12345678903"
	userID := int64(1)

	mock.ExpectExec("INSERT INTO orders").
		WithArgs(orderNumber, userID).
		WillReturnError(errors.New("database error"))

	err := storage.CreateOrder(orderNumber, userID)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAllUnprocessedOrders_Success(t *testing.T) {
	storage, mock := NewMockStorage(t)

	orders := []*domain.DBOrder{
		{ID: 1, Number: "12345678903", Status: "NEW", UploadedAt: time.Now(), UserID: 1},
		{ID: 2, Number: "98765432109", Status: "PROCESSING", UploadedAt: time.Now(), UserID: 2},
	}

	rows := sqlmock.NewRows([]string{"id", "number", "status", "uploaded_at", "user_id"}).
		AddRow(orders[0].ID, orders[0].Number, orders[0].Status, orders[0].UploadedAt, orders[0].UserID).
		AddRow(orders[1].ID, orders[1].Number, orders[1].Status, orders[1].UploadedAt, orders[1].UserID)

	mock.ExpectPrepare("SELECT id, number, status, uploaded_at, user_id FROM orders").
		ExpectQuery().
		WillReturnRows(rows)

	result, err := storage.GetAllUnprocessedOrders()

	assert.NoError(t, err)
	assert.Equal(t, orders, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAllUnprocessedOrders_NoOrders(t *testing.T) {
	storage, mock := NewMockStorage(t)

	rows := sqlmock.NewRows([]string{"id", "number", "status", "uploaded_at", "user_id"})

	mock.ExpectPrepare("SELECT id, number, status, uploaded_at, user_id FROM orders").
		ExpectQuery().
		WillReturnRows(rows)

	result, err := storage.GetAllUnprocessedOrders()

	assert.NoError(t, err)
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAllUnprocessedOrders_DatabaseError(t *testing.T) {
	storage, mock := NewMockStorage(t)

	mock.ExpectPrepare("SELECT id, number, status, uploaded_at, user_id FROM orders").
		ExpectQuery().
		WillReturnError(errors.New("database error"))

	result, err := storage.GetAllUnprocessedOrders()

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAllUserOrders_Success(t *testing.T) {
	storage, mock := NewMockStorage(t)

	userID := int64(1)
	orders := []*domain.Order{
		{Number: "12345678903", Status: "NEW", UploadedAt: time.Now()},
		{Number: "98765432109", Status: "PROCESSED", UploadedAt: time.Now(), Accrual: func() *float64 { v := 50.0; return &v }()},
	}

	rows := sqlmock.NewRows([]string{"number", "status", "accrual", "uploaded_at"}).
		AddRow(orders[0].Number, orders[0].Status, nil, orders[0].UploadedAt).
		AddRow(orders[1].Number, orders[1].Status, formatter.ConvertToSubunit(50.0), orders[1].UploadedAt)

	mock.ExpectPrepare("SELECT o.number, o.status, a.accrual, o.uploaded_at FROM orders o").
		ExpectQuery().
		WithArgs(userID).
		WillReturnRows(rows)

	result, err := storage.GetAllUserOrders(userID)

	assert.NoError(t, err)
	assert.Equal(t, orders, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAllUserOrders_NoOrders(t *testing.T) {
	storage, mock := NewMockStorage(t)

	userID := int64(1)
	rows := sqlmock.NewRows([]string{"number", "status", "accrual", "uploaded_at"})

	mock.ExpectPrepare("SELECT o.number, o.status, a.accrual, o.uploaded_at FROM orders o").
		ExpectQuery().
		WithArgs(userID).
		WillReturnRows(rows)

	result, err := storage.GetAllUserOrders(userID)

	assert.NoError(t, err)
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAllUserOrders_DatabaseError(t *testing.T) {
	storage, mock := NewMockStorage(t)

	userID := int64(1)

	mock.ExpectPrepare("SELECT o.number, o.status, a.accrual, o.uploaded_at FROM orders o").
		ExpectQuery().
		WithArgs(userID).
		WillReturnError(errors.New("database error"))

	result, err := storage.GetAllUserOrders(userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateOrderAccrualStatus_Success(t *testing.T) {
	storage, mock := NewMockStorage(t)

	orderID := int64(1)
	status := "PROCESSED"
	accrual := 50.0
	accrualInSubunit := formatter.ConvertToSubunit(accrual)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE orders SET status = \$2 WHERE id = \$1`).
		WithArgs(orderID, status).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO accruals \(order_id, accrual\) VALUES \(\$1, \$2\)`).
		WithArgs(orderID, accrualInSubunit).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := storage.UpdateOrderAccrualStatus(orderID, status, &accrual)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateOrderAccrualStatus_NoAccrual(t *testing.T) {
	storage, mock := NewMockStorage(t)

	orderID := int64(1)
	status := "PROCESSED"

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE orders SET status = \$2 WHERE id = \$1`).
		WithArgs(orderID, status).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := storage.UpdateOrderAccrualStatus(orderID, status, nil)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateOrderAccrualStatus_DatabaseError(t *testing.T) {
	storage, mock := NewMockStorage(t)

	orderID := int64(1)
	status := "PROCESSED"
	accrual := 50.0

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE orders SET status = \$2 WHERE id = \$1`).
		WithArgs(orderID, status).
		WillReturnError(errors.New("database error"))
	mock.ExpectRollback()

	err := storage.UpdateOrderAccrualStatus(orderID, status, &accrual)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
