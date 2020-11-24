package blockchain

import (
	"context"
	"encoding/json"
	"platform-backend/blockchain"
	"platform-backend/contracts"
	"platform-backend/models"
	"strconv"

	"github.com/eoscanada/eos-go"
	"github.com/rs/zerolog/log"
)

type Casino struct {
	Id        eos.Uint64           `json:"id"`
	Contract  string               `json:"contract"`
	Paused    int                  `json:"paused"`
	RsaPubkey string               `json:"rsa_pubkey"`
	Meta      blockchain.ByteArray `json:"meta"`
}

type GameParam struct {
	Type  uint16     `json:"first"`
	Value eos.Uint64 `json:"second"`
}

type CasinoGame struct {
	Id     eos.Uint64  `json:"game_id"`
	Paused int         `json:"paused"`
	Params []GameParam `json:"params"`
}

type BonusBalance struct {
	Player  string    `json:"player"`
	Balance eos.Asset `json:"balance"`
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

	if len(casinos) == 0 || uint64(casinos[0].Id) != casinoId {
		return nil, contracts.CasinoNotFound
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

func (r *CasinoBlockchainRepo) GetCasinoGamesState(ctx context.Context, casinoName string) ([]*models.GameState, error) {
	resp, err := r.bc.Api.GetTableRows(eos.GetTableRowsRequest{
		Code:  casinoName,
		Scope: casinoName,
		Table: "gamestate",
		Limit: 100,
		JSON:  true,
	})

	if err != nil {
		return nil, err
	}

	var states []*models.GameState
	err = resp.JSONToStructs(&states)
	if err != nil {
		return nil, err
	}

	return states, nil
}

func (r *CasinoBlockchainRepo) GetBonusBalances(casinos []*models.Casino, accountName string) ([]*models.BonusBalance, error) {
	bonusBalances := make([]*models.BonusBalance, 0, 1)
	for _, casino := range casinos {
		primaryKey := strconv.FormatUint(eos.MustStringToName(accountName), 10)
		resp, err := r.bc.Api.GetTableRows(eos.GetTableRowsRequest{
			Code:       casino.Contract,
			Scope:      casino.Contract,
			Table:      "bonusbalance",
			LowerBound: primaryKey,
			UpperBound: primaryKey,
			Limit:      1,
			JSON:       true,
		})
		if err != nil {
			return nil, err
		}

		bonusBalance := make([]*BonusBalance, 0, 1)
		err = resp.JSONToStructs(&bonusBalance)
		if err != nil {
			return nil, err
		}

		if len(bonusBalance) == 0 {
			continue
		}
		bonusBalances = append(bonusBalances, toModelBonusBalance(bonusBalance[0], casino.Id))
	}
	return bonusBalances, nil
}

func toModelCasino(c *Casino) *models.Casino {
	meta := &models.CasinoMeta{}
	err := json.Unmarshal(c.Meta, meta)
	if err != nil {
		log.Warn().Msgf("invalid casino meta, setting null, ID: %d, err: %s", c.Id, err.Error())
		// set null meta if invalid json
		meta = nil
	}

	return &models.Casino{
		Id:       uint64(c.Id),
		Contract: c.Contract,
		Paused:   !(c.Paused == 0),
		Meta:     meta,
	}
}

func toModelCasinoGame(game *CasinoGame) *models.CasinoGame {
	params := make([]models.GameParam, 0)

	for _, param := range game.Params {
		params = append(params, models.GameParam{
			Type:  param.Type,
			Value: uint64(param.Value),
		})
	}

	return &models.CasinoGame{
		Id:     uint64(game.Id),
		Paused: !(game.Paused == 0),
		Params: params,
	}
}

func toModelBonusBalance(bonusBalance *BonusBalance, casinoId uint64) *models.BonusBalance {
	return &models.BonusBalance{
		Balance:  bonusBalance.Balance,
		CasinoId: casinoId,
	}
}
