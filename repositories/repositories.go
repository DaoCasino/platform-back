package repositories

import (
	"platform-backend/affiliatestats"
	"platform-backend/cashback"
	"platform-backend/contracts"
	"platform-backend/game_sessions"
)

type Repos struct {
	Contracts      contracts.Repository
	GameSession    gamesessions.Repository
	AffiliateStats affiliatestats.Repository
	Cashback       cashback.Repository
}

func NewRepositories(
	Contracts contracts.Repository,
	GameSession gamesessions.Repository,
	AffiliateStats affiliatestats.Repository,
	Cashback cashback.Repository,
) *Repos {
	return &Repos{
		Contracts:      Contracts,
		GameSession:    GameSession,
		AffiliateStats: AffiliateStats,
		Cashback:       Cashback,
	}
}
