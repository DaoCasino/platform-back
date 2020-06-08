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
	Deposit    string       `json:"deposit"`
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

	params := make([]uint64, len(payload.Params))
	for i, param := range payload.Params {
		params[i] = uint64(param)
	}

	if payload.Deposit != "" {
		err = req.UseCases.GameSession.GameActionWithDeposit(context, uint64(payload.SessionId), payload.ActionType, params, payload.Deposit)
	} else {
		err = req.UseCases.GameSession.GameAction(context, uint64(payload.SessionId), payload.ActionType, params)
	}
	if err != nil {
		return nil, ws_interface.NewHandlerError(ws_interface.InternalError, err)
	}

	return struct{}{}, nil
}
