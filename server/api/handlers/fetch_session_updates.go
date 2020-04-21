package handlers

import (
	"context"
	"encoding/json"
	"platform-backend/server/api/interfaces"
)

type FetchSessionUpdatesPayload struct {
	SessionId uint64 `json:"sessionId"`
}

func ProcessFetchSessionUpdatesRequest(context context.Context, req *interfaces.ApiRequest) (*interfaces.WsResponse, error) {
	var payload FetchSessionUpdatesPayload
	if err := json.Unmarshal(req.Data.Payload, &payload); err != nil {
		return nil, err
	}

	gameSessionUpdates, err := req.Repos.GameSession.GetGameSessionUpdates(context, payload.SessionId)

	if err != nil {
		return &interfaces.WsResponse{
			Type:   "response",
			Id:     req.Data.Id,
			Status: "error",
			Payload: interfaces.WsError{
				Code:    5000,
				Message: "Session fetch error: " + err.Error(),
			},
		}, nil
	}

	return &interfaces.WsResponse{
		Type:    "response",
		Id:      req.Data.Id,
		Status:  "ok",
		Payload: gameSessionUpdates,
	}, nil
}
