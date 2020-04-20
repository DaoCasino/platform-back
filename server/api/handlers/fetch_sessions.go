package handlers

import (
	"context"
	"encoding/json"
	"platform-backend/server/api/interfaces"
)

type FetchSessionsPayload struct {
	Deposit  string `json:"sessionId"`
}

func ProcessFetchSessionsRequest(context context.Context, req *interfaces.ApiRequest) (*interfaces.WsResponse, error) {
	var payload FetchSessionsPayload
	if err := json.Unmarshal(req.Data.Payload, &payload); err != nil {
		return nil, err
	}

	return &interfaces.WsResponse{
		Type:    "response",
		Id:      req.Data.Id,
		Status:  "ok",
		Payload: struct{}{},
	}, nil
}
