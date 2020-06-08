package blockchain

import (
	"context"
	"encoding/json"
	"github.com/eoscanada/eos-go"
	"github.com/rs/zerolog/log"
	"platform-backend/blockchain"
	"platform-backend/contracts"
	"platform-backend/models"
	"strconv"
)

type Game struct {
	Id           eos.Uint64           `json:"id"`
	Contract     string               `json:"contract"`
	ParamsCnt    uint16               `json:"params_cnt"`
	Paused       int                  `json:"paused"`
	ProfitMargin uint32               `json:"profit_margin"`
	Beneficiary  string               `json:"beneficiary"`
	Meta         blockchain.ByteArray `json:"meta"`
}

func (r *CasinoBlockchainRepo) GetGame(ctx context.Context, gameId uint64) (*models.Game, error) {
	resp, err := r.bc.Api.GetTableRows(eos.GetTableRowsRequest{
		Code:       r.platformContract,
		Scope:      r.platformContract,
		Table:      "game",
		Limit:      1,
		LowerBound: strconv.FormatUint(gameId, 10),
		JSON:       true,
	})

	if err != nil {
		return nil, err
	}

	games := make([]*Game, 1)
	err = resp.JSONToStructs(&games)
	if err != nil {
		return nil, err
	}

	if len(games) == 0 || uint64(games[0].Id) != gameId {
		return nil, contracts.GameNotFound
	}

	return toModelGame(games[0]), nil
}

func (r *CasinoBlockchainRepo) AllGames(ctx context.Context) ([]*models.Game, error) {
	resp, err := r.bc.Api.GetTableRows(eos.GetTableRowsRequest{
		Code:  r.platformContract,
		Scope: r.platformContract,
		Table: "game",
		Limit: 100,
		JSON:  true,
	})

	if err != nil {
		return nil, err
	}

	games := make([]*Game, 100)
	err = resp.JSONToStructs(&games)
	if err != nil {
		return nil, err
	}

	ret := make([]*models.Game, len(games))
	for i, game := range games {
		ret[i] = toModelGame(game)
	}

	return ret, nil
}

func toModelGame(g *Game) *models.Game {
	meta := &models.GameMeta{}
	err := json.Unmarshal(g.Meta, meta)
	if err != nil {
		log.Warn().Msgf("invalid game meta, setting null, ID: %d, err: %s", g.Id, err.Error())
		// set null meta if invalid json
		meta = nil
	}

	return &models.Game{
		Id:        uint64(g.Id),
		Contract:  g.Contract,
		ParamsCnt: g.ParamsCnt,
		Paused:    g.Paused,
		Meta:      meta,
	}
}
