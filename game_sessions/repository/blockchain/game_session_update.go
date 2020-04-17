package blockchain

import (
	"context"
	"errors"
	"platform-backend/models"
)

func (r *GameSessionsBCRepo) GetGameSessionUpdates(ctx context.Context, id uint64) ([]*models.GameSessionUpdate, error) {
	return nil, errors.New("not implemented")
}

func (r *GameSessionsBCRepo) AddGameSessionUpdate(ctx context.Context, upd *models.GameSessionUpdate) error {
	return errors.New("not implemented")
}

func (r *GameSessionsBCRepo) DeleteGameSessionUpdates(ctx context.Context, sesId uint64) error {
	return errors.New("not implemented")
}
