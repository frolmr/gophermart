package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/frolmr/gophermart/internal/domain"
	"go.uber.org/zap"
)

type BalanceRepository interface {
	GetUserCurrentBalance(userID int) (float64, error)
	GetUserWithdrawalsSum(userID int) (float64, error)
}

type BalancesHandler struct {
	logger *zap.SugaredLogger
	repo   BalanceRepository
}

func NewBalancesHandler(lgr *zap.SugaredLogger, repo BalanceRepository) *BalancesHandler {
	return &BalancesHandler{
		logger: lgr,
		repo:   repo,
	}
}

func (bh *BalancesHandler) GetBalance(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", domain.JSONContentType)
	userID, err := strconv.Atoi(req.Header.Get(domain.UserIDHeader))
	if err != nil {
		http.Error(w, "Invalid user id", http.StatusInternalServerError)
		return
	}

	balanceSum, err := bh.repo.GetUserCurrentBalance(userID)
	if err != nil {
		http.Error(w, "Failed to get user accrual sum", http.StatusInternalServerError)
		return
	}

	withdrawalSum, err := bh.repo.GetUserWithdrawalsSum(userID)
	if err != nil {
		http.Error(w, "Failed to get user withdrawal sum", http.StatusInternalServerError)
		return
	}

	var balance domain.Balance
	balance.BalanceSum = balanceSum
	balance.WithdrawalSum = withdrawalSum

	if err := json.NewEncoder(w).Encode(balance); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
