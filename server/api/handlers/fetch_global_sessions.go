package handlers

import (
	"context"
	"encoding/json"
	gamesessions "platform-backend/game_sessions"
	"platform-backend/server/api/ws_interface"
)

type FetchGlobalSessionsPayload struct {
	Filter gamesessions.FilterType `json:"filter"`
}

func ProcessFetchGlobalSessionsRequest(context context.Context, req *ws_interface.ApiRequest) (interface{}, *ws_interface.HandlerError) {
	var payload FetchGlobalSessionsPayload
	if err := json.Unmarshal(req.Data.Payload, &payload); err != nil {
		return nil, ws_interface.NewHandlerError(ws_interface.RequestParseError, err)
	}

	gameSessions, err := req.Repos.GameSession.GetGlobalSessions(context, payload.Filter)

	if err != nil {
		return nil, ws_interface.NewHandlerError(ws_interface.InternalError, err)
	}

	response := make([]*GameSessionResponse, len(gameSessions))
	for i, session := range gameSessions {
		response[i] = toGameSessionResponse(session)
	}

	return response, nil
}
