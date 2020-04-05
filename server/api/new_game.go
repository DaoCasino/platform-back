package api

import (
	"context"
	"encoding/json"
	"platform-backend/usecases"
)

type NewGamePayload struct {
	*WsRequest
	Payload struct {
		Deposit  string `json:"deposit"`
		CasinoId int32  `json:"casinoid"`
		GameId   int32  `json:"gameid"`
	} `json:"payload"`
}

func ProcessNewGameRequest(context context.Context, useCases *usecases.UseCases, session *Session, message []byte) (*WsResponse, error) {
	var messageObj NewGamePayload
	if err := json.Unmarshal(message, &messageObj); err != nil {
		return nil, err
	}

	return &WsResponse{
		Type:    "response",
		Id:      messageObj.Id,
		Status:  "ok",
		Payload: struct{}{},
	}, nil
}
