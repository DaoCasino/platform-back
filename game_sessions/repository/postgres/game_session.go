package postgres

import (
	"context"
	"platform-backend/db"
	"platform-backend/models"
)

const (
	selectGameSessionByIdStmt    = "SELECT * FROM game_sessions WHERE id = $1"
	selectGameSessionByBcID      = "SELECT * FROM game_sessions WHERE blockchain_req_id = $1"
	selectAllGameSessions        = "SELECT * FROM game_sessions WHERE player = $1"
	updateSessionState           = "UPDATE game_sessions SET state = $2 WHERE id = $1"
	updateSessionOffset          = "UPDATE game_sessions SET last_offset = $2 WHERE id = $1"
	selectGameSessionCntByIdStmt = "SELECT count(*) FROM game_sessions WHERE id = $1"
	insertGameSessionStmt        = "INSERT INTO game_sessions VALUES ($1, $2, $3, $4, $5, $6, $7)"
	deleteGameSessionByIdStmt    = "DELETE FROM game_sessions WHERE id = $1"
)

type GameSession struct {
	ID              uint64 `db:"id"`
	Player          string `db:"player"`
	GameID          uint64 `db:"game_id"`
	CasinoID        uint64 `db:"casino_id"`
	BlockchainSesID uint64 `db:"blockchain_ses_id"`
	State           uint16 `db:"state"`
	LastOffset      uint64 `db:"last_offset"`
}

func (r *GameSessionsPostgresRepo) HasGameSession(ctx context.Context, id uint64) (bool, error) {
	conn, err := db.DbPool.Acquire(ctx)
	if err != nil {
		return false, err
	}
	defer conn.Release()

	var cnt uint
	err = conn.QueryRow(ctx, selectGameSessionCntByIdStmt, id).Scan(&cnt)
	if err != nil {
		return false, err
	}

	return cnt > 0, nil
}

func (r *GameSessionsPostgresRepo) GetGameSession(ctx context.Context, id uint64) (*models.GameSession, error) {
	conn, err := db.DbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	session := new(GameSession)
	err = conn.QueryRow(ctx, selectGameSessionByIdStmt, id).Scan(
		&session.ID,
		&session.Player,
		&session.GameID,
		&session.CasinoID,
		&session.BlockchainSesID,
		&session.State,
		&session.LastOffset,
	)

	if err != nil {
		return nil, err
	}
	return toModelGameSession(session), nil
}

func (r *GameSessionsPostgresRepo) GetSessionByBlockChainID(ctx context.Context, bcID uint64) (*models.GameSession, error) {
	conn, err := db.DbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	session := new(GameSession)
	err = conn.QueryRow(ctx, selectGameSessionByBcID, bcID).Scan(
		&session.ID,
		&session.Player,
		&session.GameID,
		&session.CasinoID,
		&session.BlockchainSesID,
		&session.State,
		&session.LastOffset,
	)

	if err != nil {
		return nil, err
	}
	return toModelGameSession(session), nil
}

func (r *GameSessionsPostgresRepo) UpdateSessionOffset(ctx context.Context, id uint64, offset uint64) error {
	conn, err := db.DbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, updateSessionOffset, id, offset)
	return err
}

func (r *GameSessionsPostgresRepo) UpdateSessionState(ctx context.Context, id uint64, newState models.GameSessionState) error {
	conn, err := db.DbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, updateSessionState, id, uint16(newState))
	return err
}

func (r *GameSessionsPostgresRepo) AddGameSession(ctx context.Context, ses *models.GameSession) error {
	conn, err := db.DbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, insertGameSessionStmt, ses.ID, ses.Player, ses.GameID, ses.CasinoID, ses.BlockchainSesID, ses.State, ses.LastOffset)
	if err != nil {
		return err
	}
	return nil
}

func (r *GameSessionsPostgresRepo) GetAllGameSessions(ctx context.Context, accountName string) ([]*models.GameSession, error) {
	conn, err := db.DbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, selectAllGameSessions, accountName)
	if err != nil {
		return nil, err
	}

	gameSessions := make([]*models.GameSession, 0)
	for rows.Next() {
		session := new(GameSession)
		err = rows.Scan(
			&session.ID,
			&session.Player,
			&session.GameID,
			&session.CasinoID,
			&session.BlockchainSesID,
			&session.State,
			&session.LastOffset,
		)
		if err != nil {
			return nil, err
		}
		gameSessions = append(gameSessions, toModelGameSession(session))
	}

	return gameSessions, nil
}

func (r *GameSessionsPostgresRepo) DeleteGameSession(ctx context.Context, id uint64) error {
	conn, err := db.DbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, deleteGameSessionByIdStmt, id)
	return err
}

func toModelGameSession(gs *GameSession) *models.GameSession {
	return &models.GameSession{
		ID:              gs.ID,
		Player:          gs.Player,
		GameID:          gs.GameID,
		CasinoID:        gs.CasinoID,
		BlockchainSesID: gs.BlockchainSesID,
		State:           models.GameSessionState(gs.State),
		LastOffset:      gs.LastOffset,
	}
}
