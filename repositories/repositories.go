package repositories

import (
	"platform-backend/affiliatestats"
	"platform-backend/auth"
	"platform-backend/cashback"
	"platform-backend/contracts"
	gamesessions "platform-backend/game_sessions"
	"platform-backend/location"
)

type Repos struct {
	Contracts      contracts.Repository
	GameSession    gamesessions.Repository
	AffiliateStats affiliatestats.Repository
	Cashback       cashback.Repository
	User           auth.UserRepository
	Location       location.Repository
}

func NewRepositories(
	Contracts contracts.Repository,
	GameSession gamesessions.Repository,
	AffiliateStats affiliatestats.Repository,
	Cashback cashback.Repository,
	User auth.UserRepository,
	Location location.Repository,
) *Repos {
	return &Repos{
		Contracts:      Contracts,
		GameSession:    GameSession,
		AffiliateStats: AffiliateStats,
		Cashback:       Cashback,
		User:           User,
		Location:       Location,
	}
}
