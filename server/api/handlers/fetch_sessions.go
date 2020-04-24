package handlers

import (
	"context"
	"platform-backend/server/api/interfaces"
)

func ProcessFetchSessionsRequest(context context.Context, req *interfaces.ApiRequest) (interface{}, *interfaces.HandlerError) {
	gameSessions, err := req.Repos.GameSession.GetAllGameSessions(context, req.User.AccountName)
	if err != nil {
		return nil, interfaces.NewHandlerError(interfaces.InternalError, err)
	}

	return gameSessions, nil
}
