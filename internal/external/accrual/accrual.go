package accrual

import (
	"context"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/RIBorisov/gophermart/internal/service"
)

func RunPoller(ctx context.Context, svc *service.Service) (chan string, error) {
	client := resty.New().SetBaseURL(svc.Config.Service.AccrualSystemAddress)
	const numWorkers = 5

	interval := svc.Config.Service.AccrualPollInterval

	ordersCh := make(chan string)
	resultCh := make(chan string)
	for i := 0; i < numWorkers; i++ {
		go func() {
			for o := range ordersCh {
				select {
				case <-ctx.Done():
					svc.Log.Info("Done reading from channel")
					return
				default:
					processOrder(ctx, svc, o, client, resultCh)
				}
			}
		}()
	}

	go getOrders(ctx, interval, svc, ordersCh)

	return ordersCh, nil
}

func getOrders(ctx context.Context, interval time.Duration, svc *service.Service, ordersCh chan<- string) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
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
	close(ordersCh)
}

func processOrder(ctx context.Context, svc *service.Service, orderNo string, client *resty.Client, resultCh chan<- string) {
	svc.Log.Info("starting process order", "order_id", orderNo)
	data, err := svc.FetchOrderInfo(ctx, client, orderNo)
	if err != nil {
		svc.Log.Err("failed fetch order info", err)
		return
	}

	err = svc.UpdateOrder(ctx, data)
	if err != nil {
		svc.Log.Err("failed update order", err)
		return
	}
	resultCh <- orderNo
}
