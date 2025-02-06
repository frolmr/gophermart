package storage

import (
	"database/sql"
	"errors"

	"github.com/frolmr/gophermart/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

func (s *Storage) CreateUser(login, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Errorf("Password encryption failed for user: %s, err: %s", login, err.Error())
		return err
	}

	if _, err := s.db.Exec("INSERT INTO users (login, password_hash) VALUES ($1, $2)", login, string(hashedPassword)); err != nil {
		s.logger.Errorf("New user insertion failed, user: %s, err: %s", login, err.Error())
		return err
	}

	return nil
}

func (s *Storage) CreateAndReturnUser(login, password string) (*domain.DBUser, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Errorf("Password encryption failed for user: %s, err: %s", login, err.Error())
		return nil, err
	}

	stmt, err := s.db.Prepare("INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id, login, password_hash")
	if err != nil {
		s.logger.Errorf("Can't prepare statement for user: %s, err: %s", login, err.Error())
		return nil, err
	}
	defer stmt.Close()

	var user domain.DBUser
	err = stmt.QueryRow(login, hashedPassword).Scan(&user.ID, &user.Login, &user.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		} else {
			s.logger.Errorf("User query fails for user: %s, err: %s", login, err.Error())
			return nil, err
		}
	}

	return &user, nil
}

func (s *Storage) GetUserByLogin(login string) (*domain.DBUser, error) {
	stmt, err := s.db.Prepare("SELECT id, login, password_hash FROM users WHERE login = $1")
	if err != nil {
		s.logger.Errorf("Can't prepare statement for user: %s, err: %s", login, err.Error())
		return nil, err
	}
	defer stmt.Close()

	var user domain.DBUser
	err = stmt.QueryRow(login).Scan(&user.ID, &user.Login, &user.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		} else {
			s.logger.Errorf("User query fails for user: %s, err: %s", login, err.Error())
			return nil, err
		}
	}

	return &user, nil
}
