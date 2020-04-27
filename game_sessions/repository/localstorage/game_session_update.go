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

	sessionUpdates := make([]*models.GameSessionUpdate, 0)
	ses := r.gameSessions[id]
	if ses.Updates != nil {
		sessionUpdates = append(sessionUpdates, ses.Updates...)
	}

	return sessionUpdates, nil
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
