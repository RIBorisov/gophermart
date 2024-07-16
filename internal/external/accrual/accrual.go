package accrual

import (
	"context"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/RIBorisov/gophermart/internal/service"
)

// RunPoller opens chan, starts ticker for the interval and every tick realizes next logic:
// 1. Get orders from database that should be processed
// 2. Pushes it to the chan
// If the chan is not empty, gets new order info from accrual service by the order_id from chan
// and updates order info in database.
func RunPoller(ctx context.Context, svc *service.Service) (chan string, error) {
	aClient := resty.New().SetBaseURL(svc.Config.Service.AccrualSystemAddress)

	interval := svc.Config.Service.AccrualPollInterval
	ordersCh := make(chan string)
	go func() {
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
	}()

	go func() {
		for {
			select {
			case orderNo, ok := <-ordersCh:
				if !ok {
					svc.Log.Debug("channel closed, closing cycle too")
					return
				}
				svc.Log.Debug("fetching info about order", "order_id", orderNo)
				data, err := svc.FetchOrderInfo(ctx, aClient, orderNo)
				if err != nil {
					svc.Log.Err("failed fetch order info", err)
					continue
				}

				svc.Log.Debug("updating info about order", "order_id", orderNo)
				err = svc.UpdateOrder(ctx, data)
				if err != nil {
					svc.Log.Err("failed update order", err)
				}

			case <-ctx.Done():
				svc.Log.Debug("context done...")
				return
			}
		}
	}()

	return ordersCh, nil
}
