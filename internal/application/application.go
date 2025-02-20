package application

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/frolmr/gophermart/internal/api"
	"github.com/frolmr/gophermart/internal/client"
	"github.com/frolmr/gophermart/internal/config"
	"github.com/frolmr/gophermart/internal/db/migrator"
	"github.com/frolmr/gophermart/internal/service"
	"github.com/frolmr/gophermart/internal/storage"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

type App struct {
	config        *config.AppConfig
	logger        *zap.SugaredLogger
	storage       *storage.Storage
	api           *api.API
	accrualClient *client.AccrualClient
	orderP        *service.OrderProcessor
}

func NewApp() (*App, error) {
	conf, err := config.NewAppConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to setup application config: %w", err)
	}

	lgr, err := setupLogger()
	if err != nil {
		return nil, fmt.Errorf("error initializing logger: %w", err)
	}

	db, err := setupDB(conf)
	if err != nil {
		return nil, fmt.Errorf("failed to setup database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to the database: %w", err)
	}

	stor := storage.NewStorage(db, lgr)

	srv, err := api.NewAPI(lgr, conf, stor)
	if err != nil {
		return nil, fmt.Errorf("failed to setup api: %w", err)
	}

	client := client.NewAccrualClient(resty.New(), conf, lgr)
	orderP := service.NewOrderProcessor(lgr, stor, client)

	return &App{
		config:        conf,
		logger:        lgr,
		storage:       stor,
		api:           srv,
		accrualClient: client,
		orderP:        orderP,
	}, nil
}

func (app *App) Run(stopCh <-chan struct{}, wg *sync.WaitGroup) {
	app.api.Run(stopCh, wg)
}

func (app *App) RunOrdersWorker(stopCh <-chan struct{}, wg *sync.WaitGroup) {
	app.orderP.Run(stopCh, wg)
}

func setupLogger() (*zap.SugaredLogger, error) {
	l, err := zap.NewDevelopment()

	if err != nil {
		return nil, err
	}

	return l.Sugar(), nil
}

func setupDB(conf *config.AppConfig) (*sql.DB, error) {
	db, err := sql.Open("pgx", conf.DatabaseURI)
	if err != nil {
		return nil, err
	}

	migrator := migrator.NewMigrator(conf.DatabaseURI)
	if err := migrator.RunMigrations(); err != nil {
		return nil, err
	}

	return db, nil
}
