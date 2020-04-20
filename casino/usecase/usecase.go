package usecase

import (
	"context"
	"platform-backend/casino"
	"platform-backend/models"
)

type CasinoUseCase struct {
	casinoRepo casino.Repository
}

func NewCasinoUseCase(casinoRepo casino.Repository) *CasinoUseCase {
	return &CasinoUseCase{
		casinoRepo: casinoRepo,
	}
}

func (a *CasinoUseCase) GetCasino(ctx context.Context, casinoId uint64) (*models.Casino, error) {
	return a.casinoRepo.GetCasino(ctx, casinoId)
}

func (a *CasinoUseCase) AllCasinos(ctx context.Context) ([]*models.Casino, error) {
	return a.casinoRepo.AllCasinos(ctx)
}

func (a *CasinoUseCase) GetCasinoGames(ctx context.Context, casinoName string) ([]*models.CasinoGame, error) {
	return a.casinoRepo.GetCasinoGames(ctx, casinoName)
}


func (a *CasinoUseCase) GetGame(ctx context.Context, gameId uint64) (*models.Game, error) {
	return a.casinoRepo.GetGame(ctx, gameId)
}

func (a *CasinoUseCase) AllGames(ctx context.Context) ([]*models.Game, error) {
	return a.casinoRepo.AllGames(ctx)
}