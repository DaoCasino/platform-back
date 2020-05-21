package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/eoscanada/eos-go"
	gamesessions "platform-backend/game_sessions"
	"platform-backend/models"
	"platform-backend/server/api/ws_interface"
	"strconv"
	"time"
)

type FetchSessionUpdatesPayload struct {
	SessionId eos.Uint64 `json:"sessionId"`
}

type SessionUpdateResponse struct {
	SessionID  string                       `json:"sessionId"`
	UpdateType models.GameSessionUpdateType `json:"updateType"`
	Timestamp  time.Time                    `json:"timestamp"`
	Data       json.RawMessage              `json:"data"`
}

func toSessionUpdateResponse(su *models.GameSessionUpdate) *SessionUpdateResponse {
	return &SessionUpdateResponse{
		SessionID:  strconv.FormatUint(su.SessionID, 10),
		UpdateType: su.UpdateType,
		Timestamp:  su.Timestamp,
		Data:       su.Data,
	}
}

func ProcessFetchSessionUpdatesRequest(context context.Context, req *ws_interface.ApiRequest) (interface{}, *ws_interface.HandlerError) {
	var payload FetchSessionUpdatesPayload
	if err := json.Unmarshal(req.Data.Payload, &payload); err != nil {
		return nil, ws_interface.NewHandlerError(ws_interface.RequestParseError, err)
	}

	gameSession, err := req.Repos.GameSession.GetGameSession(context, uint64(payload.SessionId))
	if err == gamesessions.ErrGameSessionNotFound {
		return nil, ws_interface.NewHandlerError(ws_interface.SessionNotFoundError, err)
	}
	if err != nil {
		return nil, ws_interface.NewHandlerError(ws_interface.InternalError, err)
	}

	if gameSession.Player != req.User.AccountName {
		return nil, ws_interface.NewHandlerError(ws_interface.UnauthorizedError, errors.New("attempt to fetch updates for not own session"))
	}

	gameSessionUpdates, err := req.Repos.GameSession.GetGameSessionUpdates(context, gameSession.ID)
	if err != nil {
		return nil, ws_interface.NewHandlerError(ws_interface.InternalError, err)
	}

	response := make([]*SessionUpdateResponse, len(gameSessionUpdates))
	for i, su := range gameSessionUpdates {
		response[i] = toSessionUpdateResponse(su)
	}

	return response, nil
}
