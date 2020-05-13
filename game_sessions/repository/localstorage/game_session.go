package localstorage

import (
	"context"
	"errors"
	gamesessions "platform-backend/game_sessions"
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

	return nil, gamesessions.ErrGameSessionNotFound
}

func (r *GameSessionsLocalRepo) GetFirstAction(ctx context.Context, sesID uint64) (*models.GameAction, error) {
	if action, ok := r.firstGameActions[sesID]; ok {
		return action, nil
	}

	return nil, gamesessions.ErrFirstGameActionNotFound
}

func (r *GameSessionsLocalRepo) GetSessionByBlockChainID(ctx context.Context, bcID uint64) (*models.GameSession, error) {
	for _, ses := range r.gameSessions {
		if ses.BlockchainSesID == bcID {
			return toModelGameSession(ses), nil
		}
	}
	return nil, gamesessions.ErrGameSessionNotFound
}

func (r *GameSessionsLocalRepo) UpdateSessionOffset(ctx context.Context, id uint64, offset uint64) error {
	ses, err := r.GetGameSession(ctx, id)
	if err != nil {
		return err
	}

	ses.LastOffset = offset
	return nil
}

func (r *GameSessionsLocalRepo) UpdateSessionState(ctx context.Context, id uint64, newState models.GameSessionState) error {
	ses, err := r.GetGameSession(ctx, id)
	if err != nil {
		return err
	}

	ses.State = newState
	return nil
}

func (r *GameSessionsLocalRepo) AddGameSession(ctx context.Context, ses *models.GameSession) error {
	if _, ok := r.gameSessions[ses.ID]; ok {
		return errors.New("session already exists")
	}
	r.gameSessions[ses.ID] = &GameSession{
		ID:              ses.ID,
		Player:          ses.Player,
		GameID:          ses.GameID,
		CasinoID:        ses.CasinoID,
		BlockchainSesID: ses.BlockchainSesID,
		State:           uint16(ses.State),
		Updates:         make([]*models.GameSessionUpdate, 0, 100),
	}
	return nil
}

func (r *GameSessionsLocalRepo) AddFirstGameAction(ctx context.Context, sesID uint64, action *models.GameAction) error {
	r.firstGameActions[sesID] = action
	return nil
}

func (r *GameSessionsLocalRepo) GetUserGameSessions(ctx context.Context, accountName string) ([]*models.GameSession, error) {
	var sessions []*models.GameSession

	for _, ses := range r.gameSessions {
		if ses.Player == accountName {
			sessions = append(sessions, toModelGameSession(ses))
		}
	}

	return sessions, nil
}

func (r *GameSessionsLocalRepo) GetAllGameSessions(ctx context.Context) ([]*models.GameSession, error) {
	var sessions []*models.GameSession

	for _, ses := range r.gameSessions {
		sessions = append(sessions, toModelGameSession(ses))
	}

	return sessions, nil
}

func (r *GameSessionsLocalRepo) DeleteGameSession(ctx context.Context, id uint64) error {
	delete(r.gameSessions, id)
	return nil
}

func (r *GameSessionsLocalRepo) DeleteFirstGameAction(ctx context.Context, sesID uint64) error {
	delete(r.firstGameActions, sesID)
	return nil
}

func toModelGameSession(gs *GameSession) *models.GameSession {
	return &models.GameSession{
		ID:              gs.ID,
		Player:          gs.Player,
		GameID:          gs.GameID,
		CasinoID:        gs.CasinoID,
		BlockchainSesID: gs.BlockchainSesID,
		State:           models.GameSessionState(gs.State),
	}
}
