package accrual

import (
	"github.com/RIBorisov/gophermart/internal/errs"
	"github.com/RIBorisov/gophermart/internal/models/orders"
)

type Status string

const (
	Registered = "REGISTERED"
	Invalid    = "INVALID"
	Processing = "PROCESSING"
	Processed  = "PROCESSED"
)

type OrderInfoResponse struct {
	Order   string  `json:"order"`
	Status  Status  `json:"status"`
	Accrual float64 `json:"accrual"`
}

// ConvertToOrderStatus converts Accrual order status into order status.
func (s Status) ConvertToOrderStatus() (orders.Status, error) {
	switch s {
	case Registered:
		return orders.New, nil
	case Invalid:
		return orders.Invalid, nil
	case Processing:
		return orders.Processing, nil
	case Processed:
		return orders.Processed, nil
	default:
		return "", errs.ErrInvalidAccrualStatus
	}
}
