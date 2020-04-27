package handlers

import (
	"context"
	"encoding/json"
	"platform-backend/casino"
	"platform-backend/models"
	"platform-backend/server/api/ws_interface"
	"time"
)

type NewGamePayload struct {
	GameId       uint64   `json:"gameId"`
	CasinoID     uint64   `json:"casinoId"`
	Deposit      string   `json:"deposit"`
	ActionType   uint16   `json:"actionType"`
	ActionParams []uint64 `json:"actionParams"`
}

func ProcessNewGameRequest(context context.Context, req *ws_interface.ApiRequest) (interface{}, *ws_interface.HandlerError) {
	var payload NewGamePayload
	if err := json.Unmarshal(req.Data.Payload, &payload); err != nil {
		return nil, ws_interface.NewHandlerError(ws_interface.RequestParseError, err)
	}

	game, err := req.Repos.Casino.GetGame(context, payload.GameId)
	if err != nil {
		if err == casino.GameNotFound {
			return nil, ws_interface.NewHandlerError(ws_interface.ContentNotFoundError, err)
		}
		return nil, ws_interface.NewHandlerError(ws_interface.InternalError, err)
	}

	cas, err := req.Repos.Casino.GetCasino(context, payload.CasinoID)
	if err != nil {
		if err == casino.CasinoNotFound {
			return nil, ws_interface.NewHandlerError(ws_interface.ContentNotFoundError, err)
		}
		return nil, ws_interface.NewHandlerError(ws_interface.InternalError, err)
	}

	session, err := req.UseCases.GameSession.NewSession(
		context, cas, game,
		req.User, payload.Deposit,
		payload.ActionType, payload.ActionParams,
	)
	if err != nil {
		return nil, ws_interface.NewHandlerError(ws_interface.InternalError, err)
	}

	err = req.Repos.GameSession.AddGameSessionUpdate(context, &models.GameSessionUpdate{
		SessionID:  session.ID,
		UpdateType: models.SessionCreatedUpdate,
		Timestamp:  time.Now(),
		Data:       nil,
	})
	if err != nil {
		return nil, ws_interface.NewHandlerError(ws_interface.InternalError, err)
	}

	return session, nil
}
