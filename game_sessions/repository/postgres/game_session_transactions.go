package postgres

import (
	"context"

	"platform-backend/db"
)

const (
	InsertGameSessionTransaction = `
        INSERT INTO game_session_txns 
            (trx_id, ses_id, action_type, action_params)
        VALUES
            ($1, $2, $3, $4)`
)

func (r *GameSessionsPostgresRepo) AddGameSessionTransaction(ctx context.Context, trxID string, sesID uint64,
	actionType uint16, actionParams []uint64) error {
	conn, err := db.DbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	_, err = conn.Exec(ctx, InsertGameSessionTransaction, trxID, sesID, actionType, actionParams)
	return err
}
