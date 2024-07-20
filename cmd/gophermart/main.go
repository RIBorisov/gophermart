package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"

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
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("failed load config", err)
	}

	if err = runApp(ctx, cfg, log); err != nil {
		log.Fatal("failed run application\n", err)
	}
}

func runApp(ctx context.Context, cfg *config.Config, log *logger.Log) error {
	store, err := storage.LoadStorage(ctx, cfg, log)
	if err != nil {
		log.Fatal("failed load storage", err)
	}

	defer func() {
		if err = store.ClosePool(); err != nil {
			log.Err("failed close connection pool", err)
		}
	}()

	svc := &service.Service{Log: log, Storage: store, Config: cfg}

	ordersCh := make(chan string)
	resultCh := make(chan string)
	client := initClient(svc)

	go func() {
		for o := range ordersCh {
			select {
			case <-ctx.Done():
				svc.Log.Info("Done reading from channel")
				return
			default:
				accrual.ProcessOrder(ctx, svc, o, client, resultCh)
			}
		}
	}()

	go accrual.GetOrders(ctx, svc, ordersCh)

	r := handlers.NewRouter(svc)

	srv := &http.Server{
		Addr:         cfg.Service.RunAddress,
		Handler:      r,
		ReadTimeout:  cfg.Service.Timeout,
		WriteTimeout: cfg.Service.Timeout,
		IdleTimeout:  cfg.Service.IdleTimeout,
	}

	svc.Log.Info(
		"starting application",
		"RUN_ADDRESS", srv.Addr,
		"ACCRUAL_SYSTEM_ADDRESS", cfg.Service.AccrualSystemAddress,
	)

	go enableGracefulShutdown(ctx, svc, srv, ordersCh)

	if err = srv.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			svc.Log.Info("server closed")
			return nil
		}
		return fmt.Errorf("failed listen and serve: %w", err)
	}

	return nil
}

func enableGracefulShutdown(ctx context.Context, svc *service.Service, srv *http.Server, ch chan string) {
	ctx, cancelCtx := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancelCtx()

	select {
	case <-ctx.Done():
		svc.Log.Warn("received signal to stop application")
		close(ch)
		cancelCtx()

		if err := srv.Shutdown(ctx); err != nil {
			svc.Log.Fatal("failed make graceful shutdown")
		}
	}
}

func initClient(svc *service.Service) *resty.Client {
	client := resty.New().SetBaseURL(svc.Config.Service.AccrualSystemAddress)
	client.AddRetryCondition(func(r *resty.Response, err error) bool {
		if r.StatusCode() == http.StatusTooManyRequests {
			retryAfter, err := strconv.Atoi(r.Header().Get("Retry-After"))
			if err != nil {
				svc.Log.Err("failed convert string to integer Retry-After header value: %w", err)
				return true
			}
			time.Sleep(time.Duration(retryAfter) * time.Second)
			return true
		}

		return false
	})
	return client
}
