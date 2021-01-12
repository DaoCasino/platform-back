package utils

import (
	"github.com/eoscanada/eos-go"
)

const (
	DAOBetAssetSymbol = "BET"
)

func ToBetAsset(deposit string) (*eos.Asset, error) {
	quantity, err := eos.NewAssetFromString(deposit)
	if err != nil {
		return nil, err
	}
	return &quantity, nil
}
