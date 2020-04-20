package blockchain

import (
	"context"
	"github.com/eoscanada/eos-go"
	"platform-backend/casino"
	"platform-backend/models"
	"strconv"
)

type Game struct {
	Id           uint64 `json:"id"`
	Contract     string `json:"contract"`
	ParamsCnt    uint16 `json:"params_cnt"`
	Paused       int    `json:"paused"`
	ProfitMargin uint32 `json:"profit_margin"`
	Beneficiary  string `json:"beneficiary"`
	Meta         []byte `json:"bytes"`
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

	if len(games) == 0 || games[0].Id != gameId {
		return nil, casino.GameNotFound
	}

	return toModelGame(games[0]), nil
}

func (r *CasinoBlockchainRepo) AllGames(ctx context.Context) ([]*models.Game, error) {
	resp, err := r.bc.Api.GetTableRows(eos.GetTableRowsRequest{
		Code:       r.platformContract,
		Scope:      r.platformContract,
		Table:      "game",
		Limit:      100,
		JSON:       true,
	})

	if err != nil {
		return nil, err
	}

	games := make([]*Game, 100)
	err = resp.JSONToStructs(&games)
	if err != nil {
		return nil, err
	}

	ret := make([]*models.Game, 0)
	for _, game := range games {
		ret = append(ret, toModelGame(game))
	}

	return ret, nil
}


func toModelGame(g *Game) *models.Game {
	return &models.Game{
		Id: g.Id,
		Contract: g.Contract,
		ParamsCnt: g.ParamsCnt,
		Paused: g.Paused,
	}
}
