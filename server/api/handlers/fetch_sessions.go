package handlers

import (
	"context"
	"encoding/json"
	"platform-backend/server/api/interfaces"
)

type FetchSessionsPayload struct {
	*interfaces.WsRequest
	Payload struct {
		Deposit  string `json:"deposit"`
		CasinoId int32  `json:"casinoid"`
		GameId   int32  `json:"gameid"`
	} `json:"payload"`
}

func ProcessFetchSessionsRequest(context context.Context, req *interfaces.ApiRequest) (*interfaces.WsResponse, error) {
	var messageObj FetchSessionsPayload
	if err := json.Unmarshal(req.Message, &messageObj); err != nil {
		return nil, err
	}

	return &interfaces.WsResponse{
		Type:    "response",
		Id:      messageObj.Id,
		Status:  "ok",
		Payload: struct{}{},
	}, nil
}
