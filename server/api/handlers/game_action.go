package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"platform-backend/models"
	"platform-backend/server/api/ws_interface"
)

type GameActionPayload struct {
	SessionId  uint64   `json:"sessionId"`
	ActionType uint16   `json:"actionType"`
	Params     []uint64 `json:"params"`
}

func ProcessGameActionRequest(context context.Context, req *ws_interface.ApiRequest) (interface{}, *ws_interface.HandlerError) {
	var payload GameActionPayload
	if err := json.Unmarshal(req.Data.Payload, &payload); err != nil {
		return nil, ws_interface.NewHandlerError(ws_interface.RequestParseError, err)
	}

	has, err := req.Repos.GameSession.HasGameSession(context, payload.SessionId)
	if err != nil {
		return nil, ws_interface.NewHandlerError(ws_interface.InternalError, err)
	}

	if !has {
		return nil, ws_interface.NewHandlerError(ws_interface.ContentNotFoundError, errors.New("game session not found"))
	}

	session, err := req.Repos.GameSession.GetGameSession(context, payload.SessionId)
	if err != nil {
		return nil, ws_interface.NewHandlerError(ws_interface.InternalError, err)
	}

	if session.State != models.RequestedGameAction {
		return nil, ws_interface.NewHandlerError(ws_interface.BadRequest, errors.New("attempt to action while invalid state"))
	}

	if req.User.AccountName != session.Player {
		return nil, ws_interface.NewHandlerError(ws_interface.UnauthorizedError, errors.New("attempt to play not own session"))
	}

	err = req.UseCases.GameSession.GameAction(context, payload.SessionId, payload.ActionType, payload.Params)
	if err != nil {
		return nil, ws_interface.NewHandlerError(ws_interface.InternalError, err)
	}

	return struct{}{}, nil
}
