package utils

import (
	"github.com/eoscanada/eos-go"
	"net/http"
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

type loggingResponseWriter struct {
	http.ResponseWriter
	StatusCode int
}

func NewLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.StatusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
