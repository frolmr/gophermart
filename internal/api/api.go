package api

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/frolmr/gophermart/internal/api/controller"
	"github.com/frolmr/gophermart/internal/config"
	"github.com/frolmr/gophermart/internal/storage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

const (
	readTimeout     = 3 * time.Second
	writeTimeout    = 5 * time.Second
	shutdownTimeout = 10 * time.Second
)

type API struct {
	router  chi.Router
	config  *config.AppConfig
	logger  *zap.SugaredLogger
	storage *storage.Storage
}

func NewAPI(lgr *zap.SugaredLogger, cfg *config.AppConfig, stor *storage.Storage) (*API, error) {
	ctrl, err := controller.NewController(stor)
	if err != nil {
		return nil, err
	}

	return &API{
		router:  ctrl.SetupRouter(lgr),
		logger:  lgr,
		config:  cfg,
		storage: stor,
	}, nil
}

func (a *API) Run(stopCh <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	srv := &http.Server{
		Addr:         a.config.RunAddress,
		Handler:      a.router,
		IdleTimeout:  time.Minute,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	go func() {
		a.logger.Infow("Starting server", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Fatalf("Server error: %v", err)
		}
	}()

	<-stopCh
	a.logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		a.logger.Fatalf("Server shutdown error: %v", err)
	}

	a.logger.Info("Server gracefully stopped")
}
