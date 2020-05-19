package handlers

import (
	"context"
	"encoding/json"
	"errors"
	gamesessions "platform-backend/game_sessions"
	"platform-backend/server/api/ws_interface"
)

type FetchSessionPayload struct {
	SessionId uint64 `json:"sessionId"`
}

func ProcessFetchSessionRequest(context context.Context, req *ws_interface.ApiRequest) (interface{}, *ws_interface.HandlerError) {
	var payload FetchSessionPayload
	if err := json.Unmarshal(req.Data.Payload, &payload); err != nil {
		return nil, ws_interface.NewHandlerError(ws_interface.RequestParseError, err)
	}

	gameSession, err := req.Repos.GameSession.GetGameSession(context, payload.SessionId)
	if err == gamesessions.ErrGameSessionNotFound {
		return nil, ws_interface.NewHandlerError(ws_interface.SessionNotFoundError, err)
	}
	if err != nil {
		return nil, ws_interface.NewHandlerError(ws_interface.InternalError, err)
	}

	if gameSession.Player != req.User.AccountName {
		return nil, ws_interface.NewHandlerError(ws_interface.UnauthorizedError, errors.New("attempt to fetch not own session"))
	}

	return gameSession, nil
}
