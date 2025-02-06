package main

import (
	"database/sql"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/frolmr/gophermart/internal/api"
	"github.com/frolmr/gophermart/internal/client"
	"github.com/frolmr/gophermart/internal/config"
	"github.com/frolmr/gophermart/internal/db/migrator"
	"github.com/frolmr/gophermart/internal/service"
	"github.com/frolmr/gophermart/internal/storage"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

func main() {
	conf, err := config.NewAppConfig()
	if err != nil {
		log.Panic("error initializing app: ", err)
	}

	lgr, err := setupLogger()
	if err != nil {
		log.Panic("error initializing logger: ", err)
	}

	db, err := setupDB(conf)
	if err != nil {
		log.Panic("failed to setup database: ", err)
	}
	if err := db.Ping(); err != nil {
		log.Panic("failed to connect to the database: ", err)
	}
	defer db.Close()

	stor := storage.NewStorage(db, lgr)

	srv, err := api.NewAPI(lgr, conf, stor)
	if err != nil {
		log.Panic("failed to setup api: ", err)
	}
	client := client.NewAccrualClient(resty.New(), conf, lgr)

	stopCh := make(chan struct{})

	var wg sync.WaitGroup

	wg.Add(1)
	orderP := service.NewOrderProcessor(lgr, stor, client)
	go orderP.Run(stopCh, &wg)

	wg.Add(1)
	go srv.Run(stopCh, &wg)

	termCh := make(chan os.Signal, 1)
	signal.Notify(termCh, syscall.SIGINT)
	<-termCh
	close(stopCh)
	wg.Wait()
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
