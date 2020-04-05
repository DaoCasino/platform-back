package api

import (
	"context"
	"encoding/json"
	"platform-backend/usecases"
)

type FetchGamesInCasinoPayload struct {
	*WsRequest
	Payload struct {
		Deposit  string `json:"deposit"`
		CasinoId int32  `json:"casinoid"`
		GameId   int32  `json:"gameid"`
	} `json:"payload"`
}

func ProcessFetchGamesInCasinoRequest(context context.Context, useCases *usecases.UseCases, session *Session, message []byte) (*WsResponse, error) {
	var messageObj FetchGamesInCasinoPayload
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
