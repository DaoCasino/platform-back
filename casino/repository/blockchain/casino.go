package blockchain

import (
	"context"
	"errors"
	"github.com/eoscanada/eos-go"
	"platform-backend/blockchain"
	"platform-backend/models"
)

type Casino struct {
	Id        uint64 `json:"id"`
	Contract  string `json:"contract"`
	Paused    int    `json:"paused"`
	RsaPubkey string `json:"rsa_pubkey"`
	Meta      []byte `json:"bytes"`
}

type CasinoBlockchainRepo struct {
	bc               *blockchain.Blockchain
	platformContract string
}

func NewCasinoBlockchainRepo(blockchain *blockchain.Blockchain, platformContract string) *CasinoBlockchainRepo {
	return &CasinoBlockchainRepo{
		bc:               blockchain,
		platformContract: platformContract,
	}
}

func (r *CasinoBlockchainRepo) AllCasinos(ctx context.Context) ([]*models.Casino, error) {
	resp, err := r.bc.Api.GetTableRows(eos.GetTableRowsRequest{
		Code:  r.platformContract,
		Scope: r.platformContract,
		Table: "casino",
		Limit: 100,
		JSON:  true,
	})

	if err != nil {
		return nil, err
	}

	casinos := make([]*Casino, 100)
	err = resp.JSONToStructs(&casinos)
	if err != nil {
		return nil, err
	}

	ret := make([]*models.Casino, 0)
	for _, casino := range casinos {
		ret = append(ret, toModelCasino(casino))
	}

	return ret, nil
}

func (r *CasinoBlockchainRepo) GetCasino(ctx context.Context, casinoId uint64) (*models.Casino, error) {
	resp, err := r.bc.Api.GetTableRows(eos.GetTableRowsRequest{
		Code:       r.platformContract,
		Scope:      r.platformContract,
		Table:      "casino",
		Limit:      1,
		LowerBound: string(casinoId),
		JSON:       true,
	})

	if err != nil {
		return nil, err
	}

	casinos := make([]*Casino, 1)
	err = resp.JSONToStructs(&casinos)
	if err != nil {
		return nil, err
	}

	if len(casinos) == 0 || casinos[0].Id != casinoId {
		return nil, errors.New("casino not found")
	}

	return toModelCasino(casinos[0]), nil
}

func toModelCasino(c *Casino) *models.Casino {
	return &models.Casino{
		Id:       c.Id,
		Contract: c.Contract,
		Paused:   !(c.Paused == 0),
	}
}
