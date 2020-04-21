package models

type GameSessionState uint16

const (
	NewGameTrxSent GameSessionState = iota
	GameStartedInBC
	RequestedGameAction
	GameActionTrxSent
	RequestedSignidicePartOne
	SignidicePartOneTrxSent
	GameFinished
	GameFailed
)

type GameSession struct {
	ID              uint64
	Player          string
	CasinoID        uint64
	GameID          uint64
	BlockchainSesID uint64
	State           GameSessionState
	LastOffset 		uint64
}
