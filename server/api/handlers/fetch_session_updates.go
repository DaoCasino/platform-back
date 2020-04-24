package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"platform-backend/server/api/interfaces"
)

type FetchSessionUpdatesPayload struct {
	SessionId uint64 `json:"sessionId"`
}

func ProcessFetchSessionUpdatesRequest(context context.Context, req *interfaces.ApiRequest) (interface{}, *interfaces.HandlerError) {
	var payload FetchSessionUpdatesPayload
	if err := json.Unmarshal(req.Data.Payload, &payload); err != nil {
		return nil, interfaces.NewHandlerError(interfaces.RequestParseError, err)
	}

	gameSession, err := req.Repos.GameSession.GetGameSession(context, payload.SessionId)
	if err != nil {
		return nil, interfaces.NewHandlerError(interfaces.InternalError, err)
	}

	if gameSession.Player != req.User.AccountName {
		return nil, interfaces.NewHandlerError(interfaces.UnauthorizedError, errors.New("attempt to fetch updates for not own session"))
	}

	gameSessionUpdates, err := req.Repos.GameSession.GetGameSessionUpdates(context, gameSession.ID)
	if err != nil {
		return nil, interfaces.NewHandlerError(interfaces.InternalError, err)
	}

	return gameSessionUpdates, nil
}
