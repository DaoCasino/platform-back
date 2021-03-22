package models

type CashbackRow struct {
	AccountName  string  `db:"account_name"`
	EthAddress   string  `db:"eth_address"`
	PaidCashback float64 `db:"paid_cashback"`
	State        string  `db:"state"`
}

type Cashback struct {
	AccountName string  `json:"accountName"`
	EthAddress  string  `json:"ethAddress"`
	Cashback    float64 `json:"cashback"`
}

type CashbackInfo struct {
	ToPay        float64 `json:"toPay"`
	Paid         float64 `json:"paid"`
	GGR          float64 `json:"ggr"`
	Ratio        float64 `json:"ratio"`
	EthToBetRate float64 `json:"ethToBetRate"`
	State        string  `json:"state"`
}
