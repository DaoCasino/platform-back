package contracts

import (
	"github.com/eoscanada/eos-go"
	"platform-backend/models"
)

func CasinoLinked(permissions *[]eos.Permission, casinoName string) bool {
	for _, permission := range *permissions {
		if permission.PermName == casinoName {
			return true
		}
	}
	return false
}

func FillPlayerInfoFromRaw(
	destInfo *models.PlayerInfo,
	sourceRaw *eos.AccountResp,
	casinos []*models.Casino,
	bonusBalances []*models.BonusBalance,
) {
	for _, perm := range sourceRaw.Permissions {
		if perm.PermName == "owner" {
			destInfo.OwnerPermission = perm.RequiredAuth
			continue
		}
		if perm.PermName == "active" {
			destInfo.ActivePermission = perm.RequiredAuth
			continue
		}
	}

	destInfo.Balance = sourceRaw.CoreLiquidBalance

	destInfo.BonusBalances = bonusBalances
	destInfo.LinkedCasinos = make([]*models.Casino, 0)
	for _, casino := range casinos {
		if CasinoLinked(&sourceRaw.Permissions, casino.Contract) {
			destInfo.LinkedCasinos = append(destInfo.LinkedCasinos, casino)
		}
	}
}
