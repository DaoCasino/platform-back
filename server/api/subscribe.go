package api

import (
	"context"
	"encoding/json"
	"platform-backend/usecases"
)

type SubscribePayload struct {
	*WsRequest
}

func ProcessSubscribeRequest(context context.Context, useCases *usecases.UseCases, session *Session, message []byte) (*WsResponse, error) {
	var messageObj SubscribePayload
	if err := json.Unmarshal(message, &messageObj); err != nil {
		return nil, err
	}
	session.subscribed = true
	return &WsResponse{
		Type:    "response",
		Id:      messageObj.Id,
		Status:  "ok",
		Payload: struct{}{},
	}, nil
}
