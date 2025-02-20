package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/frolmr/gophermart/internal/application"
)

func main() {
	app, err := application.NewApp()
	if err != nil {
		log.Panic("failed to setup application: ", err)
	}

	stopCh := make(chan struct{})
	var wg sync.WaitGroup

	wg.Add(1)
	go app.RunOrdersWorker(stopCh, &wg)

	wg.Add(1)
	go app.Run(stopCh, &wg)

	termCh := make(chan os.Signal, 1)
	signal.Notify(termCh, syscall.SIGINT)
	<-termCh
	close(stopCh)
	wg.Wait()
}
