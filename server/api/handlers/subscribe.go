package handlers

import (
	"context"
	"encoding/json"
	"platform-backend/server/api/interfaces"
)

type SubscribePayload struct {
}

func ProcessSubscribeRequest(context context.Context, req *interfaces.ApiRequest) (*interfaces.WsResponse, error) {
	var messageObj SubscribePayload
	if err := json.Unmarshal(req.Data.Payload, &messageObj); err != nil {
		return nil, err
	}

	// TODO updates service

	return &interfaces.WsResponse{
		Type:    "response",
		Id:      req.Data.Id,
		Status:  "ok",
		Payload: struct{}{},
	}, nil
}
