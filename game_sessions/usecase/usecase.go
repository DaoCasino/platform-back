package usecase

import (
	"context"
	"platform-backend/casino"
	"platform-backend/game_sessions"
	"platform-backend/models"
)

type GameSessionsUseCase struct {
	repo       gamesessions.Repository
	casinoRepo casino.Repository
}

func NewGameSessionsUseCase(repo gamesessions.Repository, casinoRepo casino.Repository) *GameSessionsUseCase {
	return &GameSessionsUseCase{repo: repo, casinoRepo: casinoRepo}
}

func (a *GameSessionsUseCase) NewSession(ctx context.Context, GameId uint64, CasinoID uint64, Deposit string, User *models.User) (*models.GameSession, error) {
	game, err := a.casinoRepo.GetGame(ctx, GameId)
	if err != nil {
		return nil, err
	}

	cas, err := a.casinoRepo.GetCasino(ctx, CasinoID)
	if err != nil {
		return nil, err
	}

	session, err := a.repo.AddGameSession(ctx, cas, game, User, Deposit)
	if err != nil {
		return nil, err
	}

	return session, nil
}
func (a *GameSessionsUseCase) HasGameSession(ctx context.Context, id uint64) (bool, error) {
	return a.repo.HasGameSession(ctx, id)
}
func (a *GameSessionsUseCase) GetGameSession(ctx context.Context, id uint64) (*models.GameSession, error) {
	return a.repo.GetGameSession(ctx, id)
}
