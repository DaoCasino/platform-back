package blockchain

import (
	"context"
	"github.com/eoscanada/eos-go"
	"platform-backend/blockchain"
	"platform-backend/casino"
	"platform-backend/models"
	"strconv"
)

type Casino struct {
	Id        uint64 `json:"id"`
	Contract  string `json:"contract"`
	Paused    int    `json:"paused"`
	RsaPubkey string `json:"rsa_pubkey"`
	Meta      []byte `json:"bytes"`
}

type GameParam struct {
	Type  uint16 `json:"first"`
	Value uint32 `json:"second"`
}

type CasinoGame struct {
	Id     uint64      `json:"game_id"`
	Paused int         `json:"paused"`
	Params []GameParam `json:"params"`
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
		LowerBound: strconv.FormatUint(casinoId, 10),
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
		return nil, casino.CasinoNotFound
	}

	return toModelCasino(casinos[0]), nil
}

func (r *CasinoBlockchainRepo) GetCasinoGames(ctx context.Context, casinoName string) ([]*models.CasinoGame, error) {
	resp, err := r.bc.Api.GetTableRows(eos.GetTableRowsRequest{
		Code:  casinoName,
		Scope: casinoName,
		Table: "game",
		Limit: 100,
		JSON:  true,
	})

	if err != nil {
		return nil, err
	}

	casinosGames := make([]*CasinoGame, 1)
	err = resp.JSONToStructs(&casinosGames)
	if err != nil {
		return nil, err
	}

	ret := make([]*models.CasinoGame, 0)
	for _, game := range casinosGames {
		ret = append(ret, toModelCasinoGame(game))
	}

	return ret, nil
}

func toModelCasino(c *Casino) *models.Casino {
	return &models.Casino{
		Id:       c.Id,
		Contract: c.Contract,
		Paused:   !(c.Paused == 0),
	}
}

func toModelCasinoGame(game *CasinoGame) *models.CasinoGame {
	params := make([]models.GameParam, 0)

	for _, param := range game.Params {
		params = append(params, models.GameParam{
			Type:  param.Type,
			Value: param.Value,
		})
	}

	return &models.CasinoGame{
		Id:     game.Id,
		Paused: !(game.Paused == 0),
		Params: params,
	}
}
