package localstorage

import (
	"context"
	"platform-backend/models"
)

func (r *GameSessionsLocalRepo) GetGameSessionUpdates(ctx context.Context, id uint64) ([]*models.GameSessionUpdate, error) {
	_, err := r.GetGameSession(ctx, id)
	if err != nil {
		return nil, err
	}

	ses := r.gameSessions[id]

	if ses.Updates != nil {
		return ses.Updates, nil
	}

	return make([]*models.GameSessionUpdate, 0), nil
}

func (r *GameSessionsLocalRepo) AddGameSessionUpdate(ctx context.Context, upd *models.GameSessionUpdate) error {
	_, err := r.GetGameSession(ctx, upd.SessionID)
	if err != nil {
		return err
	}

	ses := r.gameSessions[upd.SessionID]
	ses.Updates = append(ses.Updates, upd)

	return nil
}

func (r *GameSessionsLocalRepo) DeleteGameSessionUpdates(ctx context.Context, sesId uint64) error {
	_, err := r.GetGameSession(ctx, sesId)
	if err != nil {
		return err
	}

	ses := r.gameSessions[sesId]
	ses.Updates = nil
	return err
}
