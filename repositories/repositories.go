package repositories

import (
	"platform-backend/contracts"
	"platform-backend/game_sessions"
)

type Repos struct {
	Contracts   contracts.Repository
	GameSession gamesessions.Repository
}

func NewRepositories(Contracts contracts.Repository, GameSession gamesessions.Repository) *Repos {
	return &Repos{
		Contracts:   Contracts,
		GameSession: GameSession,
	}
}
