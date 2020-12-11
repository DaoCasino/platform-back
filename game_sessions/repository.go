package gamesessions

import (
	"context"
	"github.com/eoscanada/eos-go"
	"platform-backend/models"
)

type FilterType string

const (
	All   FilterType = "all"
	Wins  FilterType = "wins"
	Losts FilterType = "losts"
)

type Repository interface {
	HasGameSession(ctx context.Context, id uint64) (bool, error)
	GetGameSession(ctx context.Context, id uint64) (*models.GameSession, error)
	GetGlobalSessions(ctx context.Context, filter FilterType) ([]*models.GameSession, error)
	GetSessionByBlockChainID(ctx context.Context, bcID uint64) (*models.GameSession, error)
	UpdateSessionState(ctx context.Context, id uint64, state models.GameSessionState) error
	UpdateSessionStateBeforeFail(ctx context.Context, id uint64, prevState models.GameSessionState) error
	UpdateSessionOffset(ctx context.Context, id uint64, offset uint64) error
	UpdateSessionPlayerWin(ctx context.Context, id uint64, playerWin string, value int64) error
	UpdateSessionDeposit(ctx context.Context, id uint64, deposit string, symbol string, value int64) error
	AddGameSession(ctx context.Context, ses *models.GameSession) error
	GetUserGameSessions(ctx context.Context, accountName string) ([]*models.GameSession, error)
	GetAllGameSessions(ctx context.Context) ([]*models.GameSession, error)
	DeleteGameSession(ctx context.Context, id uint64) error

	GetFirstAction(ctx context.Context, sesID uint64) (*models.GameAction, error)
	AddFirstGameAction(ctx context.Context, sesID uint64, action *models.GameAction) error
	DeleteFirstGameAction(ctx context.Context, sesID uint64) error

	GetGameSessionUpdates(ctx context.Context, id uint64) ([]*models.GameSessionUpdate, error)
	AddGameSessionUpdate(ctx context.Context, upd *models.GameSessionUpdate) error
	DeleteGameSessionUpdates(ctx context.Context, sesId uint64) error
	GetCasinoSessions(ctx context.Context, filter FilterType, casinoId eos.Uint64) ([]*models.GameSession, error)

	AddGameSessionTransaction(ctx context.Context, trxID string, sesID uint64,
		actionType uint16, actionParams []uint64) error
}
