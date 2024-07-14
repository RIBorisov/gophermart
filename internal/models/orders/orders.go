package orders

import "time"

type Status string

const (
	New        Status = "NEW"
	Processing Status = "PROCESSING"
	Invalid    Status = "INVALID"
	Processed  Status = "PROCESSED"
)

type Order struct {
	Status     Status    `json:"status"`
	UploadedAt time.Time `json:"uploaded_at"` // 2020-12-09T16:09:53+03:00
	Number     string    `json:"number"`
	Accrual    float32   `json:"accrual,omitempty"`
}

type UpdateOrder struct {
	Status  Status
	Number  string
	Accrual float32
}
