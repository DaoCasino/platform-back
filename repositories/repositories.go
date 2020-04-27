package repositories

import (
	"platform-backend/casino"
	"platform-backend/game_sessions"
)

type Repos struct {
	Casino      casino.Repository
	GameSession gamesessions.Repository
}

func NewRepositories(Casino casino.Repository, GameSession gamesessions.Repository) *Repos {
	return &Repos{
		Casino:      Casino,
		GameSession: GameSession,
	}
}
