package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"platform-backend/server/api/interfaces"
)

type FetchSessionPayload struct {
	SessionId uint64 `json:"sessionId"`
}

func ProcessFetchSessionRequest(context context.Context, req *interfaces.ApiRequest) (interface{}, *interfaces.HandlerError) {
	var payload FetchSessionPayload
	if err := json.Unmarshal(req.Data.Payload, &payload); err != nil {
		return nil, interfaces.NewHandlerError(interfaces.RequestParseError, err)
	}

	gameSession, err := req.Repos.GameSession.GetGameSession(context, payload.SessionId)
	if err != nil {
		return nil, interfaces.NewHandlerError(interfaces.InternalError, err)
	}

	if gameSession.Player != req.User.AccountName {
		return nil, interfaces.NewHandlerError(interfaces.UnauthorizedError, errors.New("attempt to fetch not own session"))
	}

	return gameSession, nil
}
