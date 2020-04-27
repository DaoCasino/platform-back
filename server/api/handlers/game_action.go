package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"platform-backend/models"
	"platform-backend/server/api/interfaces"
)

type GameActionPayload struct {
	SessionId  uint64   `json:"sessionId"`
	ActionType uint16   `json:"actionType"`
	Params     []uint64 `json:"params"`
}

func ProcessGameActionRequest(context context.Context, req *interfaces.ApiRequest) (interface{}, *interfaces.HandlerError) {
	var payload GameActionPayload
	if err := json.Unmarshal(req.Data.Payload, &payload); err != nil {
		return nil, interfaces.NewHandlerError(interfaces.RequestParseError, err)
	}

	has, err := req.Repos.GameSession.HasGameSession(context, payload.SessionId)
	if err != nil {
		return nil, interfaces.NewHandlerError(interfaces.InternalError, err)
	}

	if !has {
		return nil, interfaces.NewHandlerError(interfaces.ContentNotFoundError, errors.New("game session not found"))
	}

	session, err := req.Repos.GameSession.GetGameSession(context, payload.SessionId)
	if err != nil {
		return nil, interfaces.NewHandlerError(interfaces.InternalError, err)
	}

	if session.State != models.RequestedGameAction {
		return nil, interfaces.NewHandlerError(interfaces.BadRequest, errors.New("attempt to action while invalid state"))
	}

	if req.User.AccountName != session.Player {
		return nil, interfaces.NewHandlerError(interfaces.UnauthorizedError, errors.New("attempt to play not own session"))
	}

	err = req.UseCases.GameSession.GameAction(context, payload.SessionId, payload.ActionType, payload.Params)
	if err != nil {
		return nil, interfaces.NewHandlerError(interfaces.InternalError, err)
	}

	return struct{}{}, nil
}
