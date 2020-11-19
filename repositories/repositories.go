package repositories

import (
	"platform-backend/affiliatestats"
	"platform-backend/contracts"
	"platform-backend/game_sessions"
)

type Repos struct {
	Contracts      contracts.Repository
	GameSession    gamesessions.Repository
	AffiliateStats affiliatestats.Repository
}

func NewRepositories(
	Contracts contracts.Repository,
	GameSession gamesessions.Repository,
	AffiliateStats affiliatestats.Repository,
) *Repos {
	return &Repos{
		Contracts:      Contracts,
		GameSession:    GameSession,
		AffiliateStats: AffiliateStats,
	}
}
