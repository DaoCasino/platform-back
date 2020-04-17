package handlers

import (
	"context"
	"encoding/json"
	"platform-backend/server/api/interfaces"
)

type NewGamePayload struct {
	Deposit string `json:"deposit"`
	GameId  int32  `json:"gameid"`
}

func ProcessNewGameRequest(context context.Context, req *interfaces.ApiRequest) (*interfaces.WsResponse, error) {
	var payload NewGamePayload
	if err := json.Unmarshal(req.Data.Payload, &payload); err != nil {
		return nil, err
	}

	_ = req.UseCases.GameSession.NewSession(context, 0)

	return &interfaces.WsResponse{
		Type:    "response",
		Id:      req.Data.Id,
		Status:  "ok",
		Payload: struct{}{},
	}, nil
}
