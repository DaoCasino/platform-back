package handlers

import (
	"context"
	"encoding/json"
	"platform-backend/server/api/interfaces"
)

type FetchCasinosPayload struct {
	Deposit  string `json:"deposit"`
	CasinoId int32  `json:"casinoid"`
	GameId   int32  `json:"gameid"`
}

func ProcessFetchCasinosRequest(context context.Context, req *interfaces.ApiRequest) (*interfaces.WsResponse, error) {
	var payload FetchCasinosPayload
	if err := json.Unmarshal(req.Data.Payload, &payload); err != nil {
		return nil, err
	}

	casinos, err := req.UseCases.Casino.AllCasinos(context)
	if err != nil {
		return nil, err
	}

	return &interfaces.WsResponse{
		Type:    "response",
		Id:      req.Data.Id,
		Status:  "ok",
		Payload: casinos,
	}, nil
}
