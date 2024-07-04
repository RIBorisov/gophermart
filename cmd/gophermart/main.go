package main

import (
	"context"
	"net/http"

	"github.com/RIBorisov/gophermart/internal/config"
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
	svc := &service.Service{Log: log, Storage: store, Config: cfg}

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
	return srv.ListenAndServe()
}
