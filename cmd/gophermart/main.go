package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"
	"golang.org/x/sync/errgroup"

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

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("failed load config", err)
	}

	if err = runApp(cfg, log); err != nil {
		log.Fatal("failed run application", err)
	}
}

func runApp(cfg *config.Config, log *logger.Log) error {
	rootCtx, cancelCtx := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancelCtx()

	g, ctx := errgroup.WithContext(rootCtx)

	store, err := storage.LoadStorage(ctx, cfg, log)
	if err != nil {
		return fmt.Errorf("failed load storage: %w", err)
	}

	defer func() {
		if err = store.ClosePool(); err != nil {
			log.Err("failed close connection pool", err)
		}
	}()

	svc := &service.Service{Log: log, Storage: store, Config: cfg}

	ordersCh := make(chan string)

	client := initClient(svc)
	retry := &retryCtrl{}

	const (
		workerNum       = 5
		timeoutShutdown = time.Second * 5
	)

	for range workerNum {
		g.Go(func() error {
			for o := range ordersCh {
				svc.Log.Info("incoming new order", "order_id", o)
				for {
					select {
					case <-ctx.Done():
						return ctx.Err()
					default:
						if retry.retry {
							svc.Log.Info("Waiting for retry", "seconds", retry.wait)
							time.Sleep(retry.wait)
							retry.mu.Lock()
							retry.wait = 0 * time.Second
							retry.retry = false
							retry.mu.Unlock()
						}
						svc.Log.Info("starting process order", "order_id", o)
						data, fetchErr := svc.FetchOrderInfo(ctx, client, o)
						if fetchErr != nil {
							var errToManyRequests *service.ToManyRequestsError
							if errors.As(fetchErr, &errToManyRequests) {
								retry.mu.Lock()
								retry.retry = true
								retry.wait = errToManyRequests.RetryAfter
								retry.mu.Unlock()
								continue
							} else {
								return fmt.Errorf("failed fetch order info: %w", fetchErr)
							}
						}

						if data == nil {
							svc.Log.Info("not found orders for processing in accrual service")
							continue
						}
						err = svc.UpdateOrder(ctx, data)
						if err != nil {
							return fmt.Errorf("failed update order: %w", err)
						}
					}
					break
				}
			}
			return nil
		})
	}

	g.Go(func() error {
		accrual.GetOrders(ctx, svc, ordersCh)
		<-ctx.Done()
		svc.Log.Debug("closing GetOrders goroutine")
		return nil
	})

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

	g.Go(func() error {
		enableGracefulShutdown(ctx, svc, srv)
		<-ctx.Done()
		svc.Log.Debug("closing GracefulShutdown goroutine")
		return nil
	})

	context.AfterFunc(ctx, func() {
		ctx, cancelCtx := context.WithTimeout(context.Background(), timeoutShutdown)
		defer cancelCtx()

		<-ctx.Done()
		log.Fatal("failed do graceful shutdown")
	})

	g.Go(func() error {
		if err = srv.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				svc.Log.Info("server closed")
			} else {
				return fmt.Errorf("failed listen and serve: %w", err)
			}
		}
		<-ctx.Done()
		svc.Log.Debug("closing ListenAndServe goroutine")

		return nil
	})

	if err = g.Wait(); err != nil {
		svc.Log.Err("failed wait for goroutines finished", err)
	}

	return nil
}

func enableGracefulShutdown(ctx context.Context, svc *service.Service, srv *http.Server) {
	ctx, cancelCtx := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancelCtx()

	<-ctx.Done()
	svc.Log.Warn("received signal to stop application")
	cancelCtx()

	if err := srv.Shutdown(ctx); err != nil {
		svc.Log.Fatal("failed make graceful shutdown")
	}
}

func initClient(svc *service.Service) *resty.Client {
	return resty.New().SetBaseURL(svc.Config.Service.AccrualSystemAddress)
}

type retryCtrl struct {
	retry bool
	wait  time.Duration
	mu    sync.Mutex
}
