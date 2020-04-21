package handlers

import (
	"context"
	"encoding/json"
	"platform-backend/server/api/interfaces"
)

type NewGamePayload struct {
	GameId   uint64 `json:"gameId"`
	CasinoID uint64 `json:"casinoId"`
	Deposit  string `json:"deposit"`
}

func ProcessNewGameRequest(context context.Context, req *interfaces.ApiRequest) (*interfaces.WsResponse, error) {
	var payload NewGamePayload
	if err := json.Unmarshal(req.Data.Payload, &payload); err != nil {
		return nil, err
	}

	game, err := req.Repos.Casino.GetGame(context, payload.GameId)
	if err != nil {
		return nil, err
	}

	casino, err := req.Repos.Casino.GetCasino(context, payload.CasinoID)
	if err != nil {
		return nil, err
	}

	session, err := req.UseCases.GameSession.NewSession(context, casino, game, req.User, payload.Deposit)
	if err != nil {
		return nil, err
	}

	return &interfaces.WsResponse{
		Type:    "response",
		Id:      req.Data.Id,
		Status:  "ok",
		Payload: session,
	}, nil
}
