package models

type GameSession struct {
	ID              uint64
	Player          string
	CasinoID        uint64
	GameID          uint64
	BlockchainSesID uint64
	State           uint16
}
