package domain

type BalanceOut struct {
	Current   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

type WithdrawIn struct {
	OrderNumber string  `json:"order"`
	Sum         float32 `json:"sum"`
}

type WithdrawalsOut struct {
	Order       string  `json:"order"`
	Sum         float32 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

type WithdrawOutList []WithdrawalsOut
