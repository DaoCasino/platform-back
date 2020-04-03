package localstorage

import (
	"platform-backend/models"
)

type GameSession struct {
	ID              uint64
	Player          string
	CasinoID        uint64
	GameID          uint64
	BlockchainSesID uint64
	State           uint16
	Updates         []*models.GameSessionUpdate
}

type GameSessionsLocalRepo struct {
	gameSessions       map[uint64]*GameSession
}

func NewGameSessionsPostgresRepo() *GameSessionsLocalRepo {
	return &GameSessionsLocalRepo{
		gameSessions:       make(map[uint64]*GameSession),
	}
}
