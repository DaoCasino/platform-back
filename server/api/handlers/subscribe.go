package handlers

import (
	"context"
	"encoding/json"
	"platform-backend/server/api/interfaces"
)

type SubscribePayload struct {
	*interfaces.WsRequest
}

func ProcessSubscribeRequest(context context.Context, req *interfaces.ApiRequest) (*interfaces.WsResponse, error) {
	var messageObj SubscribePayload
	if err := json.Unmarshal(req.Message, &messageObj); err != nil {
		return nil, err
	}

	// TODO updates service

	return &interfaces.WsResponse{
		Type:    "response",
		Id:      messageObj.Id,
		Status:  "ok",
		Payload: struct{}{},
	}, nil
}
