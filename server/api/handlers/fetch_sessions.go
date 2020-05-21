package handlers

import (
	"context"
	"platform-backend/server/api/ws_interface"
)

func ProcessFetchSessionsRequest(context context.Context, req *ws_interface.ApiRequest) (interface{}, *ws_interface.HandlerError) {
	gameSessions, err := req.Repos.GameSession.GetUserGameSessions(context, req.User.AccountName)
	if err != nil {
		return nil, ws_interface.NewHandlerError(ws_interface.InternalError, err)
	}

	response := make([]*GameSessionResponse, len(gameSessions))
	for i, session := range gameSessions {
		response[i] = toGameSessionResponse(session)
	}

	return response, nil
}
