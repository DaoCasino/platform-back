package utils

import (
	"github.com/eoscanada/eos-go"
	"regexp"
	"strconv"
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

func ToAsset(value *int64, token string, precision int) *eos.Asset {
	if value == nil {
		return nil
	}
	return &eos.Asset{Amount: eos.Int64(*value), Symbol: eos.MustStringToSymbol(strconv.Itoa(precision) + "," + token)}
}

func ExtractAssetValueAndSymbol(asset *eos.Asset) (int64, string, int) {
	value := int64(asset.Amount)
	symbol := asset.Symbol
	return value, symbol.Symbol, int(asset.Precision)
}

func IsValidEthAddress(v string) bool {
	re := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")
	return re.MatchString(v)
}
