package blockchain

import (
	"context"
	"errors"
	"platform-backend/models"
)

func (r *GameSessionsBCRepo) HasGameSession(ctx context.Context, id uint64) (bool, error) {
	return false, errors.New("not implemented")
}

func (r *GameSessionsBCRepo) GetGameSession(ctx context.Context, id uint64) (*models.GameSession, error) {
	return nil, errors.New("not implemented")
}

func (r *GameSessionsBCRepo) AddGameSession(ctx context.Context, ses *models.GameSession) error {
	return errors.New("not implemented")
}

func (r *GameSessionsBCRepo) DeleteGameSession(ctx context.Context, id uint64) error {
	return errors.New("not implemented")
}
