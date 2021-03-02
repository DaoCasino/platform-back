package repositories

import (
	"platform-backend/affiliatestats"
	"platform-backend/auth"
	"platform-backend/cashback"
	"platform-backend/contracts"
	"platform-backend/game_sessions"
)

type Repos struct {
	Contracts      contracts.Repository
	GameSession    gamesessions.Repository
	AffiliateStats affiliatestats.Repository
	Cashback       cashback.Repository
	User           auth.UserRepository
}

func NewRepositories(
	Contracts contracts.Repository,
	GameSession gamesessions.Repository,
	AffiliateStats affiliatestats.Repository,
	Cashback cashback.Repository,
	User auth.UserRepository,
) *Repos {
	return &Repos{
		Contracts:      Contracts,
		GameSession:    GameSession,
		AffiliateStats: AffiliateStats,
		Cashback:       Cashback,
		User:           User,
	}
}
