package models

import "github.com/eoscanada/eos-go"

type ReferralStats struct {
	ReferralStatsRow
	Data map[string]ReferralStatsRow
}

type ReferralStatsRow struct {
	Rolls           int       `json:"roll_count"`
	DepositSum      float64   `json:"deposit_sum"`
	ProfitSum       float64   `json:"profit_sum"`
	DepositSumAsset eos.Asset `json:"deposit_sum_asset"`
	ProfitSumAsset  eos.Asset `json:"profit_sum_asset"`
}
