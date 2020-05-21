package handlers

import (
	"context"
	"encoding/json"
	"github.com/eoscanada/eos-go"
	"github.com/rs/zerolog/log"
	"platform-backend/contracts"
	"platform-backend/models"
	"platform-backend/server/api/ws_interface"
	"time"
)

type NewGamePayload struct {
	GameId       eos.Uint64   `json:"gameId"`
	CasinoID     eos.Uint64   `json:"casinoId"`
	Deposit      string       `json:"deposit"`
	ActionType   uint16       `json:"actionType"`
	ActionParams []eos.Uint64 `json:"actionParams"`
}

func ProcessNewGameRequest(context context.Context, req *ws_interface.ApiRequest) (interface{}, *ws_interface.HandlerError) {
	var payload NewGamePayload
	if err := json.Unmarshal(req.Data.Payload, &payload); err != nil {
		return nil, ws_interface.NewHandlerError(ws_interface.RequestParseError, err)
	}

	game, err := req.Repos.Contracts.GetGame(context, uint64(payload.GameId))
	if err != nil {
		if err == contracts.GameNotFound {
			return nil, ws_interface.NewHandlerError(ws_interface.GameNotFoundError, err)
		}
		return nil, ws_interface.NewHandlerError(ws_interface.InternalError, err)
	}

	cas, err := req.Repos.Contracts.GetCasino(context, uint64(payload.CasinoID))
	if err != nil {
		if err == contracts.CasinoNotFound {
			return nil, ws_interface.NewHandlerError(ws_interface.CasinoNotFoundError, err)
		}
		return nil, ws_interface.NewHandlerError(ws_interface.InternalError, err)
	}

	session, err := req.UseCases.GameSession.NewSession(
		context, cas, game,
		req.User, payload.Deposit,
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

	var actionParams []uint64
	for _, param := range payload.ActionParams {
		actionParams = append(actionParams, uint64(param))
	}

	// try to instantly make first game action
	err = req.UseCases.GameSession.GameAction(context, session.ID, payload.ActionType, actionParams)
	if err != nil { // if error just save action to db
		log.Debug().Msgf("Instant first action failed, saving action params for session: %d", session.ID)
		err = req.Repos.GameSession.AddFirstGameAction(context, session.ID, &models.GameAction{
			Type:   payload.ActionType,
			Params: actionParams,
		})
		if err != nil {
			return nil, ws_interface.NewHandlerError(ws_interface.InternalError, err)
		}
	}

	return toGameSessionResponse(session), nil
}
