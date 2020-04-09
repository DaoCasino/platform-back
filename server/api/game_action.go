package api

import (
	"context"
	"encoding/json"
	"platform-backend/usecases"
)

type GameActionPayload struct {
	*WsRequest
	Payload struct {
		SessionId  string   `json:"sessionid"`
		ActionType int32    `json:"actiontype"`
		Params     []string `json:"params"`
	} `json:"payload"`
}

func ProcessGameActionRequest(context context.Context, useCases *usecases.UseCases, session *Session, message []byte) (*WsResponse, error) {
	var messageObj GameActionPayload
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
