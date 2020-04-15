package models

import (
	"github.com/eoscanada/eos-go"
)

type PlayerInfo struct {
	Balance          eos.Asset      `json:"balance"`
	ActivePermission *eos.Authority `json:"activePermission"`
	OwnerPermission  *eos.Authority `json:"ownerPermission"`
	LinkedCasinos    []*Casino      `json:"linkedCasinos"`
}
