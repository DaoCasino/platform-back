package handlers

import (
	"context"
	"platform-backend/server/api/interfaces"
)

func ProcessAccountInfo(context context.Context, req *interfaces.ApiRequest) (interface{}, *interfaces.HandlerError) {
	player, err := req.Repos.Casino.GetPlayerInfo(context, req.User.AccountName)

	if err != nil {
		return nil, interfaces.NewHandlerError(interfaces.InternalError, err)
	}

	return player, nil
}
