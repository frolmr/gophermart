package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/frolmr/gophermart/internal/domain"
	"github.com/frolmr/gophermart/pkg/formatter"
	"go.uber.org/zap"
)

type BalanceRepository interface {
	GetUserCurrentBalance(userID int64) (float64, error)
	GetUserWithdrawalsSum(userID int64) (float64, error)
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
	userID, err := formatter.StringToInt64(req.Header.Get(domain.UserIDHeader))
	if err != nil {
		writeJSONError(w, "Invalid user id")
		return
	}

	balanceSum, err := bh.repo.GetUserCurrentBalance(userID)
	if err != nil {
		writeJSONError(w, "Failed to get user accrual sum")
		return
	}

	withdrawalSum, err := bh.repo.GetUserWithdrawalsSum(userID)
	if err != nil {
		writeJSONError(w, "Failed to get user withdrawal sum")
		return
	}

	balance := domain.Balance{
		BalanceSum:    balanceSum,
		WithdrawalSum: withdrawalSum,
	}

	if err := json.NewEncoder(w).Encode(balance); err != nil {
		writeJSONError(w, "Failed to encode response")
	}
}

func writeJSONError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", domain.JSONContentType)
	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
