package handlers

import (
	"context"
	"encoding/json"
	"github.com/rs/zerolog/log"
	"platform-backend/server/api/interfaces"
)

type GameActionPayload struct {
	SessionId  uint64   `json:"sessionId"`
	ActionType uint16   `json:"actionType"`
	Params     []uint32 `json:"params"`
}

func respondWithError(reqId string, code int32, message string) *interfaces.WsResponse {
	return &interfaces.WsResponse{
		Type:   "response",
		Id:     reqId,
		Status: "error",
		Payload: interfaces.WsError{
			Code:    code,
			Message: message,
		},
	}
}

func ProcessGameActionRequest(context context.Context, req *interfaces.ApiRequest) (*interfaces.WsResponse, error) {
	var payload GameActionPayload
	if err := json.Unmarshal(req.Data.Payload, &payload); err != nil {
		return nil, err
	}

	has, err := req.Repos.GameSession.HasGameSession(context, payload.SessionId)
	if err != nil {
		log.Err(err)
		return respondWithError(req.Data.Id, 5000, "Game session check error"), nil
	}

	if !has {
		log.Debug().Msgf("Session not found, id: %s", payload.SessionId)
		return respondWithError(req.Data.Id, 4004, "Game session not found"), nil
	}

	session, err := req.Repos.GameSession.GetGameSession(context, payload.SessionId)
	if err != nil {
		log.Err(err)
		return respondWithError(req.Data.Id, 5000, "Game session fetch error"), nil
	}

	if req.User.AccountName != session.Player {
		log.Debug().Msgf("Session player mismatch, req: %s, player: %s", req.User.AccountName, session.Player)
		return respondWithError(req.Data.Id, 4003, "Requested session owned by other account"), nil
	}

	err = req.UseCases.Casino.GameAction(context, payload.SessionId, payload.ActionType, payload.Params)
	if err != nil {
		return respondWithError(req.Data.Id, 5000, "Game action trx error"), nil
	}

	return &interfaces.WsResponse{
		Type:    "response",
		Id:      req.Data.Id,
		Status:  "ok",
		Payload: struct{}{},
	}, nil
}
