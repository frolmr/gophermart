package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/frolmr/gophermart/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

func (s *Storage) CreateUser(login, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Errorf("Password encryption failed for user: %s, err: %s", login, err.Error())
		return fmt.Errorf("error creating user: %w", err)
	}

	if _, err := s.db.Exec("INSERT INTO users (login, password_hash) VALUES ($1, $2)", login, string(hashedPassword)); err != nil {
		s.logger.Errorf("New user insertion failed, user: %s, err: %s", login, err.Error())
		return fmt.Errorf("error creating user: %w", err)
	}

	return nil
}

func (s *Storage) CreateAndReturnUser(login, password string) (*domain.DBUser, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Errorf("Password encryption failed for user: %s, err: %s", login, err.Error())
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	stmt, err := s.db.Prepare("INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id, login, password_hash")
	if err != nil {
		s.logger.Errorf("Can't prepare statement for user: %s, err: %s", login, err.Error())
		return nil, fmt.Errorf("error creating user: %w", err)
	}
	defer stmt.Close()

	var user domain.DBUser
	err = stmt.QueryRow(login, hashedPassword).Scan(&user.ID, &user.Login, &user.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		} else {
			s.logger.Errorf("User query fails for user: %s, err: %s", login, err.Error())
			return nil, fmt.Errorf("error creating user: %w", err)
		}
	}

	return &user, nil
}

func (s *Storage) GetUserByLogin(login string) (*domain.DBUser, error) {
	stmt, err := s.db.Prepare("SELECT id, login, password_hash FROM users WHERE login = $1")
	if err != nil {
		s.logger.Errorf("Can't prepare statement for user: %s, err: %s", login, err.Error())
		return nil, fmt.Errorf("error getting user: %w", err)
	}
	defer stmt.Close()

	var user domain.DBUser
	err = stmt.QueryRow(login).Scan(&user.ID, &user.Login, &user.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		} else {
			s.logger.Errorf("User query fails for user: %s, err: %s", login, err.Error())
			return nil, fmt.Errorf("error getting user: %w", err)
		}
	}

	return &user, nil
}

func (s *Storage) StoreRefreshToken(userID int64, token string, expiresAt time.Time) error {
	query := `INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)`
	if _, err := s.db.Exec(query, userID, token, expiresAt); err != nil {
		s.logger.Errorf("Failed to store refresh token for user: %d, err: %s", userID, err.Error())
		return fmt.Errorf("error storing refresh token: %w", err)
	}
	return nil
}

func (s *Storage) GetRefreshToken(token string) (*domain.RefreshToken, error) {
	stmt, err := s.db.Prepare("SELECT id, user_id, token, expires_at FROM refresh_tokens WHERE token = $1")
	if err != nil {
		s.logger.Errorf("Can't prepare statement for refresh token: %s, err: %s", token, err.Error())
		return nil, fmt.Errorf("error getting refresh token: %w", err)
	}
	defer stmt.Close()

	var refreshToken domain.RefreshToken
	err = stmt.QueryRow(token).Scan(&refreshToken.ID, &refreshToken.UserID, &refreshToken.Token, &refreshToken.ExpiresAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		s.logger.Errorf("Failed to get refresh token: %s", err.Error())
		return nil, fmt.Errorf("error getting refresh token: %w", err)
	}

	return &refreshToken, nil
}

func (s *Storage) DeleteRefreshToken(token string) error {
	_, err := s.db.Exec("DELETE FROM refresh_tokens WHERE token = $1", token)
	if err != nil {
		s.logger.Errorf("Failed to delete refresh token: %s", err.Error())
		return fmt.Errorf("error deleting refresh token: %w", err)
	}
	return nil
}
