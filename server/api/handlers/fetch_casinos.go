package handlers

import (
	"context"
	"encoding/json"
	"platform-backend/server/api/interfaces"
)

type FetchCasinosPayload struct {
	*interfaces.WsRequest
	Payload struct {
		Deposit  string `json:"deposit"`
		CasinoId int32  `json:"casinoid"`
		GameId   int32  `json:"gameid"`
	} `json:"payload"`
}

func ProcessFetchCasinosRequest(context context.Context, req *interfaces.ApiRequest) (*interfaces.WsResponse, error) {
	var messageObj FetchCasinosPayload
	if err := json.Unmarshal(req.Message, &messageObj); err != nil {
		return nil, err
	}

	casinos, err := req.UseCases.Casino.AllCasinos(context)
	if err != nil {
		return nil, err
	}

	return &interfaces.WsResponse{
		Type:    "response",
		Id:      messageObj.Id,
		Status:  "ok",
		Payload: casinos,
	}, nil
}
