package blockchain

import (
	"context"
	"github.com/eoscanada/eos-go"
	"platform-backend/contracts"
	"platform-backend/models"
)

func (r *CasinoBlockchainRepo) GetRawAccount(accountName string) (*eos.AccountResp, error) {
	resp, err := r.bc.Api.GetAccount(eos.AN(accountName))
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (r *CasinoBlockchainRepo) GetPlayerInfo(ctx context.Context, accountName string) (*models.PlayerInfo, error) {
	resp, err := r.GetRawAccount(accountName)
	if err != nil {
		return nil, err
	}

	casinos, err := r.AllCasinos(ctx)
	if err != nil {
		return nil, err
	}

	casinos = contracts.GetLinkedCasinos(resp, casinos)

	bonusBalances, err := r.GetBonusBalances(casinos, accountName)
	if err != nil {
		return nil, err
	}

	info := &models.PlayerInfo{}
	contracts.FillPlayerInfoFromRaw(info, resp, casinos, bonusBalances)

	return info, nil
}
