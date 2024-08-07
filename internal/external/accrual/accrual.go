package accrual

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/RIBorisov/gophermart/internal/service"
)

func GetOrders(ctx context.Context, svc *service.Service, ordersCh chan<- string) {
	ticker := time.NewTicker(svc.Config.Service.AccrualPollInterval)
	for {
		select {
		case <-ctx.Done():
			close(ordersCh)
			return
		case <-ticker.C:
			oList, err := svc.GetOrdersForProcessing(ctx)
			if err != nil {
				svc.Log.Err("failed get unprocessed orders", err)
				continue
			}
			if len(oList) > 0 {
				svc.Log.Info("got order ids for processing", "count", len(oList))
				for _, o := range oList {
					ordersCh <- o
				}
			} else {
				svc.Log.Info("not found orders for processing in db")
			}
		}
	}
}

type retryCtrl struct {
	retry bool
	wait  time.Duration
	mu    sync.Mutex
}

func FetchAndUpdateOrders(ctx context.Context, svc *service.Service, orderID string) error {
	retry := &retryCtrl{}
	client := resty.New().SetBaseURL(svc.Config.Service.AccrualSystemAddress)
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
			svc.Log.Info("starting process order", "order_id", orderID)
			data, fetchErr := svc.FetchOrderInfo(ctx, client, orderID)
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
			if err := svc.UpdateOrder(ctx, data); err != nil {
				return fmt.Errorf("failed update order: %w", err)
			}
		}
		break
	}

	return nil
}
