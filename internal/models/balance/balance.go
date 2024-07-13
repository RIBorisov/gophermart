package balance

import "time"

type Response struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type WithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

type Withdrawal struct {
	ProcessedAt time.Time `json:"processed_at"`
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
}
