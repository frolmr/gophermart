package storage

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
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
