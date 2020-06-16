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
)

type FetchSessionPayload struct {
	SessionId eos.Uint64 `json:"sessionId"`
}

type GameSessionResponse struct {
	ID              string                  `json:"id"`
	Player          string                  `json:"player"`
	CasinoID        string                  `json:"casinoId"`
	GameID          string                  `json:"gameId"`
	BlockchainSesID string                  `json:"blockchainSesId"`
	State           models.GameSessionState `json:"state"`
	LastUpdate      int64                   `json:"lastUpdate"`
	Deposit         *eos.Asset              `json:"deposit"`
	PlayerWinAmount *eos.Asset              `json:"playerWinAmount"`
}

func toGameSessionResponse(s *models.GameSession) *GameSessionResponse {
	return &GameSessionResponse{
		ID:              strconv.FormatUint(s.ID, 10),
		Player:          s.Player,
		CasinoID:        strconv.FormatUint(s.CasinoID, 10),
		GameID:          strconv.FormatUint(s.GameID, 10),
		BlockchainSesID: strconv.FormatUint(s.BlockchainSesID, 10),
		State:           s.State,
		LastUpdate:      s.LastUpdate,
		Deposit:         s.Deposit,
		PlayerWinAmount: s.PlayerWinAmount,
	}
}

func ProcessFetchSessionRequest(context context.Context, req *ws_interface.ApiRequest) (interface{}, *ws_interface.HandlerError) {
	var payload FetchSessionPayload
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
		return nil, ws_interface.NewHandlerError(ws_interface.UnauthorizedError, errors.New("attempt to fetch not own session"))
	}

	return toGameSessionResponse(gameSession), nil
}
