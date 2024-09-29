package domain

import "time"

const (
	Processed  = "PROCESSED"
	Processing = "PROCESSING"
	New        = "NEW"
	Registered = "REGISTERED"
)

type OrderOut struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    *float32  `json:"accrual,omitempty"`
	UserID     int       `json:"-"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type OrderIn struct {
	Number string
}

type OrderOutList []OrderOut
