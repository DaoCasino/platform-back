package handlers

import (
	"context"
	"platform-backend/server/api/ws_interface"
)

func ProcessAccountInfo(context context.Context, req *ws_interface.ApiRequest) (interface{}, *ws_interface.HandlerError) {
	player, err := req.Repos.Casino.GetPlayerInfo(context, req.User.AccountName)

	if err != nil {
		return nil, ws_interface.NewHandlerError(ws_interface.InternalError, err)
	}

	return player, nil
}
