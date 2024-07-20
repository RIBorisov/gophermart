package accrual

import (
	"context"
	"fmt"
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
				svc.Log.Info("not found orders for processing")
			}
		}
	}
}

func ProcessOrder(ctx context.Context, svc *service.Service, orderNo string, client *resty.Client) error {
	svc.Log.Info("starting process order", "order_id", orderNo)
	data, err := svc.FetchOrderInfo(ctx, client, orderNo)
	if err != nil {
		return fmt.Errorf("failed fetch order info: %w", err)
	}

	err = svc.UpdateOrder(ctx, data)
	if err != nil {
		return fmt.Errorf("failed update order: %w", err)
	}

	return nil
}
