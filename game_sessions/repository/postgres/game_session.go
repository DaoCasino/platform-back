package postgres

import (
	"context"
	"errors"
	"github.com/eoscanada/eos-go"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"platform-backend/db"
	gamesessions "platform-backend/game_sessions"
	"platform-backend/models"
	"platform-backend/utils"
	"time"
)

const GlobalCnt = 30

const (
	selectGameSessionByIdStmt    = "SELECT * FROM game_sessions WHERE id = $1"
	selectGameSessionByBcIDStmt  = "SELECT * FROM game_sessions WHERE blockchain_req_id = $1"
	selectUserGameSessionsStmt   = "SELECT * FROM game_sessions WHERE player = $1 ORDER BY last_update DESC"
	selectGlobalSessionsStmt     = "SELECT * FROM game_sessions WHERE state = $1 ORDER BY last_update DESC LIMIT $2"
	selectGlobalSessionsLostStmt = "SELECT * FROM game_sessions WHERE state = $1 AND player_win_amount SIMILAR TO '-%' ORDER BY last_update DESC LIMIT $2"
	selectGlobalSessionsWinsStmt = "SELECT * FROM game_sessions WHERE state = $1 AND player_win_amount NOT SIMILAR TO '(-%|0.0000 %)' ORDER BY last_update DESC LIMIT $2"
	selectCasinoSessionsStmt     = "SELECT * FROM game_sessions WHERE state = $1 AND casino_id = $2 ORDER BY last_update DESC LIMIT $3"
	selectCasinoSessionsLostStmt = "SELECT * FROM game_sessions WHERE state = $1 AND casino_id = $2 AND player_win_amount SIMILAR TO '-%' ORDER BY last_update DESC LIMIT $3"
	selectCasinoSessionsWinsStmt = "SELECT * FROM game_sessions WHERE state = $1 AND casino_id = $2 AND player_win_amount NOT SIMILAR TO '(-%|0.0000 %)' ORDER BY last_update DESC LIMIT $3"
	selectAllGameSessionsStmt    = "SELECT * FROM game_sessions"
	selectFirstGameActionStmt    = "SELECT * FROM first_game_actions WHERE ses_id = $1"
	updateSessionStateStmt       = "UPDATE game_sessions SET state = $2, last_update = $3 WHERE id = $1"
	updateSessionDepositStmt     = "UPDATE game_sessions SET deposit = $2 WHERE id = $1"
	updateSessionPlayerWinStmt   = "UPDATE game_sessions SET player_win_amount = $2 WHERE id = $1"
	updateSessionOffsetStmt      = "UPDATE game_sessions SET last_offset = $2 WHERE id = $1"
	selectGameSessionCntByIdStmt = "SELECT count(*) FROM game_sessions WHERE id = $1"
	insertGameSessionStmt        = "INSERT INTO game_sessions VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)"
	insertFirstGameActionStmt    = "INSERT INTO first_game_actions VALUES ($1, $2, $3)"
	deleteGameSessionByIdStmt    = "DELETE FROM game_sessions WHERE id = $1"
	deleteFirstGameActionStmt    = "DELETE FROM first_game_actions WHERE ses_id = $1"
)

type GameSession struct {
	ID              uint64  `db:"id"`
	Player          string  `db:"player"`
	GameID          uint64  `db:"game_id"`
	CasinoID        uint64  `db:"casino_id"`
	BlockchainSesID uint64  `db:"blockchain_ses_id"`
	State           uint16  `db:"state"`
	LastOffset      uint64  `db:"last_offset"`
	Deposit         *string `db:"deposit"`
	LastUpdate      int64   `db:"last_update"`
	PlayerWinAmount *string `db:"player_win_amount"`
}

func (s *GameSession) Scan(row pgx.Row) error {
	return row.Scan(
		&s.ID,
		&s.Player,
		&s.GameID,
		&s.CasinoID,
		&s.BlockchainSesID,
		&s.State,
		&s.LastOffset,
		&s.Deposit,
		&s.LastUpdate,
		&s.PlayerWinAmount,
	)
}

type GameAction struct {
	SesID  uint64              `db:"ses_id"`
	Type   uint16              `db:"type"`
	Params pgtype.NumericArray `db:"params"`
}

func (a *GameAction) Scan(row pgx.Row) error {
	return row.Scan(
		&a.SesID,
		&a.Type,
		&a.Params,
	)
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
	err = session.Scan(conn.QueryRow(ctx, selectGameSessionByIdStmt, id))

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, gamesessions.ErrGameSessionNotFound
		}
		return nil, err
	}
	return toModelGameSession(session)
}

func (r *GameSessionsPostgresRepo) GetCasinoSessions(ctx context.Context, filter gamesessions.FilterType, casinoId eos.Uint64) ([]*models.GameSession, error) {
	conn, err := db.DbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	var rows pgx.Rows
	switch filter {
	case gamesessions.All:
		rows, err = conn.Query(ctx, selectCasinoSessionsStmt, models.GameFinished, casinoId, GlobalCnt)
	case gamesessions.Wins:
		rows, err = conn.Query(ctx, selectCasinoSessionsWinsStmt, models.GameFinished, casinoId, GlobalCnt)
	case gamesessions.Losts:
		rows, err = conn.Query(ctx, selectCasinoSessionsLostStmt, models.GameFinished, casinoId, GlobalCnt)
	default:
		return nil, errors.New("bad filter")
	}

	if err != nil {
		return nil, err
	}

	gameSessions := make([]*models.GameSession, 0, GlobalCnt)
	for rows.Next() {
		session := new(GameSession)
		err = session.Scan(rows)
		if err != nil {
			return nil, err
		}
		ses, err := toModelGameSession(session)
		if err != nil {
			return nil, err
		}
		gameSessions = append(gameSessions, ses)
	}

	return gameSessions, nil
}

func (r *GameSessionsPostgresRepo) GetGlobalSessions(ctx context.Context, filter gamesessions.FilterType) ([]*models.GameSession, error) {
	conn, err := db.DbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	var rows pgx.Rows
	switch filter {
	case gamesessions.All:
		rows, err = conn.Query(ctx, selectGlobalSessionsStmt, models.GameFinished, GlobalCnt)
	case gamesessions.Wins:
		rows, err = conn.Query(ctx, selectGlobalSessionsWinsStmt, models.GameFinished, GlobalCnt)
	case gamesessions.Losts:
		rows, err = conn.Query(ctx, selectGlobalSessionsLostStmt, models.GameFinished, GlobalCnt)
	default:
		return nil, errors.New("bad filter")
	}

	if err != nil {
		return nil, err
	}

	gameSessions := make([]*models.GameSession, 0, GlobalCnt)
	for rows.Next() {
		session := new(GameSession)
		err = session.Scan(rows)
		if err != nil {
			return nil, err
		}
		ses, err := toModelGameSession(session)
		if err != nil {
			return nil, err
		}
		gameSessions = append(gameSessions, ses)
	}

	return gameSessions, nil
}

func (r *GameSessionsPostgresRepo) GetFirstAction(ctx context.Context, sesID uint64) (*models.GameAction, error) {
	conn, err := db.DbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	var action GameAction
	err = action.Scan(conn.QueryRow(ctx, selectFirstGameActionStmt, sesID))

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, gamesessions.ErrFirstGameActionNotFound
		}
		return nil, err
	}

	return toModelGameAction(&action)
}

func (r *GameSessionsPostgresRepo) GetSessionByBlockChainID(ctx context.Context, bcID uint64) (*models.GameSession, error) {
	conn, err := db.DbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	session := new(GameSession)
	err = session.Scan(conn.QueryRow(ctx, selectGameSessionByBcIDStmt, bcID))

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, gamesessions.ErrGameSessionNotFound
		}
		return nil, err
	}
	return toModelGameSession(session)
}

func (r *GameSessionsPostgresRepo) UpdateSessionOffset(ctx context.Context, id uint64, offset uint64) error {
	conn, err := db.DbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, updateSessionOffsetStmt, id, offset)
	return err
}

func (r *GameSessionsPostgresRepo) UpdateSessionDeposit(ctx context.Context, id uint64, deposit string) error {
	conn, err := db.DbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, updateSessionDepositStmt, id, deposit)
	return err
}

func (r *GameSessionsPostgresRepo) UpdateSessionPlayerWin(ctx context.Context, id uint64, playerWin string) error {
	conn, err := db.DbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, updateSessionPlayerWinStmt, id, playerWin)
	return err
}

func (r *GameSessionsPostgresRepo) UpdateSessionState(ctx context.Context, id uint64, newState models.GameSessionState) error {
	conn, err := db.DbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, updateSessionStateStmt, id, uint16(newState), time.Now().Unix())
	return err
}

func (r *GameSessionsPostgresRepo) AddGameSession(ctx context.Context, ses *models.GameSession) error {
	conn, err := db.DbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, insertGameSessionStmt,
		ses.ID,
		ses.Player,
		ses.GameID,
		ses.CasinoID,
		ses.BlockchainSesID,
		ses.State,
		ses.LastOffset,
		ses.Deposit.String(),
		ses.LastUpdate,
		nil,
	)
	if err != nil {
		return err
	}
	return nil
}

func (r *GameSessionsPostgresRepo) AddFirstGameAction(ctx context.Context, sesID uint64, action *models.GameAction) error {
	conn, err := db.DbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	params := pgtype.NumericArray{}
	err = params.Set(action.Params)
	if err != nil {
		return err
	}

	_, err = conn.Exec(ctx, insertFirstGameActionStmt, sesID, action.Type, params)
	if err != nil {
		return err
	}
	return nil
}

func (r *GameSessionsPostgresRepo) GetUserGameSessions(ctx context.Context, accountName string) ([]*models.GameSession, error) {
	conn, err := db.DbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, selectUserGameSessionsStmt, accountName)
	if err != nil {
		return nil, err
	}

	gameSessions := make([]*models.GameSession, 0)
	for rows.Next() {
		session := new(GameSession)
		err = session.Scan(rows)
		if err != nil {
			return nil, err
		}
		ses, err := toModelGameSession(session)
		if err != nil {
			return nil, err
		}
		gameSessions = append(gameSessions, ses)
	}

	return gameSessions, nil
}

func (r *GameSessionsPostgresRepo) GetAllGameSessions(ctx context.Context) ([]*models.GameSession, error) {
	conn, err := db.DbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, selectAllGameSessionsStmt)
	if err != nil {
		return nil, err
	}

	gameSessions := make([]*models.GameSession, 0)
	for rows.Next() {
		session := new(GameSession)
		err = session.Scan(rows)
		if err != nil {
			return nil, err
		}
		ses, err := toModelGameSession(session)
		if err != nil {
			return nil, err
		}
		gameSessions = append(gameSessions, ses)
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

func (r *GameSessionsPostgresRepo) DeleteFirstGameAction(ctx context.Context, sesID uint64) error {
	conn, err := db.DbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, deleteFirstGameActionStmt, sesID)
	return err
}

func toModelGameSession(gs *GameSession) (*models.GameSession, error) {
	ses := &models.GameSession{
		ID:              gs.ID,
		Player:          gs.Player,
		GameID:          gs.GameID,
		CasinoID:        gs.CasinoID,
		BlockchainSesID: gs.BlockchainSesID,
		State:           models.GameSessionState(gs.State),
		LastOffset:      gs.LastOffset,
		LastUpdate:      gs.LastUpdate,
	}

	if gs.Deposit == nil {
		ses.Deposit = nil
	} else {
		deposit, err := utils.ToBetAsset(*gs.Deposit)
		if err != nil {
			return nil, err
		}
		ses.Deposit = deposit
	}

	if gs.PlayerWinAmount == nil {
		ses.PlayerWinAmount = nil
	} else {
		winAmount, err := utils.ToBetAsset(*gs.PlayerWinAmount)
		if err != nil {
			return nil, err
		}
		ses.PlayerWinAmount = winAmount
	}

	return ses, nil
}

func toModelGameAction(ga *GameAction) (*models.GameAction, error) {
	ret := models.GameAction{
		Type: ga.Type,
	}
	err := ga.Params.AssignTo(&ret.Params)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}
