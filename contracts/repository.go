package contracts

import (
	"context"
	"github.com/eoscanada/eos-go"
	"platform-backend/models"
)

type Repository interface {
	GetCasino(ctx context.Context, casinoId uint64) (*models.Casino, error)
	AllCasinos(ctx context.Context) ([]*models.Casino, error)
	GetCasinoGames(ctx context.Context, casinoName string) ([]*models.CasinoGame, error)

	GetGame(ctx context.Context, gameId uint64) (*models.Game, error)
	AllGames(ctx context.Context) ([]*models.Game, error)

	GetPlayerInfo(ctx context.Context, accountName string) (*models.PlayerInfo, error)
	GetRawAccount(accountName string) (*eos.AccountResp, error)

	GetBonusBalances(casinos []*models.Casino, accountName string) ([]*models.BonusBalance, error)
}
