package handlers

import (
	"context"
	"platform-backend/server/api/interfaces"
)

func ProcessFetchGamesRequest(context context.Context, req *interfaces.ApiRequest) (*interfaces.WsResponse, error) {
	games, err := req.Repos.Casino.AllGames(context)

	if err != nil {
		return &interfaces.WsResponse{
			Type:   "response",
			Id:     req.Data.Id,
			Status: "error",
			Payload: interfaces.WsError{
				Code:    5000,
				Message: "Games fetch error: " + err.Error(),
			},
		}, nil
	}

	return &interfaces.WsResponse{
		Type:    "response",
		Id:      req.Data.Id,
		Status:  "ok",
		Payload: games,
	}, nil
}
