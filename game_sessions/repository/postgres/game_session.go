package postgres

import (
	"context"
	"github.com/randallmlough/pgxscan"
	"platform-backend/db"
	"platform-backend/models"
)

const (
	selectGameSessionByIdStmt    = "SELECT * FROM game_sessions WHERE id = $1"
	selectGameSessionCntByIdStmt = "SELECT count(*) FROM game_sessions WHERE id = $1"
	insertGameSessionStmt        = "INSERT INTO game_sessions VALUES ($1, $2, $3, $4, $5, $6)"
	deleteGameSessionByIdStmt    = "DELETE FROM game_sessions WHERE id = $1"
)

type GameSession struct {
	ID              uint64 `db:"id"`
	Player          string `db:"player"`
	CasinoID        uint64 `db:"casino_id"`
	GameID          uint64 `db:"game_id"`
	BlockchainSesID uint64 `db:"blockchain_ses_id"`
	State           uint16 `db:"state"`
}

func (r *GameSessionsPostgresRepo) HasGameSession(ctx context.Context, id uint64) (bool, error) {
	conn, err := db.DbPool.Acquire(ctx)
	if err != nil {
		return false, err
	}

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

	session := new(GameSession)
	row := conn.QueryRow(ctx, selectGameSessionByIdStmt, id)
	err = pgxscan.NewScanner(row).Scan(session)

	if err != nil {
		return nil, err
	}
	return toModelGameSession(session), nil
}

func (r *GameSessionsPostgresRepo) AddGameSession(ctx context.Context, ses *models.GameSession) error {
	conn, err := db.DbPool.Acquire(ctx)
	if err != nil {
		return err
	}

	_, err = conn.Exec(ctx, insertGameSessionStmt, ses.ID, ses.Player, ses.CasinoID, ses.GameID, ses.BlockchainSesID, ses.State)
	return err
}

func (r *GameSessionsPostgresRepo) DeleteGameSession(ctx context.Context, id uint64) error {
	conn, err := db.DbPool.Acquire(ctx)
	if err != nil {
		return err
	}

	_, err = conn.Exec(ctx, deleteGameSessionByIdStmt, id)
	return err
}

func toPostgresGameSession(gs *models.GameSession) *GameSession {
	return &GameSession{
		ID:              gs.ID,
		Player:          gs.Player,
		GameID:          gs.GameID,
		CasinoID:        gs.CasinoID,
		BlockchainSesID: gs.BlockchainSesID,
		State:           gs.State,
	}
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
