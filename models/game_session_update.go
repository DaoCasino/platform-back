package models

import (
	"context"
	"github.com/randallmlough/pgxscan"
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

func GetGameSessionUpdates(ctx context.Context, id uint64) ([]*GameSessionUpdate, error) {
	conn, err := dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := conn.Query(ctx, selectGameSessionUpdatesByIdStmt, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessionUpdates := make([]*GameSessionUpdate, 0)

	for rows.Next() {
		upd := new(GameSessionUpdate)
		if err := pgxscan.NewScanner(rows).Scan(upd); err != nil {
			return nil, err
		}
		sessionUpdates = append(sessionUpdates, upd)
	}

	return sessionUpdates, nil
}

func AddGameSessionUpdate(ctx context.Context, upd *GameSessionUpdate) error {
	conn, err := dbPool.Acquire(ctx)
	if err != nil {
		return err
	}

	_, err = conn.Exec(ctx, insertGameSessionUpdateStmt, upd.SessionID, upd.UpdateType, upd.Timestamp, upd.Data)
	return err
}

func DeleteGameSessionUpdates(ctx context.Context, sesId uint64) error {
	conn, err := dbPool.Acquire(ctx)
	if err != nil {
		return err
	}

	_, err = conn.Exec(ctx, deleteGameSessionUpdatesByIdStmt, sesId)
	return err
}
