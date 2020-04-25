package blockchain

import (
	"context"
	"github.com/eoscanada/eos-go"
	"platform-backend/models"
)

func (r *CasinoBlockchainRepo) GetPlayerInfo(ctx context.Context, accountName string) (*models.PlayerInfo, error) {
	resp, err := r.bc.Api.GetAccount(eos.AN(accountName))
	if err != nil {
		return nil, err
	}

	var info models.PlayerInfo

	for _, perm := range resp.Permissions {
		if perm.PermName == "owner" {
			info.OwnerPermission = perm.RequiredAuth
			continue
		}
		if perm.PermName == "active" {
			info.ActivePermission = perm.RequiredAuth
			continue
		}
	}

	info.Balance = resp.CoreLiquidBalance
	info.LinkedCasinos = make([]*models.Casino, 0)

	casinos, err := r.AllCasinos(ctx)
	if err != nil {
		return nil, err
	}

	for _, cas := range casinos {
		if casinoLinked(&resp.Permissions, cas.Contract) {
			info.LinkedCasinos = append(info.LinkedCasinos, cas)
		}
	}

	return &info, nil
}

func casinoLinked(permissions *[]eos.Permission, casinoName string) bool {
	for _, permission := range *permissions {
		if permission.PermName == casinoName {
			return true
		}
	}
	return false
}
