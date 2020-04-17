package usecases

import (
	"context"
	"github.com/eoscanada/eos-go"
	"platform-backend/blockchain"
	casino "platform-backend/casino"
	"platform-backend/models"
)

type PlayerUseCase struct {
	blockchain *blockchain.Blockchain
	casinoRepo casino.Repository
}

func NewPlayerUseCase(blockchain *blockchain.Blockchain, casinoRepo casino.Repository) *PlayerUseCase {
	return &PlayerUseCase{
		blockchain: blockchain,
		casinoRepo: casinoRepo,
	}
}

func (p *PlayerUseCase) GetInfo(ctx context.Context, accountName string) (*models.PlayerInfo, error) {
	resp, err := p.blockchain.Api.GetAccount(eos.AN(accountName))
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

	casinos, err := p.casinoRepo.AllCasinos(ctx)

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