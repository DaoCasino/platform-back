package postgres

import (
	"context"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"platform-backend/db"
	gamesessions "platform-backend/game_sessions"
	"platform-backend/models"
	"time"
)

const (
	selectGameSessionUpdatesByIdStmt = "SELECT * FROM game_session_updates WHERE ses_id = $1 ORDER BY timestamp ASC"
	insertGameSessionUpdateStmt      = "INSERT INTO game_session_updates VALUES ($1, $2, $3, $4, $5)"
	deleteGameSessionUpdatesByIdStmt = "DELETE FROM game_session_updates WHERE ses_id = $1"

	sqlDuplicateUniqueErrorCode = "23505"
)

type GameSessionUpdate struct {
	SessionID  uint64    `db:"ses_id"`
	UpdateType uint16    `db:"update_type"`
	Timestamp  time.Time `db:"timestamp"`
	Data       []byte    `db:"data"`
	Offset     *uint64   `db:"offset"`
}

func (u *GameSessionUpdate) Scan(row pgx.Row) error {
	return row.Scan(
		&u.SessionID,
		&u.UpdateType,
		&u.Timestamp,
		&u.Data,
		&u.Offset,
	)
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
		err := upd.Scan(rows)
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

	_, err = conn.Exec(ctx, insertGameSessionUpdateStmt, upd.SessionID, upd.UpdateType, upd.Timestamp, upd.Data, upd.Offset)
	if pgErr, ok := err.(*pgconn.PgError); ok {
		if pgErr.Code == sqlDuplicateUniqueErrorCode {
			return gamesessions.ErrUpdateAlreadyProcessed
		}
	}
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
		Offset:     gsu.Offset,
	}
}
