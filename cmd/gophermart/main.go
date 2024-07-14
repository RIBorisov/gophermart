package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/RIBorisov/gophermart/internal/config"
	"github.com/RIBorisov/gophermart/internal/external/accrual"
	"github.com/RIBorisov/gophermart/internal/handlers"
	"github.com/RIBorisov/gophermart/internal/logger"
	"github.com/RIBorisov/gophermart/internal/service"
	"github.com/RIBorisov/gophermart/internal/storage"
)

func main() {
	log := &logger.Log{}
	log.Initialize("DEBUG")

	ctx := context.Background()
	cfg, err := config.LoadConfig(log)
	if err != nil {
		log.Fatal("failed load config\n", err)
	}

	err = runApp(ctx, cfg, log)
	if err != nil {
		log.Fatal("failed run application\n", err)
	}
}

func runApp(ctx context.Context, cfg *config.Config, log *logger.Log) error {
	store, err := storage.LoadStorage(ctx, cfg, log)
	if err != nil {
		log.Fatal("failed load storage\n", err)
	}

	defer func() {
		if err = store.ClosePool(); err != nil {
			log.Err("failed close connection pool", err)
		}
	}()

	svc := &service.Service{Log: log, Storage: store, Config: cfg}

	ordersCh, err := accrual.RunPoller(ctx, svc)
	if err != nil {
		return fmt.Errorf("failed run accrual poller: %w", err)
	}

	r := handlers.NewRouter(svc)

	srv := &http.Server{
		Addr:    cfg.Service.RunAddress,
		Handler: r,
	}

	svc.Log.Info(
		"starting application",
		"RUN_ADDRESS", srv.Addr,
		"ACCRUAL_SYSTEM_ADDRESS", cfg.Service.AccrualSystemAddress,
	)
	go enableGracefulShutdown(ctx, svc, srv, ordersCh)

	return srv.ListenAndServe()
}

func enableGracefulShutdown(ctx context.Context, svc *service.Service, srv *http.Server, ch chan string) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	svc.Log.Warn("received signal, shutting down...")
	close(ch)

	if err := srv.Shutdown(ctx); err != nil {
		svc.Log.Fatal("failed make graceful shutdown")
	}
}
