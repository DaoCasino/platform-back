package models

import "github.com/eoscanada/eos-go"

type GameSessionState uint16

const (
	NewGameTrxSent GameSessionState = iota
	GameStartedInBC
	RequestedGameAction
	GameActionTrxSent
	SignidicePartOneTrxSent
	GameFinished
	GameFailed
)

type GameSession struct {
	ID              uint64           `json:"id"`
	Player          string           `json:"player"`
	CasinoID        uint64           `json:"casinoId"`
	GameID          uint64           `json:"gameId"`
	BlockchainSesID uint64           `json:"blockchainSesId"`
	State           GameSessionState `json:"state"`
	LastOffset      uint64           `json:"lastOffset"`
	Deposit         *eos.Asset       `json:"deposit"`
	LastUpdate      int64            `json:"lastUpdate"`
	PlayerWinAmount *eos.Asset       `json:"playerWinAmount"`
	StateBeforeFail GameSessionState `json:"stateBeforeFail"`
}
