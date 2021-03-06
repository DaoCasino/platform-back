package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
	"platform-backend/models"
	"platform-backend/repositories"
	"platform-backend/server/api/handlers"
	"platform-backend/server/api/ws_interface"
	"platform-backend/usecases"
	"strconv"
	"time"
)

type WsApi struct {
	UseCases        *usecases.UseCases
	Repos           *repositories.Repos
	eventHistograms map[string]*prometheus.HistogramVec
}

func NewWsApi(useCases *usecases.UseCases, repos *repositories.Repos, registerer prometheus.Registerer) *WsApi {
	wsApi := new(WsApi)
	wsApi.UseCases = useCases
	wsApi.Repos = repos
	wsApi.eventHistograms = make(map[string]*prometheus.HistogramVec)

	wsApiBuckets := prometheus.LinearBuckets(0, 5, 200)

	for reqName := range handlersMap {
		hist := prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "ws_request_" + reqName + "_ms",
			Buckets: wsApiBuckets,
		}, []string{"result"})
		registerer.MustRegister(hist)
		wsApi.eventHistograms[reqName] = hist
	}

	return wsApi
}

type RequestHandlerInfo struct {
	handler     func(context context.Context, req *ws_interface.ApiRequest) (interface{}, *ws_interface.HandlerError)
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
		needAuth:    true,
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
	"fetch_global_sessions": {
		handler:     handlers.ProcessFetchGlobalSessionsRequest,
		messageType: websocket.TextMessage,
		needAuth:    false,
	},
	"fetch_casino_sessions": {
		handler:     handlers.ProcessCasinoSessionsRequest,
		messageType: websocket.TextMessage,
		needAuth:    false,
	},
	"fetch_session_updates": {
		handler:     handlers.ProcessFetchSessionUpdatesRequest,
		messageType: websocket.TextMessage,
		needAuth:    true,
	},
	"fetch_casinos": {
		handler:     handlers.ProcessFetchCasinosRequest,
		messageType: websocket.TextMessage,
		needAuth:    false,
	},
	"fetch_games": {
		handler:     handlers.ProcessFetchGamesRequest,
		messageType: websocket.TextMessage,
		needAuth:    false,
	},
	"fetch_games_in_casino": {
		handler:     handlers.ProcessFetchGamesInCasinoRequest,
		messageType: websocket.TextMessage,
		needAuth:    false,
	},
}

func respondWithError(reqId string, code ws_interface.WsErrorCode) *ws_interface.WsResponse {
	return &ws_interface.WsResponse{
		Type:   "response",
		Id:     reqId,
		Status: "error",
		Payload: &ws_interface.WsError{
			Code:    code,
			Message: ws_interface.GetErrorMsg(code),
		},
	}
}

func respondWithOK(reqId string, payload interface{}) *ws_interface.WsResponse {
	return &ws_interface.WsResponse{
		Type:    "response",
		Id:      reqId,
		Status:  "ok",
		Payload: payload,
	}
}

func (api *WsApi) ProcessRawRequest(context context.Context, messageType int, message []byte) (*ws_interface.WsResponse, string, error) {
	var messageObj ws_interface.WsRequest
	if err := json.Unmarshal(message, &messageObj); err != nil {
		return nil, "", err
	}

	if messageObj.Id == "" || messageObj.Request == "" {
		return nil, "", fmt.Errorf("invalid request JSON format")
	}

	suid := context.Value("suid").(uuid.UUID).String()
	log.Info().Msgf("WS started '%s' request from suid: %s, req: %s", messageObj.Request, suid, messageObj.Payload)

	// get user info from context
	user := context.Value("user").(*models.User)

	if handler, found := handlersMap[messageObj.Request]; found {
		if handler.messageType != messageType {
			log.Info().Msgf("WS request from: %s has wrong message type: %d", suid, messageType)
			return nil, "", fmt.Errorf("message type is wrong")
		}

		if handler.needAuth && user == nil {
			log.Info().Msgf("WS request from: %s unauthorized", suid)
			return respondWithError(messageObj.Id, ws_interface.UnauthorizedError), messageObj.Request, nil
		}

		start := time.Now()
		// process request
		wsResp, handlerError := handler.handler(context, &ws_interface.ApiRequest{
			UseCases: api.UseCases,
			Repos:    api.Repos,
			User:     user,
			Data:     &messageObj,
		})
		t := time.Now()
		elapsed := t.Sub(start)

		if handlerError != nil {
			api.eventHistograms[messageObj.Request].WithLabelValues(strconv.FormatInt(int64(handlerError.Code), 10)).Observe(float64(elapsed.Milliseconds()))
			if handlerError.Code == ws_interface.InternalError {
				log.Error().Msgf("WS request internal error from suid: %s, err: %s", suid, handlerError.InternalError.Error())
			} else {
				log.Info().Msgf("WS request failed from suid: %s, code: %d, err: %s", suid, handlerError.Code, handlerError.InternalError.Error())
			}
			return respondWithError(messageObj.Id, handlerError.Code), messageObj.Request, nil
		}

		api.eventHistograms[messageObj.Request].WithLabelValues("success").Observe(float64(elapsed.Milliseconds()))

		log.Info().Msgf("WS successfully finished request from suid: %s", suid)
		return respondWithOK(messageObj.Id, wsResp), messageObj.Request, nil
	}

	log.Info().Msgf("WS request from '%s' has wrong request type: %s", suid, messageObj.Request)
	return nil, messageObj.Request, fmt.Errorf("unknown request type: %s", messageObj.Request)
}
