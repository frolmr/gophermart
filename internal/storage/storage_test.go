package storage

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"
)

func NewMockStorage(t *testing.T) (*Storage, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}

	logger := zap.NewNop().Sugar()
	storage := NewStorage(db, logger)

	return storage, mock
}
