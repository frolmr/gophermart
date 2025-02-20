package domain

import "time"

type Order struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    *float64  `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type DBOrder struct {
	ID         int64
	Number     string
	Status     string
	UploadedAt time.Time
	UserID     int64
}

type AccrualOrder struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}
