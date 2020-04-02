package localstorage

import (
	"context"
	"errors"
	"platform-backend/models"
)


func (r *GameSessionsLocalRepo) HasGameSession(ctx context.Context, id uint64) (bool, error) {
	if _, ok := r.gameSessions[id]; ok {
		return true, nil
	}

	return false, nil
}

func (r *GameSessionsLocalRepo) GetGameSession(ctx context.Context, id uint64) (*models.GameSession, error) {
	if ses, ok := r.gameSessions[id]; ok {
		return toModelGameSession(ses), nil
	}

	return nil, errors.New("session not found")
}

func (r *GameSessionsLocalRepo) AddGameSession(ctx context.Context, ses *models.GameSession) error {
	if _, ok := r.gameSessions[ses.ID]; ok {
		return errors.New("session already exists")
	}
	r.gameSessions[ses.ID] = &GameSession{
		ID: ses.ID,
		Player: ses.Player,
		GameID: ses.GameID,
		CasinoID: ses.CasinoID,
		BlockchainSesID: ses.BlockchainSesID,
		State: ses.State,
		Updates: make([]*models.GameSessionUpdate, 0, 100),
	}
	return nil
}

func (r *GameSessionsLocalRepo) DeleteGameSession(ctx context.Context, id uint64) error {
	if _, ok := r.gameSessions[id]; !ok {
		return errors.New("session not found")
	}

	delete(r.gameSessions, id)
	return nil
}

func toModelGameSession(gs *GameSession) *models.GameSession {
	return &models.GameSession{
		ID:              gs.ID,
		Player:          gs.Player,
		GameID:          gs.GameID,
		CasinoID:        gs.CasinoID,
		BlockchainSesID: gs.BlockchainSesID,
		State:           gs.State,
	}
}
