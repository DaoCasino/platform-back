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

func ToAsset(value *int64, symbol string) *eos.Asset {
	if value == nil {
		return nil
	}
	return &eos.Asset{Amount: eos.Int64(*value), Symbol: eos.MustStringToSymbol(symbol)}
}

func ExtractAssetValueAndSymbol(asset *eos.Asset) (int64, string) {
	value := int64(asset.Amount)
	symbol := asset.Symbol.String()
	return value, symbol
}
