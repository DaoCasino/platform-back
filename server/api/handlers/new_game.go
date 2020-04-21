package handlers

import (
	"context"
	"encoding/json"
	"platform-backend/server/api/interfaces"
)

type NewGamePayload struct {
	GameId   uint64 `json:"gameid"`
	CasinoID uint64 `json:"casinoid"`
	Deposit  string `json:"deposit"`
}

func ProcessNewGameRequest(context context.Context, req *interfaces.ApiRequest) (*interfaces.WsResponse, error) {
	var payload NewGamePayload
	if err := json.Unmarshal(req.Data.Payload, &payload); err != nil {
		return nil, err
	}

	session, err := req.UseCases.GameSession.NewSession(context, payload.GameId, payload.CasinoID, payload.Deposit, req.User)
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
