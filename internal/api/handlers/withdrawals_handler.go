package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/frolmr/gophermart/internal/domain"
	"github.com/frolmr/gophermart/pkg/luhn"
	"go.uber.org/zap"
)

type WithdrawalRepository interface {
	CreateWithdrawal(orderNumber string, sum float64, userID int) error
	GetUserCurrentBalance(userID int) (float64, error)
	GetAllUserWithdrawals(userID int) ([]*domain.Withdrawal, error)
}

type WithdrawalsHandler struct {
	logger *zap.SugaredLogger
	repo   WithdrawalRepository
}

func NewWithdrawalsHandler(lgr *zap.SugaredLogger, repo WithdrawalRepository) *WithdrawalsHandler {
	return &WithdrawalsHandler{
		logger: lgr,
		repo:   repo,
	}
}

func (wh *WithdrawalsHandler) RegisterWithdrawal(w http.ResponseWriter, req *http.Request) {
	var withdrawal domain.Withdrawal

	if err := json.NewDecoder(req.Body).Decode(&withdrawal); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if withdrawal.Order == "" || withdrawal.Sum <= 0 {
		http.Error(w, "Invalid order number or sum", http.StatusBadRequest)
		return
	}

	if orderNumberValid := luhn.Check(withdrawal.Order); !orderNumberValid {
		http.Error(w, "Order number is invalid", http.StatusUnprocessableEntity)
		return
	}

	userID, err := strconv.Atoi(req.Header.Get(domain.UserIDHeader))
	if err != nil {
		http.Error(w, "Invalid user id", http.StatusInternalServerError)
		return
	}

	currentBalance, err := wh.repo.GetUserCurrentBalance(userID)
	if err != nil {
		http.Error(w, "Failed to check user balance", http.StatusInternalServerError)
		return
	}

	if currentBalance < withdrawal.Sum {
		http.Error(w, "Not enough funds", http.StatusPaymentRequired)
		return
	}

	if err := wh.repo.CreateWithdrawal(withdrawal.Order, withdrawal.Sum, userID); err != nil {
		http.Error(w, "Failed to register withdrawal", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

//nolint:dupl // Actually code is not the same as in ordes_handler. United code will be more difficult to understand
func (wh *WithdrawalsHandler) GetWithdrawals(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", domain.JSONContentType)
	userID, err := strconv.Atoi(req.Header.Get(domain.UserIDHeader))
	if err != nil {
		http.Error(w, "Invalid user id", http.StatusInternalServerError)
		return
	}

	withdrawals, err := wh.repo.GetAllUserWithdrawals(userID)
	if err != nil {
		http.Error(w, "Failed to load withdrawals", http.StatusInternalServerError)
		return
	}
	if len(withdrawals) > 0 {
		if err := json.NewEncoder(w).Encode(withdrawals); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}
