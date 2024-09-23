package domain

type AccrualOut struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float32 `json:"accrual,omitempty"`
}

type AccrualIn struct {
	Order string `json:"order"`
}
