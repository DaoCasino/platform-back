package models

type ReferralStats struct {
	RollCount  int     `json:"roll_count"`
	DepositSum float64 `json:"deposit_sum"`
	ProfitSum  float64 `json:"profit_sum"`
}
