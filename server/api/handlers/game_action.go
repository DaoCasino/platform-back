package handlers

import (
	"context"
	"encoding/json"
	"platform-backend/server/api/interfaces"
)

type GameActionPayload struct {
	SessionId  string   `json:"sessionid"`
	ActionType int32    `json:"actiontype"`
	Params     []string `json:"params"`
}

func ProcessGameActionRequest(context context.Context, req *interfaces.ApiRequest) (*interfaces.WsResponse, error) {
	var payload GameActionPayload
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
