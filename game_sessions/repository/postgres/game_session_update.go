package postgres

import (
	"context"
	"platform-backend/db"
	"platform-backend/models"
	"time"
)

const (
	selectGameSessionUpdatesByIdStmt = "SELECT * FROM game_session_updates WHERE ses_id = $1"
	insertGameSessionUpdateStmt      = "INSERT INTO game_session_updates VALUES ($1, $2, $3, $4)"
	deleteGameSessionUpdatesByIdStmt = "DELETE FROM game_session_updates WHERE ses_id = $1"
)

type GameSessionUpdate struct {
	SessionID  uint64    `db:"ses_id"`
	UpdateType uint16    `db:"update_type"`
	Timestamp  time.Time `db:"timestamp"`
	Data       []byte    `db:"data"`
}

func (r *GameSessionsPostgresRepo) GetGameSessionUpdates(ctx context.Context, id uint64) ([]*models.GameSessionUpdate, error) {
	conn, err := db.DbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, selectGameSessionUpdatesByIdStmt, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessionUpdates := make([]*models.GameSessionUpdate, 0)

	for rows.Next() {
		upd := new(GameSessionUpdate)
		err := rows.Scan(
			&upd.SessionID,
			&upd.UpdateType,
			&upd.Timestamp,
			&upd.Data,
		)
		if err != nil {
			return nil, err
		}
		sessionUpdates = append(sessionUpdates, toModelGameSessionUpdate(upd))
	}

	return sessionUpdates, nil
}

func (r *GameSessionsPostgresRepo) AddGameSessionUpdate(ctx context.Context, upd *models.GameSessionUpdate) error {
	conn, err := db.DbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, insertGameSessionUpdateStmt, upd.SessionID, upd.UpdateType, upd.Timestamp, upd.Data)
	return err
}

func (r *GameSessionsPostgresRepo) DeleteGameSessionUpdates(ctx context.Context, sesId uint64) error {
	conn, err := db.DbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, deleteGameSessionUpdatesByIdStmt, sesId)
	return err
}

func toModelGameSessionUpdate(gsu *GameSessionUpdate) *models.GameSessionUpdate {
	return &models.GameSessionUpdate{
		SessionID:  gsu.SessionID,
		UpdateType: models.GameSessionUpdateType(gsu.UpdateType),
		Timestamp:  gsu.Timestamp,
		Data:       gsu.Data,
	}
}
