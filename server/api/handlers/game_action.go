package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/eoscanada/eos-go"
	gamesessions "platform-backend/game_sessions"
	"platform-backend/models"
	"platform-backend/server/api/ws_interface"
)

type GameActionPayload struct {
	SessionId  eos.Uint64   `json:"sessionId"`
	ActionType uint16       `json:"actionType"`
	Params     []eos.Uint64 `json:"params"`
}

func ProcessGameActionRequest(context context.Context, req *ws_interface.ApiRequest) (interface{}, *ws_interface.HandlerError) {
	var payload GameActionPayload
	if err := json.Unmarshal(req.Data.Payload, &payload); err != nil {
		return nil, ws_interface.NewHandlerError(ws_interface.RequestParseError, err)
	}

	session, err := req.Repos.GameSession.GetGameSession(context, uint64(payload.SessionId))
	if err == gamesessions.ErrGameSessionNotFound {
		return nil, ws_interface.NewHandlerError(ws_interface.SessionNotFoundError, err)
	}
	if err != nil {
		return nil, ws_interface.NewHandlerError(ws_interface.InternalError, err)
	}

	if session.State != models.RequestedGameAction {
		return nil, ws_interface.NewHandlerError(ws_interface.SessionInvalidStateError, errors.New("attempt to action while invalid state"))
	}

	if req.User.AccountName != session.Player {
		return nil, ws_interface.NewHandlerError(ws_interface.UnauthorizedError, errors.New("attempt to play not own session"))
	}

	var params []uint64
	for _, param := range payload.Params {
		params = append(params, uint64(param))
	}

	err = req.UseCases.GameSession.GameAction(context, uint64(payload.SessionId), payload.ActionType, params)
	if err != nil {
		return nil, ws_interface.NewHandlerError(ws_interface.InternalError, err)
	}

	return struct{}{}, nil
}
