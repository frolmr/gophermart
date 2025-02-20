package domain

import "time"

type Withdrawal struct {
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at,omitempty"`
}

type DBWithdrawal struct {
	ID          int64
	OrderNumber string
	Sum         int64
	ProcessedAt time.Time
	UserID      int64
}
