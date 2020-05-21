package handlers

import (
	"context"
	"encoding/json"
	"platform-backend/models"
	"platform-backend/server/api/ws_interface"
	"strconv"
)

type GameResponse struct {
	Id        string          `json:"id"`
	Contract  string          `json:"contract"`
	ParamsCnt uint16          `json:"paramsCnt"`
	Paused    int             `json:"paused"`
	Meta      json.RawMessage `json:"meta"`
}

func toGameResponse(g *models.Game) *GameResponse {
	return &GameResponse{
		Id:        strconv.FormatUint(g.Id, 10),
		Contract:  g.Contract,
		ParamsCnt: g.ParamsCnt,
		Paused:    g.Paused,
		Meta:      g.Meta,
	}
}

func ProcessFetchGamesRequest(context context.Context, req *ws_interface.ApiRequest) (interface{}, *ws_interface.HandlerError) {
	games, err := req.Repos.Contracts.AllGames(context)
	if err != nil {
		return nil, ws_interface.NewHandlerError(ws_interface.InternalError, err)
	}

	var response []*GameResponse
	for _, game := range games {
		response = append(response, toGameResponse(game))
	}

	return response, nil
}
