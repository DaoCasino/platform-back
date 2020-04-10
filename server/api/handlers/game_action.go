package handlers

import (
	"context"
	"encoding/json"
	"platform-backend/server/api/interfaces"
)

type GameActionPayload struct {
	*interfaces.WsRequest
	Payload struct {
		SessionId  string   `json:"sessionid"`
		ActionType int32    `json:"actiontype"`
		Params     []string `json:"params"`
	} `json:"payload"`
}

func ProcessGameActionRequest(context context.Context, req *interfaces.ApiRequest) (*interfaces.WsResponse, error) {
	var messageObj GameActionPayload
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
