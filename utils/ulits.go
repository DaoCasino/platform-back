package utils

import (
	"crypto/rand"
	"math"
	"math/big"
	"net/url"
	"regexp"
	"strconv"

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

func GetRandomUint64() uint64 {
	n, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		panic(err) // panics if math.MaxInt64 <= 0, but it's impossible
	}
	return n.Uint64()
}

func StripQueryString(inputUrl string) (string, error) {
	u, err := url.Parse(inputUrl)
	if err != nil {
		return "", err
	}
	u.RawQuery = ""
	return u.String(), nil
}
