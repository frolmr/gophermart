package controller

import (
	"github.com/frolmr/gophermart/internal/api/handlers"
	mw "github.com/frolmr/gophermart/internal/api/middleware"
	"github.com/frolmr/gophermart/internal/config"
	"github.com/frolmr/gophermart/internal/domain"
	"github.com/frolmr/gophermart/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type Controller struct {
	Storage    *storage.Storage
	AuthConfig *config.AuthConfig
}

func NewController(stor *storage.Storage) (*Controller, error) {
	authCfg, err := config.NewAuthConfig()
	if err != nil {
		return nil, err
	}

	return &Controller{
		Storage:    stor,
		AuthConfig: authCfg,
	}, nil
}

func (c *Controller) SetupRouter(lgr *zap.SugaredLogger) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	rh := handlers.NewRequestHandlers(lgr, c.Storage)

	r.Route("/api/user/", func(r chi.Router) {
		r.Use(middleware.AllowContentType(domain.JSONContentType))
		r.Post("/register", rh.UsersHandler.RegisterUser(c.AuthConfig))
		r.Post("/login", rh.UsersHandler.LoginUser(c.AuthConfig))
	})

	r.Route("/api/user/orders", func(r chi.Router) {
		r.Use(mw.WithAuth(c.AuthConfig))
		r.Post("/", rh.OrdersHandler.LoadOrder)
		r.Get("/", rh.OrdersHandler.GetOrders)
	})

	r.Route("/api/user/balance", func(r chi.Router) {
		r.Use(mw.WithAuth(c.AuthConfig))
		r.Get("/", rh.BalancesHandler.GetBalance)
		r.Post("/withdraw", rh.WithdrawalsHandler.RegisterWithdrawal)
	})

	r.With(mw.WithAuth(c.AuthConfig)).Get("/api/user/withdrawals", rh.WithdrawalsHandler.GetWithdrawals)

	return r
}
