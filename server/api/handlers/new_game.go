package handlers

import (
	"context"
	"encoding/json"
	"github.com/eoscanada/eos-go"
	"platform-backend/contracts"
	"platform-backend/models"
	"platform-backend/server/api/ws_interface"
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
	if game.Paused != 0 {
		return nil, ws_interface.NewHandlerError(ws_interface.GamePaused, nil)
	}

	cas, err := req.Repos.Contracts.GetCasino(context, uint64(payload.CasinoID))
	if err != nil {
		if err == contracts.CasinoNotFound {
			return nil, ws_interface.NewHandlerError(ws_interface.CasinoNotFoundError, err)
		}
		return nil, ws_interface.NewHandlerError(ws_interface.InternalError, err)
	}
	if cas.Paused {
		return nil, ws_interface.NewHandlerError(ws_interface.CasinoPaused, nil)
	}

	casGames, err := req.Repos.Contracts.GetCasinoGames(context, cas.Contract)
	if err != nil {
		return nil, ws_interface.NewHandlerError(ws_interface.InternalError, err)
	}
	casGame := getGameListedInCasino(game.Id, casGames)
	if casGame == nil {
		return nil, ws_interface.NewHandlerError(ws_interface.GameNotListedInCasino, nil)
	}
	if casGame.Paused {
		return nil, ws_interface.NewHandlerError(ws_interface.GamePaused, nil)
	}

	actionParams := make([]uint64, len(payload.ActionParams))
	for i, param := range payload.ActionParams {
		actionParams[i] = uint64(param)
	}

	session, err := req.UseCases.GameSession.NewSession(
		context, cas, game,
		req.User, payload.Deposit,
		payload.ActionType, actionParams,
	)
	if err != nil {
		return nil, ws_interface.NewHandlerError(ws_interface.InternalError, err)
	}

	return toGameSessionResponse(session), nil
}

func getGameListedInCasino(gameId uint64, casGames []*models.CasinoGame) *models.CasinoGame {
	for _, casGame := range casGames {
		if casGame.Id == gameId {
			return casGame
		}
	}
	return nil
}