package utils

import (
	"github.com/eoscanada/eos-go"
)

const (
	DAOBetAssetSymbol = "BET"
)

func ToBetAsset(deposit string) (*eos.Asset, error) {
	quantity, err := eos.NewFixedSymbolAssetFromString(eos.Symbol{Precision: 4, Symbol: DAOBetAssetSymbol}, deposit)
	if err != nil {
		return nil, err
	}
	return &quantity, nil
}

func ToAsset(value *float64, symbol string) *eos.Asset {
	return &eos.Asset{
		Amount: eos.Int64(*value * 10000),
		Symbol: eos.Symbol{
			Precision: 4,
			Symbol:    symbol,
		},
	}
}

func ExtractAssetValueAndSymbol(asset *eos.Asset) (float64, string) {
	value := float64(asset.Amount) / 10000
	symbol := asset.Symbol.MustSymbolCode().String()
	return value, symbol
}
