package models

import (
	"github.com/eoscanada/eos-go"
)

type PlayerInfo struct {
	Balance          eos.Asset       `json:"balance"`
	BonusBalances    []*BonusBalance `json:"bonusBalances"`
	ActivePermission eos.Authority   `json:"activePermission"`
	OwnerPermission  eos.Authority   `json:"ownerPermission"`
	LinkedCasinos    []*Casino       `json:"linkedCasinos"`
}

type BonusBalance struct {
	Balance  eos.Asset `json:"balance"`
	CasinoId uint64    `json:"casinoId"`
}
