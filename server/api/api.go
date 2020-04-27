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
	handler     func(context context.Context, req *interfaces.ApiRequest) (interface{}, *interfaces.HandlerError)
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
	"fetch_session": {
		handler:     handlers.ProcessFetchSessionRequest,
		messageType: websocket.TextMessage,
		needAuth:    true,
	},
	"fetch_sessions": {
		handler:     handlers.ProcessFetchSessionsRequest,
		messageType: websocket.TextMessage,
		needAuth:    true,
	},
	"fetch_session_updates": {
		handler:     handlers.ProcessFetchSessionUpdatesRequest,
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

func respondWithError(reqId string, code interfaces.WsErrorCode) *interfaces.WsResponse {
	return &interfaces.WsResponse{
		Type:   "response",
		Id:     reqId,
		Status: "error",
		Payload: &interfaces.WsError{
			Code:    code,
			Message: interfaces.GetErrorMsg(code),
		},
	}
}

func respondWithOK(reqId string, payload interface{}) *interfaces.WsResponse {
	return &interfaces.WsResponse{
		Type:    "response",
		Id:      reqId,
		Status:  "ok",
		Payload: payload,
	}
}

func (api *WsApi) ProcessRawRequest(context context.Context, messageType int, message []byte) (*interfaces.WsResponse, error) {
	var messageObj interfaces.WsRequest
	if err := json.Unmarshal(message, &messageObj); err != nil {
		return nil, err
	}

	if messageObj.Id == "" || messageObj.Request == "" {
		return nil, fmt.Errorf("invalid request JSON format")
	}

	suid := context.Value("suid").(uuid.UUID).String()
	log.Info().Msgf("WS started '%s' request from suid: %s, req: %s", messageObj.Request, suid, messageObj.Payload)

	// get user info from context
	user := context.Value("user").(*models.User)

	if handler, found := handlersMap[messageObj.Request]; found {
		if handler.messageType != messageType {
			log.Info().Msgf("WS request from: %s has wrong message type: %d", suid, messageType)
			return nil, fmt.Errorf("message type is wrong")
		}

		if handler.needAuth && user == nil {
			log.Info().Msgf("WS request from: %s unauthorized", suid)
			return respondWithError(messageObj.Id, interfaces.UnauthorizedError), nil
		}

		// process request
		wsResp, handlerError := handler.handler(context, &interfaces.ApiRequest{
			UseCases: api.useCases,
			Repos:    api.repos,
			User:     user,
			Data:     &messageObj,
		})

		if handlerError != nil {
			if handlerError.Code == interfaces.InternalError {
				log.Error().Msgf("WS request internal error from suid: %s, err: %s", suid, handlerError.InternalError.Error())
			} else {
				log.Info().Msgf("WS request failed from suid: %s, code: %d, err: %s", suid, handlerError.Code, handlerError.InternalError.Error())
			}
			return respondWithError(messageObj.Id, handlerError.Code), nil
		}

		log.Info().Msgf("WS successfully finished request from suid: %s", suid)
		return respondWithOK(messageObj.Id, wsResp), nil
	}

	log.Info().Msgf("WS request from '%s' has wrong request type: %s", suid, messageObj.Request)
	return nil, fmt.Errorf("unknown request type: %s", messageObj.Request)
}
