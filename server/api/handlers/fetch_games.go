package handlers

import (
	"context"
	"platform-backend/server/api/ws_interface"
)

func ProcessFetchGamesRequest(context context.Context, req *ws_interface.ApiRequest) (interface{}, *ws_interface.HandlerError) {
	games, err := req.Repos.Contracts.AllGames(context)
	if err != nil {
		return nil, ws_interface.NewHandlerError(ws_interface.InternalError, err)
	}

	return games, nil
}
