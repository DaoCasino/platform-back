package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"platform-backend/models"
	"platform-backend/repositories"
	"platform-backend/server/api/handlers"
	"platform-backend/server/api/interfaces"
	"platform-backend/usecases"
)

type WsApi struct {
	useCases *usecases.UseCases
	repos    *repositories.Repos
}

func NewWsApi(useCases *usecases.UseCases, repos *repositories.Repos) *WsApi {
	wsApi := new(WsApi)
	wsApi.useCases = useCases
	wsApi.repos = repos
	return wsApi
}

type RequestHandlerInfo struct {
	handler     func(context context.Context, req *interfaces.ApiRequest) (*interfaces.WsResponse, error)
	messageType int
	needAuth    bool
}

var handlersMap = map[string]RequestHandlerInfo{
	"auth": {
		handler:     handlers.ProcessAuthRequest,
		messageType: websocket.TextMessage,
		needAuth:    false,
	},
	"account_info": {
		handler:     handlers.ProcessAccountInfo,
		messageType: websocket.TextMessage,
		needAuth:    false,
	},
	"subscribe": {
		handler:     handlers.ProcessSubscribeRequest,
		messageType: websocket.TextMessage,
		needAuth:    true,
	},
	"new_game": {
		handler:     handlers.ProcessNewGameRequest,
		messageType: websocket.TextMessage,
		needAuth:    true,
	},
	"game_action": {
		handler:     handlers.ProcessGameActionRequest,
		messageType: websocket.TextMessage,
		needAuth:    true,
	},
	"fetch_sessions": {
		handler:     handlers.ProcessFetchSessionsRequest,
		messageType: websocket.TextMessage,
		needAuth:    true,
	},
	"fetch_casinos": {
		handler:     handlers.ProcessFetchCasinosRequest,
		messageType: websocket.TextMessage,
		needAuth:    true,
	},
	"fetch_games": {
		handler:     handlers.ProcessFetchGamesRequest,
		messageType: websocket.TextMessage,
		needAuth:    true,
	},
	"fetch_games_in_casino": {
		handler:     handlers.ProcessFetchGamesInCasinoRequest,
		messageType: websocket.TextMessage,
		needAuth:    true,
	},
}

func (api *WsApi) ProcessRawRequest(context context.Context, messageType int, message []byte) (*interfaces.WsResponse, error) {
	var messageObj interfaces.WsRequest
	if err := json.Unmarshal(message, &messageObj); err != nil {
		return nil, err
	}

	if messageObj.Id == "" || messageObj.Request == "" {
		return nil, fmt.Errorf("invalid request JSON format")
	}

	// get user info from context
	user := context.Value("user").(*models.User)

	if handler, found := handlersMap[messageObj.Request]; found {
		if handler.messageType != messageType {
			return nil, fmt.Errorf("message type is wrong")
		}
		if handler.needAuth && user == nil {
			return nil, fmt.Errorf("user unauthorized")
		}

		log.Debug().Msgf("Started '%s' request from suid: %s", messageObj.Request, context.Value("suid").(uuid.UUID).String())

		return handler.handler(context, &interfaces.ApiRequest{
			UseCases: api.useCases,
			Repos:    api.repos,
			User:     user,
			Data:     &messageObj,
		})
	}

	return nil, fmt.Errorf("unknown request type: %s", messageObj.Request)
}
