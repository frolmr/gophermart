package domain

type Balance struct {
	BalanceSum    float64 `json:"current"`
	WithdrawalSum float64 `json:"withdrawn"`
}
