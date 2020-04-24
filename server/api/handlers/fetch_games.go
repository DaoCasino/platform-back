package handlers

import (
	"context"
	"platform-backend/server/api/interfaces"
)

func ProcessFetchGamesRequest(context context.Context, req *interfaces.ApiRequest) (interface{}, *interfaces.HandlerError) {
	games, err := req.Repos.Casino.AllGames(context)
	if err != nil {
		return nil, interfaces.NewHandlerError(interfaces.InternalError, err)
	}

	return games, nil
}
