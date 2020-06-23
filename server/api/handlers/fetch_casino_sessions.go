package handlers

import (
	"context"
	"encoding/json"
	"github.com/eoscanada/eos-go"
	gamesessions "platform-backend/game_sessions"
	"platform-backend/server/api/ws_interface"
)

type FetchCasinoSessionsPayload struct {
	Filter   gamesessions.FilterType `json:"filter"`
	CasinoId eos.Uint64              `json:"casinoId"`
}

func ProcessCasinoSessionsRequest(context context.Context, req *ws_interface.ApiRequest) (interface{}, *ws_interface.HandlerError) {
	var payload FetchCasinoSessionsPayload
	if err := json.Unmarshal(req.Data.Payload, &payload); err != nil {
		return nil, ws_interface.NewHandlerError(ws_interface.RequestParseError, err)
	}

	gameSessions, err := req.Repos.GameSession.GetCasinoSessions(context, payload.Filter, payload.CasinoId)

	if err != nil {
		return nil, ws_interface.NewHandlerError(ws_interface.InternalError, err)
	}

	response := make([]*GameSessionResponse, len(gameSessions))
	for i, session := range gameSessions {
		response[i] = toGameSessionResponse(session)
	}

	return response, nil
}
