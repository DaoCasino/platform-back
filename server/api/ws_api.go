package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"platform-backend/usecases"
)

type WsApi struct {
	useCases *usecases.UseCases
}

func NewWsApi(useCases *usecases.UseCases) *WsApi {
	wsApi := new(WsApi)
	wsApi.useCases = useCases
	return wsApi
}

type WsRequest struct {
	Request string      `json:"request"`
	Id      string      `json:"id"`
	Payload interface{} `json:"payload"`
}

type WsResponse struct {
	Type    string      `json:"type"`
	Id      string      `json:"id"`
	Status  string      `json:"status"`
	Payload interface{} `json:"payload"`
}

type WsError struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

type RequestHandlerInfo struct {
	handler     func(context context.Context, useCases *usecases.UseCases, session *Session, message []byte) (*WsResponse, error)
	messageType int
}

var handlers = map[string]RequestHandlerInfo{
	"subscribe": {
		handler:     ProcessSubscribeRequest,
		messageType: websocket.TextMessage,
	},
	"new_game": {
		handler:     ProcessNewGameRequest,
		messageType: websocket.TextMessage,
	},
	"game_action": {
		handler:     ProcessGameActionRequest,
		messageType: websocket.TextMessage,
	},
	"fetch_sessions": {
		handler:     ProcessFetchSessionsRequest,
		messageType: websocket.TextMessage,
	},
	"fetch_casinos": {
		handler:     ProcessFetchCasinosRequest,
		messageType: websocket.TextMessage,
	},
	"fetch_games": {
		handler:     ProcessFetchGamesRequest,
		messageType: websocket.TextMessage,
	},
	"fetch_games_in_casino": {
		handler:     ProcessFetchGamesInCasinoRequest,
		messageType: websocket.TextMessage,
	},
}

func ProcessRequest(context context.Context, wsApi *WsApi, session *Session, messageType int, message []byte) (*WsResponse, error) {
	var messageObj WsRequest
	if err := json.Unmarshal(message, &messageObj); err != nil {
		return nil, err
	}

	if messageObj.Id == "" || messageObj.Request == "" {
		return nil, fmt.Errorf("invalid request JSON format")
	}

	if handler, found := handlers[messageObj.Request]; found {
		if handler.messageType != messageType {
			return nil, fmt.Errorf("message type is wrong")
		}
		return handler.handler(context, wsApi.useCases, session, message)
	}

	return nil, fmt.Errorf("unknown request type: %s", messageObj.Request)
}
