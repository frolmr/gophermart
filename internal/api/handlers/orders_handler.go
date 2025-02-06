package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/frolmr/gophermart/internal/domain"
	"github.com/frolmr/gophermart/pkg/luhn"
	"go.uber.org/zap"
)

const (
	reqBodySizeLimit = int64(10 << 20) // 10 MB
)

type OrdersRepository interface {
	FindOrderByNumber(number string) (*domain.DBOrder, error)
	GetAllUserOrders(userID int) ([]*domain.Order, error)
	CreateOrder(number string, userID int) error
}

type OrdersHandler struct {
	logger *zap.SugaredLogger
	repo   OrdersRepository
}

func NewOrdersHandler(lgr *zap.SugaredLogger, repo OrdersRepository) *OrdersHandler {
	return &OrdersHandler{
		logger: lgr,
		repo:   repo,
	}
}

func (oh *OrdersHandler) LoadOrder(w http.ResponseWriter, req *http.Request) {
	req.Body = http.MaxBytesReader(w, req.Body, reqBodySizeLimit)

	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "Wrong request format", http.StatusBadRequest)
		return
	}

	orderNumber := string(body)
	if orderNumberValid := luhn.Check(orderNumber); !orderNumberValid {
		http.Error(w, "Order number is invalid", http.StatusUnprocessableEntity)
		return
	}
	defer req.Body.Close()

	userID, err := strconv.Atoi(req.Header.Get(domain.UserIDHeader))
	if err != nil {
		http.Error(w, "Invalid user id", http.StatusInternalServerError)
		return
	}

	existingOrder, err := oh.repo.FindOrderByNumber(orderNumber)
	if err != nil {
		oh.logger.Error("database error: ", err.Error())
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if existingOrder != nil {
		if existingOrder.UserID == userID {
			w.WriteHeader(http.StatusOK)
			return
		} else {
			http.Error(w, "Already downloaded", http.StatusConflict)
			return
		}
	}

	if err := oh.repo.CreateOrder(orderNumber, userID); err != nil {
		http.Error(w, "Failed to load order", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write([]byte("Order uploaded"))
}

//nolint:dupl // Actually code is not the same as in ordes_handler. United code will be more difficult to understand
func (oh *OrdersHandler) GetOrders(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", domain.JSONContentType)
	userID, err := strconv.Atoi(req.Header.Get(domain.UserIDHeader))
	if err != nil {
		http.Error(w, "Invalid user id", http.StatusInternalServerError)
		return
	}

	orders, err := oh.repo.GetAllUserOrders(userID)
	if err != nil {
		http.Error(w, "Failed to load orders", http.StatusInternalServerError)
		return
	}
	if len(orders) > 0 {
		if err := json.NewEncoder(w).Encode(orders); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}
