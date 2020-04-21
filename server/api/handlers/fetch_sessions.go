package handlers

import (
	"context"
	"platform-backend/server/api/interfaces"
)

func ProcessFetchSessionsRequest(context context.Context, req *interfaces.ApiRequest) (*interfaces.WsResponse, error) {
	gameSessions, err := req.Repos.GameSession.GetAllGameSessions(context)

	if err != nil {
		return &interfaces.WsResponse{
			Type:   "response",
			Id:     req.Data.Id,
			Status: "error",
			Payload: interfaces.WsError{
				Code:    5000,
				Message: "Sessions fetch error: " + err.Error(),
			},
		}, nil
	}

	return &interfaces.WsResponse{
		Type:    "response",
		Id:      req.Data.Id,
		Status:  "ok",
		Payload: gameSessions,
	}, nil
}
