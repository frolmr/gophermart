package storage

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/frolmr/gophermart/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestCreateUser_Success(t *testing.T) {
	storage, mock := NewMockStorage(t)

	login := "testuser"
	password := "testpassword"

	mock.ExpectExec("INSERT INTO users").
		WithArgs(login, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := storage.CreateUser(login, password)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateUser_DatabaseError(t *testing.T) {
	storage, mock := NewMockStorage(t)

	login := "testuser"
	password := "testpassword"

	mock.ExpectExec("INSERT INTO users").
		WithArgs(login, sqlmock.AnyArg()).
		WillReturnError(errors.New("database error"))

	err := storage.CreateUser(login, password)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateAndReturnUser_Success(t *testing.T) {
	storage, mock := NewMockStorage(t)

	login := "testuser"
	password := "testpassword"
	hashedPassword := "hashedpassword"
	userID := int64(1)

	mock.ExpectPrepare("INSERT INTO users").
		ExpectQuery().
		WithArgs(login, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "login", "password_hash"}).
			AddRow(userID, login, hashedPassword))

	user, err := storage.CreateAndReturnUser(login, password)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, login, user.Login)
	assert.Equal(t, hashedPassword, user.PasswordHash)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateAndReturnUser_DatabaseError(t *testing.T) {
	storage, mock := NewMockStorage(t)

	login := "testuser"
	password := "testpassword"

	mock.ExpectPrepare("INSERT INTO users").
		ExpectQuery().
		WithArgs(login, sqlmock.AnyArg()).
		WillReturnError(errors.New("database error"))

	user, err := storage.CreateAndReturnUser(login, password)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserByLogin_Success(t *testing.T) {
	storage, mock := NewMockStorage(t)

	login := "testuser"
	hashedPassword := "hashedpassword"
	userID := int64(1)

	mock.ExpectPrepare("SELECT id, login, password_hash FROM users").
		ExpectQuery().
		WithArgs(login).
		WillReturnRows(sqlmock.NewRows([]string{"id", "login", "password_hash"}).
			AddRow(userID, login, hashedPassword))

	user, err := storage.GetUserByLogin(login)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, login, user.Login)
	assert.Equal(t, hashedPassword, user.PasswordHash)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserByLogin_NotFound(t *testing.T) {
	storage, mock := NewMockStorage(t)

	login := "testuser"

	mock.ExpectPrepare("SELECT id, login, password_hash FROM users").
		ExpectQuery().
		WithArgs(login).
		WillReturnError(sql.ErrNoRows)

	user, err := storage.GetUserByLogin(login)

	assert.NoError(t, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserByLogin_DatabaseError(t *testing.T) {
	storage, mock := NewMockStorage(t)

	login := "testuser"

	mock.ExpectPrepare("SELECT id, login, password_hash FROM users").
		ExpectQuery().
		WithArgs(login).
		WillReturnError(errors.New("database error"))

	user, err := storage.GetUserByLogin(login)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStoreRefreshToken_Success(t *testing.T) {
	storage, mock := NewMockStorage(t)

	userID := int64(1)
	token := "valid-refresh-token"
	expiresAt := time.Now().Add(time.Hour)

	mock.ExpectExec("INSERT INTO refresh_tokens").
		WithArgs(userID, token, expiresAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := storage.StoreRefreshToken(userID, token, expiresAt)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStoreRefreshToken_DatabaseError(t *testing.T) {
	storage, mock := NewMockStorage(t)

	userID := int64(1)
	token := "valid-refresh-token"
	expiresAt := time.Now().Add(time.Hour)

	mock.ExpectExec("INSERT INTO refresh_tokens").
		WithArgs(userID, token, expiresAt).
		WillReturnError(errors.New("database error"))

	err := storage.StoreRefreshToken(userID, token, expiresAt)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetRefreshToken_Success(t *testing.T) {
	storage, mock := NewMockStorage(t)

	token := "valid-refresh-token"
	refreshToken := &domain.RefreshToken{
		ID:        1,
		UserID:    1,
		Token:     token,
		ExpiresAt: time.Now().Add(time.Hour),
	}

	mock.ExpectPrepare("SELECT id, user_id, token, expires_at FROM refresh_tokens WHERE token = \\$1").
		ExpectQuery().
		WithArgs(token).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "token", "expires_at"}).
			AddRow(refreshToken.ID, refreshToken.UserID, refreshToken.Token, refreshToken.ExpiresAt))

	result, err := storage.GetRefreshToken(token)

	assert.NoError(t, err)
	assert.Equal(t, refreshToken, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetRefreshToken_NotFound(t *testing.T) {
	storage, mock := NewMockStorage(t)

	token := "invalid-refresh-token"

	mock.ExpectPrepare("SELECT id, user_id, token, expires_at FROM refresh_tokens WHERE token = \\$1").
		ExpectQuery().
		WithArgs(token).
		WillReturnError(sql.ErrNoRows)

	result, err := storage.GetRefreshToken(token)

	assert.NoError(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetRefreshToken_DatabaseError(t *testing.T) {
	storage, mock := NewMockStorage(t)

	token := "invalid-refresh-token"

	mock.ExpectPrepare("SELECT id, user_id, token, expires_at FROM refresh_tokens WHERE token = \\$1").
		ExpectQuery().
		WithArgs(token).
		WillReturnError(errors.New("database error"))

	result, err := storage.GetRefreshToken(token)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteRefreshToken_Success(t *testing.T) {
	storage, mock := NewMockStorage(t)

	token := "valid-refresh-token"

	mock.ExpectExec("DELETE FROM refresh_tokens").
		WithArgs(token).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := storage.DeleteRefreshToken(token)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteRefreshToken_DatabaseError(t *testing.T) {
	storage, mock := NewMockStorage(t)

	token := "invalid-refresh-token"

	mock.ExpectExec("DELETE FROM refresh_tokens").
		WithArgs(token).
		WillReturnError(errors.New("database error"))

	err := storage.DeleteRefreshToken(token)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
